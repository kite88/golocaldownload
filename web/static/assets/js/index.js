getList()

// esc 把任意文本转义后再拼进 HTML。下载库里的文件名完全由使用者控制，
// 直接拼字符串会让 `<img src=x onerror=alert(1)>.txt` 这样的文件名变成 XSS。
function esc(text) {
    return String(text)
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;')
}

// attr 统一用 getAttribute 取值：jQuery 的 .data() 会把 "123"、"true" 这类值自动转成
// 数字 / 布尔，碰上纯数字的目录名就会把路径类型改掉。
function attr(obj, name) {
    if (obj && obj.getAttribute) return obj.getAttribute(name)
    return $(obj).attr(name)
}

function getList() {
    const path = getQueryParam('s') ?? ''
    const file = getQueryParam('f') ?? ''
    $.get('/api/list', {path: path}, function (res) {
        renderBreadcrumb(res)
        renderList(res, file)
    }, 'json').fail(function (xhr) {
        renderError((xhr.responseJSON && xhr.responseJSON.error) || '目录读取失败')
    })
}

function renderBreadcrumb(res) {
    const $nav = $('.nav')
    $nav.children().remove()
    for (const item of res.relative_dirs) {
        for (const key in item) {
            const name = key === '' ? '根目录' : esc(key)
            $nav.append('<li><a data-path="' + esc(item[key]) + '" onclick="getNextList(this)">' + name + '</a></li>')
        }
    }

    // Bootstrap 5 不会自动初始化 tooltip，而且标题只在初始化时读取一次，
    // 所以先销毁旧实例再按最新路径重建。
    const title = '本地服务器存放路径：' + res.root_dir
    const tip = bootstrap.Tooltip.getInstance($nav[0])
    if (tip) tip.dispose()
    new bootstrap.Tooltip($nav[0], {title: title, placement: 'bottom'})
}

function renderList(res, file) {
    const $tbody = $('tbody')
    const list = res.list

    if (list.length === 0) {
        $tbody.html('<tr><td style="text-align: center" colspan="3">文件夹空空如也！</td></tr>')
        return
    }

    // 从检索结果跳转过来时高亮命中的文件；它如果排在很后面，再挪到最前面。
    if (file !== '') {
        const index = list.findIndex(item => item.name === file)
        if (index >= 0) {
            if (index > 10) {
                const target = list.splice(index, 1)[0]
                target.highlight = true
                list.unshift(target)
            } else {
                list[index].highlight = true
            }
        }
    }

    const rows = list.map(item => {
        const style = item.highlight ? ' style="background: #d1c2c2;color: #000000;"' : ''
        const name = esc(item.name)
        if (item.is_dir) {
            return '<tr>' +
                '<td' + style + '><a onclick="getNextList(this)" data-path="' + esc(item.path) + '">' +
                '<img src="/web/static/icon/folder.png" alt="">' + name + '</a></td>' +
                '<td' + style + '></td>' +
                '<td' + style + '>' + esc(item.mod_time) + '</td>' +
                '</tr>'
        }
        return '<tr>' +
            '<td' + style + '><img src="/web/static/icon/file.png" alt="">' + name +
            '<button onclick="download(this)" data-pathname-key="' + esc(item.pathname_key) + '" class="btn btn-sm btn-link">下载</button></td>' +
            '<td' + style + '>' + esc(item.size) + ' ' + esc(item.size_unit) + '</td>' +
            '<td' + style + '>' + esc(item.mod_time) + '</td>' +
            '</tr>'
    })
    $tbody.html(rows.join(''))
}

function renderError(message) {
    $('.nav').children().remove()
    $('tbody').html('<tr><td style="text-align: center; color: #b02a37" colspan="3">' + esc(message) + '</td></tr>')
}

function getNextList(obj) {
    changeUrlParam(attr(obj, 'data-path'))
    getList()
}

function changeUrlParam(path = '') {
    const url = new URL(window.location.href)
    url.search = ''
    url.searchParams.set('s', path)
    window.history.replaceState(null, null, url.toString())
}

function getQueryParam(param) {
    const match = new RegExp('[?&]' + param + '=([^&]*)').exec(window.location.search)
    return match && decodeURIComponent(match[1].replace(/\+/g, ' '))
}

function download(obj) {
    location.href = '/api/download?data=' + encodeURIComponent(attr(obj, 'data-pathname-key'))
}

$('#search-open-btn').on('click', function () {
    searchOpen()
})

$('#search-get-btn').on('click', function () {
    searchGet()
})

// 回车即检索，省得每次都得去点按钮
$('#search-input-text').on('keydown', function (e) {
    if (e.key === 'Enter') searchOpen()
})

$('#search-input-text-i').on('keydown', function (e) {
    if (e.key === 'Enter') searchGet()
})

function searchOpen() {
    const content = $.trim($('#search-input-text').val())
    if (content === '') {
        return alert('请输入检索信息')
    }
    new bootstrap.Modal(document.getElementById('searchModal')).show()
    $('#search-input-text-i').val(content)
    getSearchList(content)
}

function searchGet() {
    const content = $.trim($('#search-input-text-i').val())
    if (content === '') return alert('请输入检索信息')
    getSearchList(content)
}

function getSearchList(keyword) {
    const $mBody = $('#search-model-body')
    const $resLen = $('#res-length')
    $mBody.html('<div class="spinner-border" role="status"><span class="visually-hidden">Loading...</span></div><span>正在检索，请稍等...</span>')
    $resLen.html('')

    $.post('/api/search', {keyword: keyword}, function (res) {
        if (!Array.isArray(res) || res.length === 0) {
            $mBody.html('<h6>没有检索到相应的文件或目录</h6>')
            return
        }
        $resLen.html('检索到<b style="color: red">' + res.length + '</b>条记录')

        // 关键字先转义再高亮：转义后两边仍是同一段文本，普通字符串替换即可命中。
        const kw = esc(keyword)
        const rows = res.map((item, index) => {
            const name = esc(item.name)
            let img, size = '', btn
            if (item.is_dir) {
                img = '<img src="/web/static/icon/folder.png" alt=""> '
                btn = '<a class="btn-link" href="?s=' + encodeURIComponent(item.path) + '">定位到此目录</a>'
            } else {
                img = '<img src="/web/static/icon/file.png" alt=""> '
                size = '<i class="text-info">' + esc(item.size) + ' ' + esc(item.size_unit) + '</i>'
                btn = '<button onclick="download(this)" data-pathname-key="' + esc(item.pathname_key) + '" class="btn btn-sm btn-link">下载</button>'
                btn += '<a class="btn-link" href="?s=' + encodeURIComponent(item.parent_path) + '&f=' + encodeURIComponent(item.name) + '">定位到文件目录</a>'
            }
            const path = esc(String(item.path).replace(/[\\/]/g, '/'))
            const highlighted = path.split(kw).join('<b style="color: red">' + kw + '</b>')
            return '<li class="list-group-item">' + (index + 1) + '&nbsp;&nbsp;' + img + ' 根目录' + highlighted + ' ' + size + ' ' + btn + '</li>'
        })
        $mBody.html('<ul class="list-group">' + rows.join('') + '</ul>')
    }, 'json').fail(function (xhr) {
        $mBody.html('<h6>检索失败：' + esc((xhr.responseJSON && xhr.responseJSON.error) || '服务异常') + '</h6>')
    })
}
