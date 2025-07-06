getList()

function getList() {
    let path = getQueryParam('s') ?? '';
    let file = getQueryParam('f') ?? '';
    $.get("/api/list", {"path": path}, function (res) {
        const $nav = $('.nav')
        $nav.attr("title", "本地服务器存放路径：" + res.absolute_dir)
        $nav.children().remove();
        for (const item of res.relative_dirs) {
            for (let k in item) {
                let name = k
                if (k === '') name = '根目录'
                $nav.append('<li><a data-path="' + item[k] + '" onclick="getNextList(this)">' + name + '</a></li>');
            }
        }
        let list = res.list
        let $tbody = $('tbody');
        if (list.length === 0) {
            $tbody.html('<tr><td style="text-align: center" colspan="4">文件夹空空如也！</td></tr>');
            return
        }

        if (file !== '') {
            for (const i in res.list) {
                if (file === list[i].name) {
                    list[i].isFindFile = 1;
                    if (i < 10) break
                    let item = list.splice(i, 1)[0];// 从数组中移除该项并获得它
                    list.unshift(item);
                    break;
                }
            }
        }

        let tbody = '';
        list.map(v => {
            let styleColor = ''
            if (v.isFindFile && v.isFindFile === 1) {
                styleColor = 'background: #d1c2c2;color: #000000;'
            }
            tbody += '<tr>'
            if (v.is_dir) {
                tbody += '<td><a onclick="getNextList(this)" data-path="' + v.path + '"><img src="/web/static/icon/folder.png" alt="">' + v.name + '</a></td>'
                tbody += '<td></td>'
            } else {
                tbody += '<td style="' + styleColor + '"><img src="/web/static/icon/file.png" alt="">' + v.name + '<button onclick="download(this)" data-pathname_key="' + v.pathname_key + '" class="btn btn-sm btn-link">下载</button>' + '</td>'
                tbody += '<td style="' + styleColor + '">' + v.size + ' ' + v.size_unit + '</td>'
            }
            tbody += '<td style="' + styleColor + '">' + v.mod_time + '</td>'
            tbody += '</tr>'
        })
        $tbody.html(tbody);
    }, "json")
}

function getNextList(obj) {
    let path = $(obj).data('path')
    changeUrlParam(path)
    getList()
}

function changeUrlParam(path = '') {
    let url = new URL(window.location.origin);
    url.searchParams.append('s', path);
    window.history.replaceState(null, null, url.toString())
}

function getQueryParam(param) {
    const match = RegExp('[?&]' + param + '=([^&]*)').exec(window.location.search);
    return match && decodeURIComponent(match[1].replace(/\+/g, ' '));
}

function download(obj) {
    const pathname_key = $(obj).data('pathname_key');
    location.href = "/api/download?data=" + pathname_key;
}

function searchOpen() {
    let content = $.trim($('#search-input-text').val())
    if (content === '') {
        return alert("请输入检索信息")
    }
    const searchModal = new bootstrap.Modal(document.getElementById('searchModal'))
    searchModal.show()
    $('#search-input-text-i').val(content)
    getSearchList(content)
}

function searchGet() {
    let content = $.trim($('#search-input-text-i').val())
    if (content === '') return alert("请输入检索信息")
    getSearchList(content)
}

$('#search-open-btn').on('click', function () {
    searchOpen()
})

$('#search-get-btn').on('click', function () {
    searchGet()
})

function getSearchList(keyword) {
    const $mBody = $('#search-model-body');
    const $resLen = $('#res-length');
    const loading = '<div class="spinner-border" role="status"><span class="visually-hidden">Loading...</span></div><span>正在检索，请稍等...</span>'
    $mBody.html(loading);
    $resLen.html('')
    $.post("/api/search", {keyword: keyword}, function (res) {
        if (res.length === 0) return $mBody.html('<h6>没有检索到相应的文件或目录</h6>');
        $resLen.html('检索到<b style="color: red">' + res.length + '</b>条记录')
        let html = '<ul class="list-group">'
        let index = 0;
        res.map(item => {
            index++;
            html += '<li class="list-group-item">'
            let img = '', size = '', btn = ''
            if (item.is_dir) {
                img = '<img src="/web/static/icon/folder.png" alt=""> '
                size = ''
                btn = '<a class="btn-link" href="' + window.location.origin + '?s=' + item.path + '">定位到此目录</a>'
            } else {
                img = '<img src="/web/static/icon/file.png" alt=""> '
                size = '<i class="text-info">' + item.size + ' ' + item.size_unit + '</i>'
                btn = '<button onclick="download(this)" data-pathname_key="' + item.pathname_key + '" class="btn btn-sm btn-link">下载</button>'
                btn += '<a class="btn-link" href="' + window.location.origin + '?s=' + item.parent_path + '&f=' + item.name + '">定位到文件目录</a>'
            }
            let resultPath = item.path.replace(/\\|\//g, function (x) {
                return '/'
            });
            resultPath = resultPath.replaceAll(keyword, '<b style="color: red">' + keyword + '</b>')
            html += index + '   ' + img + ' 根目录' + resultPath + ' ' + size + ' ' + btn
            html += '</li>'
        })
        html += '</ul>'
        $mBody.html(html);
    }, "json")
}