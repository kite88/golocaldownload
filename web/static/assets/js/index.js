// 说明：页面里只用 Bootstrap 的 Modal（不依赖 Popper）。
// 内嵌的 bootstrap.min.js 是「非 bundle 版」，不含 Popper，所以按钮组之外不要引入
// 下拉框 / tooltip / popover 这类组件，否则会抛 `_createPopper is not a function`。

const THEME_KEY = 'gld-theme'

initTheme()
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

// ---------------------------------------------------------------- 主题（亮色 / 跟随系统 / 暗色）

// 读取保存的主题偏好；localStorage 在隐私模式下可能不可用，读不到就当「跟随系统」。
function getTheme() {
    try {
        const saved = localStorage.getItem(THEME_KEY)
        if (saved === 'light' || saved === 'dark' || saved === 'system') return saved
    } catch (e) {
        // 忽略，按跟随系统处理
    }
    return 'system'
}

function isDarkTheme(mode) {
    if (mode === 'dark') return true
    if (mode === 'light') return false
    return window.matchMedia('(prefers-color-scheme: dark)').matches
}

// 把主题写到 <html data-bs-theme>，Bootstrap 5.3 的所有组件颜色都跟着它走。
function applyTheme(mode) {
    document.documentElement.setAttribute('data-bs-theme', isDarkTheme(mode) ? 'dark' : 'light')
    $('.theme-switch button[data-theme]').each(function () {
        const active = attr(this, 'data-theme') === mode
        $(this).toggleClass('active', active).attr('aria-pressed', active)
    })
}

function setTheme(mode) {
    try {
        localStorage.setItem(THEME_KEY, mode)
    } catch (e) {
        // 存不上不影响本次生效
    }
    applyTheme(mode)
}

function initTheme() {
    // 首帧的主题已由 index.html 里的内联脚本设置好，这里只同步按钮状态与监听。
    applyTheme(getTheme())
    $('.theme-switch button[data-theme]').on('click', function () {
        setTheme(attr(this, 'data-theme'))
    })
    // 「跟随系统」时，系统切换深色要实时生效
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function () {
        if (getTheme() === 'system') applyTheme('system')
    })
}

// ---------------------------------------------------------------- 文件类型图标

// 图标取自 Material Icon Theme（VS Code 默认图标集，MIT，见 web/static/icon/material/LICENSE.txt）。
// 该图标集里压缩包只有一个 zip 图标、表格类统一叫 table、通用文档叫 document，
// 所以下面的映射里 rar/7z 等都指向 zip，xls/csv 指向 table。
const ICON_DIR = '/web/static/icon/material/'
const ICON_FOLDER = ICON_DIR + 'folder-base.svg'
const ICON_BY_EXT = {
    // 压缩包
    zip: 'zip', rar: 'zip', '7z': 'zip', tar: 'zip', gz: 'zip', tgz: 'zip', bz2: 'zip', xz: 'zip', zst: 'zip',
    // 图片
    jpg: 'image', jpeg: 'image', png: 'image', gif: 'image', webp: 'image', bmp: 'image', tif: 'image', tiff: 'image', heic: 'image', raw: 'image',
    svg: 'svg',
    // 视频
    mp4: 'video', mkv: 'video', avi: 'video', mov: 'video', wmv: 'video', flv: 'video', webm: 'video', m4v: 'video', mpg: 'video', mpeg: 'video', rmvb: 'video', '3gp': 'video',
    // 音频
    mp3: 'audio', wav: 'audio', flac: 'audio', aac: 'audio', ogg: 'audio', m4a: 'audio', wma: 'audio', opus: 'audio',
    // 文档
    pdf: 'pdf',
    doc: 'word', docx: 'word', rtf: 'word', odt: 'word',
    xls: 'table', xlsx: 'table', csv: 'table', tsv: 'table', ods: 'table',
    ppt: 'powerpoint', pptx: 'powerpoint', odp: 'powerpoint',
    md: 'markdown', markdown: 'markdown',
    txt: 'document', log: 'document', ini: 'document', conf: 'document', cfg: 'document',
    // 数据 / 配置
    json: 'json', json5: 'json',
    xml: 'xml', plist: 'xml',
    yml: 'yaml', yaml: 'yaml', toml: 'yaml',
    db: 'database', sqlite: 'database', sqlite3: 'database', sql: 'database', mdb: 'database',
    // 代码
    html: 'html', htm: 'html',
    css: 'css',
    scss: 'sass', sass: 'sass',
    less: 'less',
    js: 'javascript', mjs: 'javascript', cjs: 'javascript', jsx: 'javascript',
    ts: 'typescript', tsx: 'typescript',
    go: 'go',
    py: 'python', pyw: 'python',
    java: 'java', jar: 'java', class: 'java',
    rs: 'rust',
    php: 'php',
    sh: 'console', bash: 'console', zsh: 'console', bat: 'console', cmd: 'console',
    ps1: 'powershell', psm1: 'powershell',
    // 其它
    ttf: 'font', otf: 'font', woff: 'font', woff2: 'font',
    exe: 'exe', msi: 'exe', com: 'exe',
    dll: 'dll', so: 'dll', dylib: 'dll',
    iso: 'disc', img: 'disc', dmg: 'disc', vhd: 'disc',
    pem: 'key', key: 'key', crt: 'key', cer: 'key', pfx: 'key', p12: 'key',
    lock: 'lock',
}

