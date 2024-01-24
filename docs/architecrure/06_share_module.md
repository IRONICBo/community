# Share module

## Module purpose

This module mainly shares technical articles with Twitter and Facebook

## Module scope

```该模块主要用到了tw和fb的分享路由```

## Module structure

在这里详细描述模块的内部结构和行为。

## Module Interface

若Meta标签解析出错或者传参错误会导致解析失败
## Functions

### Login

- 功能：函数的功能描述
- 入参：参数的类型和说明
- 返回：返回值的类型和说明
- 错误：可能的错误和异常

示例：

```javascript
  function shareOnTwitter() {
    // var url = encodeURIComponent('http://8.130.26.238/animation/fb.html');
    // var text = encodeURIComponent('Linz');
    var url = 'http://8.130.26.238/animation/fb.html';
    var text = "Test"
    var via = "Test";
    var hashtags = "Test";
    var intentUrl = "https://twitter.com/intent/tweet?text="
            + encodeURIComponent(text) + "&url=" + encodeURIComponent(url)
            + "&via=" + encodeURIComponent(via) + "&hashtags=" + encodeURIComponent(hashtags);
    // var shareUrl = 'https://twitter.com/intent/tweet?text=' + text + '&url=' + url;
    window.open(intentUrl, "_blank", "width=550,height=420");
  }

```

