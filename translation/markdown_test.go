/*
 * Copyright (c) 2023 The GoPlus Authors (goplus.org). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package translation

import (
	"fmt"
	"os"
	"testing"
	"time"

	language "golang.org/x/text/language"
)

var (
	// Your api key
	mockKey = os.Getenv("NIUTRANS_API_KEY")
)

func TestTranslatePlainText(t *testing.T) {
	tests := []struct {
		src  string
		from language.Tag
		to   language.Tag
	}{
		{"你好", language.Chinese, language.English},
		{"hello", language.English, language.Chinese},
	}

	trans := New(mockKey)
	for _, test := range tests {
		to, err := trans.TranslatePlainText(test.src, test.from, test.to)
		fmt.Println(to, err)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Second * 5)
	}
}

func TestTranslateMarkdown(t *testing.T) {
	tests := []struct {
		src  string
		from language.Tag
		to   language.Tag
	}{
		// {`# Hello`, "en", "zh"},
		// {`# 你好`, "zh", "en"},
		{`
# Heading

## 这是一个二级标题

这是一段普通的文本。

这是一段*粗体*的文本。

这是一段**粗体**的文本。

这是一段***粗斜体***的文本。

这是一段` + "`" + `行内代码` + "`" + `。

这是一段代码块：

` + "```" + `go
func main() {
	fmt.Println("Hello, World!")
}
` + "```" + `

这是一个列表：

- 列表项 1
- 列表项 2
	- 列表项 2.1
	- 列表项 2.2
- 列表项 3

这是一个有序列表：

1. 列表项 1
2. 列表项 2
	1. 列表项 2.1
	2. 列表项 2.2
3. 列表项 3

这是一个引用：

> 这是一段引用。

这里一个缩进代码块：

	func main() {
		fmt.Println("Hello, World!")
	}

这是一段[链接](https://www.example.com)

这是一段![图像](https://www.example.com/image.jpg)
`, language.Chinese, language.English},
	}

	trans := New(mockKey)
	for _, test := range tests {
		translatedResult, err := trans.TranslateMarkdownText(test.src, test.from, test.to)
		fmt.Println(translatedResult, err)
		if err != nil {
			t.Fatal(err)
		}

		// Wait for request limit
		time.Sleep(time.Second * 5)
	}
}

func TestTranslateBatch(t *testing.T) {
	tests := []struct {
		src  []string
		from language.Tag
		to   language.Tag
	}{
		{[]string{"你好", "好的", "非常棒"}, language.Chinese, language.English},
		{[]string{"What 's your name", "nice job"}, language.English, language.Chinese},
	}

	trans := New(mockKey)
	for _, test := range tests {
		translatedResult, err := trans.TranslateBatchPlain(test.src, test.from, test.to)
		fmt.Println(translatedResult, err)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Second * 5)
	}
}