// fileIcon 按扩展名挑图标；没有扩展名（包括 .gitignore 这类点开头的文件）或未收录的
// 扩展名，一律回落到通用文档图标。
function fileIcon(name) {
    const dot = String(name).lastIndexOf('.')
    const ext = dot > 0 ? String(name).slice(dot + 1).toLowerCase() : ''
    return ICON_DIR + (ICON_BY_EXT[ext] || 'document') + '.svg'
}

// ---------------------------------------------------------------- 目录列表

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

    // 下载库路径同时写进页面副标题（比只挂在 title 上更容易发现）；这里的 title 用原生
    // 属性实现，不用 bootstrap.Tooltip —— 它依赖 Popper，而非 bundle 版里没有 Popper。
    $nav.attr('title', '本地服务器存放路径：' + res.root_dir)
    // 副标题在窄列里会省略号截断，所以完整路径另挂在 title 上，悬停可看全
    $('#lib-path').text('下载库路径：' + res.root_dir).attr('title', res.root_dir)
}

function renderList(res, file) {
    const $tbody = $('tbody')
    const list = res.list

    if (list.length === 0) {
        $tbody.html('<tr><td class="cell-center" colspan="3">文件夹空空如也！</td></tr>')
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
        const cls = item.highlight ? ' class="hl"' : ''
        const name = esc(item.name)
        if (item.is_dir) {
            return '<tr>' +
                '<td' + cls + '><a onclick="getNextList(this)" data-path="' + esc(item.path) + '">' +
                '<img class="file-icon" src="' + ICON_FOLDER + '" alt="">' + name + '</a></td>' +
                '<td' + cls + '></td>' +
                '<td' + cls + '>' + esc(item.mod_time) + '</td>' +
                '</tr>'
        }
        return '<tr>' +
            '<td' + cls + '><img class="file-icon" src="' + fileIcon(item.name) + '" alt="">' + name +
            '<button onclick="download(this)" data-pathname-key="' + esc(item.pathname_key) + '" class="btn btn-sm btn-outline-primary ms-2">下载</button></td>' +
            '<td' + cls + '>' + esc(item.size) + ' ' + esc(item.size_unit) + '</td>' +
            '<td' + cls + '>' + esc(item.mod_time) + '</td>' +
            '</tr>'
    })
    $tbody.html(rows.join(''))
}

function renderError(message) {
    $('.nav').children().remove()
    $('tbody').html('<tr><td class="cell-center err" colspan="3">' + esc(message) + '</td></tr>')
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

// ---------------------------------------------------------------- 检索

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
        $resLen.html('检索到<b class="kw">' + res.length + '</b>条记录')

        // 关键字先转义再高亮：转义后两边仍是同一段文本，普通字符串替换即可命中。
        const kw = esc(keyword)
        const rows = res.map((item, index) => {
            const name = esc(item.name)
            let img, size = '', btn
            if (item.is_dir) {
                img = '<img class="file-icon" src="' + ICON_FOLDER + '" alt=""> '
                btn = '<a class="btn-link" href="?s=' + encodeURIComponent(item.path) + '">定位到此目录</a>'
            } else {
                img = '<img class="file-icon" src="' + fileIcon(item.name) + '" alt=""> '
                size = '<i class="text-info">' + esc(item.size) + ' ' + esc(item.size_unit) + '</i>'
                btn = '<button onclick="download(this)" data-pathname-key="' + esc(item.pathname_key) + '" class="btn btn-sm btn-outline-primary ms-2">下载</button>'
                btn += '<a class="btn-link" href="?s=' + encodeURIComponent(item.parent_path) + '&f=' + encodeURIComponent(item.name) + '">定位到文件目录</a>'
            }
            const path = esc(String(item.path).replace(/[\\/]/g, '/'))
            const highlighted = path.split(kw).join('<b class="kw">' + kw + '</b>')
            return '<li class="list-group-item">' + (index + 1) + '&nbsp;&nbsp;' + img + ' 根目录' + highlighted + ' ' + size + ' ' + btn + '</li>'
        })
        $mBody.html('<ul class="list-group">' + rows.join('') + '</ul>')
    }, 'json').fail(function (xhr) {
        $mBody.html('<h6>检索失败：' + esc((xhr.responseJSON && xhr.responseJSON.error) || '服务异常') + '</h6>')
    })
}
