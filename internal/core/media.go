package core

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goplus/yap"
	_ "github.com/qiniu/go-cdk-driver/kodoblob"
	"github.com/qiniu/x/xlog"
)

// todo
type VideoSubtitle struct {
	VideoId    int
	SubtitleId int
	UserId     int
	Language   string
}

type File struct {
	Id       int
	FileKey  string
	Format   string
	UserId   int
	Size     int64
	CreateAt time.Time
	UpdateAt time.Time
}

func (c *Community) DelMedias(ctx context.Context, userId string, ids []string) error {

	uId, err := strconv.Atoi(userId)
	if err != nil {
		return err
	}
	idStr := strings.Join(ids, ",")
	_, err = c.db.ExecContext(ctx, "delete from file where user_id = ? and id in( ? )", uId, idStr)

	return err
}

func (c *Community) DelMedia(ctx context.Context, userId, mediaId string) error {
	// del cloud oss media
	fileKey, err := c.GetMediaUrl(ctx, mediaId)
	if err != nil {
		return err
	}
	if err := c.bucket.Delete(context.Background(), fileKey); err != nil {
		return err
	}
	// del db media
	res, err := c.db.ExecContext(ctx, "delete from file where user_id = ? and id = ?", userId, mediaId)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		c.xLog.Warn("no need del data")
		return nil
	}
	return nil
}

// get file key
func (c *Community) GetMediaUrl(ctx context.Context, mediaId string) (string, error) {
	ID, err := strconv.Atoi(mediaId)
	if err != nil {
		return "", err
	}
	row := c.db.QueryRow(`select file_key from file where id = ?`, ID)
	var fileKey string
	err = row.Scan(&fileKey)
	if err != nil {
		return "", err
	}

	return fileKey, nil
}

func (c *Community) GetMediaType(ctx context.Context, mediaId string) (string, error) {
	ID, err := strconv.Atoi(mediaId)
	if err != nil {
		return "", err
	}
	row := c.db.QueryRow(`select format from file where id = ?`, ID)
	var Format string
	err = row.Scan(&Format)
	if err != nil {
		return "", err
	}
	return Format, nil

}
func (c *Community) GetVideoSubtitle(ctx context.Context, mediaId string) (string, error) {
	ID, err := strconv.Atoi(mediaId)
	if err != nil {
		return "", err
	}
	row := c.db.QueryRow(`select output from video_task where resource_id = ?`, ID)
	var fileKey string
	err = row.Scan(&fileKey)
	if err != nil {
		return "", err
	}

	return fileKey, nil
}

func (c *Community) SaveMedia(ctx context.Context, userId string, data []byte) (int64, error) {

	// upload cloud oss
	fileKey := uuid.New().String()
	err := c.uploadMedia(fileKey, data)
	if err != nil {
		return 0, err
	}
	// get fileInfo
	fileInfo, err := c.getMediaInfo(fileKey)
	if err != nil {
		return 0, err
	}
	// save
	stem, err := c.db.Prepare(`insert into file (file_key,format,size,user_id,create_at,update_at) VALUES (?,?,?,?,?,?)`)
	if err != nil {
		return 0, err
	}
	res, err := stem.Exec(fileKey, fileInfo.Format, fileInfo.Size, userId, time.Now(), time.Now())

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

// for internal use,no need to add ctx
func (c *Community) getMediaInfo(fileKey string) (*File, error) {

	bucket := c.bucket
	r, err := bucket.NewReader(context.Background(), fileKey, nil)
	defer func() {
		err = r.Close()
		if err != nil {
			c.xLog.Error("close file error:", err)
		}
	}()

	if err != nil {
		return nil, err
	}
	fileInfo := &File{
		Format:  r.ContentType(),
		Size:    r.Size(),
		FileKey: fileKey,
	}

	return fileInfo, nil

}

func (c *Community) uploadMedia(fileKey string, data []byte) error {

	w, err := c.bucket.NewWriter(context.Background(), fileKey, nil)
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (c *Community) UploadFile(ctx *yap.Context) {
	xLog := xlog.New("")
	file, header, err := ctx.FormFile("file")
	if err != nil {
		xLog.Error("upload file error:", header)
		ctx.JSON(500, err.Error())
		return
	}
	filename := header.Filename
	err = ctx.ParseMultipartForm(10 << 20)
	if err != nil {
		xLog.Error("upload file error:", filename)
		ctx.JSON(500, err.Error())
		return
	}

	dst, err := os.Create(filename)
	if err != nil {
		xLog.Error("create file error:", file)
		ctx.JSON(500, err.Error())
		return
	}
	defer func() {
		file.Close()
		dst.Close()
		err = os.Remove(filename)
		if err != nil {
			xLog.Error("delete file error:", filename)
			return
		}
	}()

	_, err = io.Copy(dst, file)
	if err != nil {
		xLog.Error("copy file errer:", filename)
		ctx.JSON(500, err.Error())
		return
	}

	bytes, err := os.ReadFile(filename)
	if err != nil {
		xLog.Error("read file errer:", filename)
		ctx.JSON(500, err.Error())
		return
	}

	token, err := GetToken(ctx)
	if err != nil {
		ctx.JSON(500, err.Error())
	}

	uid, err := c.ParseJwtToken(token.Value)
	if err != nil {
		ctx.JSON(500, err.Error())
	}

	id, err := c.SaveMedia(context.Background(), uid, bytes)
	if err != nil {
		xLog.Error("save file", err.Error())
		ctx.JSON(500, err.Error())
		return
	}

	// Judge the file type and start the corresponding task
	fileType := http.DetectContentType(bytes)
	if strings.Contains(fileType, "video") {
		// Start captioning task
		err := c.NewVideoTask(context.Background(), uid, strconv.FormatInt(id, 10))
		if err != nil {
			xLog.Error("start video task error:", err)
		}
	}

	ctx.JSON(200, id)
}

func (c *Community) ListMediaByUserId(ctx context.Context, userId string, format string, from, limitInt int) ([]File, string, error) {
	sqlStr := "select * from file where user_id = ? "
	var args []any
	args = append(args, userId)
	var rows *sql.Rows
	var err error
	if format != "" {
		sqlStr += " and format like ?"
		args = append(args, "%"+format+"%")
	}
	sqlStr = sqlStr + " limit ? offset ?"
	args = append(args, limitInt)
	args = append(args, from)
	rows, err = c.db.Query(sqlStr, args...)

	var files []File
	if err != nil {
		return files, MarkEnd, err
	}
	var rowLen int
	for rows.Next() {
		var file File
		if err := rows.Scan(&file.Id, &file.CreateAt, &file.UpdateAt, &file.FileKey, &file.Format, &file.UserId, &file.Size); err != nil {
			return files, MarkEnd, err
		}
		file.FileKey = c.domain + file.FileKey
		files = append(files, file)
		rowLen++
	}
	if rowLen == 0 {
		return []File{}, MarkEnd, io.EOF
	}
	next := strconv.Itoa(from + rowLen)
	return files, next, nil
}

func (c *Community) ListSubtitleByVideoId(ctx context.Context, videoId int) ([]VideoSubtitle, error) {
	sqlStr := "select * from video_subtitle where video_id =? and "
	rows, err := c.db.Query(sqlStr, videoId)
	var videoSubtitles []VideoSubtitle
	if err != nil {
		return videoSubtitles, err
	}
	defer rows.Close()

	for rows.Next() {
		var videoSubtitle VideoSubtitle
		err = rows.Scan(&videoSubtitle.VideoId, &videoSubtitle.SubtitleId, &videoSubtitle.UserId, &videoSubtitle.Language)
		if err != nil {
			fmt.Println("get data fail:", err.Error())
			return videoSubtitles, err
		}
		videoSubtitles = append(videoSubtitles, videoSubtitle)
	}
	return videoSubtitles, nil
}

func (c *Community) AddSubtitle(ctx context.Context, videoId, subtitleId, userId int, language string) error {
	sqlStr := "INSERT INTO video_subtitle (video_id,user_id,subtitle_id,language) values (?,?,?,?)"
	_, err := c.db.Exec(sqlStr, videoId, userId, subtitleId, language)
	if err != nil {
		c.xLog.Fatalln(err.Error())
		return err
	}
	return nil
}

func (c *Community) DelSubtitle(ctx context.Context, videoId, subtitleId, userId int) error {
	sqlStr := "DELETE FROM video_subtitle where video_id = ? and subtitle_id = ? and user_id = ?"
	_, err := c.db.Exec(sqlStr, videoId, subtitleId, userId)
	if err != nil {
		c.xLog.Fatalln(err.Error())
		return err
	}
	return nil
}
