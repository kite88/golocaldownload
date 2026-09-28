// 说明：页面里只用 Bootstrap 的 Modal（不依赖 Popper）。
// 内嵌的 bootstrap.min.js 是「非 bundle 版」，不含 Popper，所以按钮组之外不要引入
// 下拉框 / tooltip / popover 这类组件，否则会抛 `_createPopper is not a function`。

const THEME_KEY = 'gld-theme'
const VIEW_KEY = 'gld-view'

// 当前视图（list / grid）与最近一次列表响应：切换视图时用同一份数据重绘，不再请求接口。
let currentView = 'list'
let lastList = null

initTheme()
initView()
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

// ---------------------------------------------------------------- 视图切换（列表 / 网格）

// 视图偏好和主题一样存在浏览器本地；localStorage 不可用时按默认的列表处理。
function getView() {
    try {
        const saved = localStorage.getItem(VIEW_KEY)
        if (saved === 'list' || saved === 'grid') return saved
    } catch (e) {
        // 忽略，按默认视图处理
    }
    return 'list'
}

function syncViewButtons() {
    $('.view-switch button[data-view]').each(function () {
        const active = attr(this, 'data-view') === currentView
        $(this).toggleClass('active', active).attr('aria-pressed', active)
    })
}

// applyView 切换表格 / 网格两个容器的显示，并按当前视图重绘一次。
function applyView(view) {
    currentView = view
    $('.list-card').toggleClass('d-none', view !== 'list')
    $('#grid-view').toggleClass('d-none', view !== 'grid')
    syncViewButtons()
    renderList()
}

function initView() {
    // 首帧就按保存的偏好吃住，否则会先闪一下列表再跳到网格
    applyView(getView())
    $('.view-switch button[data-view]').on('click', function () {
        const view = attr(this, 'data-view')
        try {
            localStorage.setItem(VIEW_KEY, view)
        } catch (e) {
            // 存不上不影响本次生效
        }
        applyView(view)
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
        // 存下来给视图切换用：列表与网格共用这一份数据，切换时只重绘不重新请求
        lastList = {res: res, file: file}
        renderBreadcrumb(res)
        renderList()
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
    // 「复制路径」复制的是纯路径（不带「下载库路径：」前缀）。拿到路径后才显示出来，
    // 免得列表还没加载完就点到一个空字符串
    $('#copy-lib-path').attr('data-path', res.root_dir).removeAttr('hidden')
}

// renderList 按当前视图重绘。数据取自最近一次 /api/list 响应，所以切换视图不必再请求一次。
function renderList() {
    if (lastList === null) return
    const list = highlightedList(lastList.res.list, lastList.file)
    if (currentView === 'grid') {
        renderGrid(list)
        return
    }
    renderTable(list)
}

// highlightedList 从检索结果跳转过来时标出命中的文件；它如果排在很后面，再挪到最前面。
// 返回新数组、不修改接口返回的原始数据（切换视图会重绘多次，不能反复改写它）。
function highlightedList(list, file) {
    const rows = list.slice()
    if (file === '') return rows

    const index = rows.findIndex(item => item.name === file)
    if (index < 0) return rows

    if (index > 10) {
        const target = rows.splice(index, 1)[0]
        target.highlight = true
        rows.unshift(target)
    } else {
        rows[index] = Object.assign({}, rows[index], {highlight: true})
    }
    return rows
}

function renderTable(list) {
    const $tbody = $('tbody')
    if (list.length === 0) {
        $tbody.html('<tr><td class="cell-center" colspan="3">文件夹空空如也！</td></tr>')
        return
    }

    const rows = list.map(item => {
        const cls = item.highlight ? ' class="hl"' : ''
        const name = esc(item.name)
        if (item.is_dir) {
            return '<tr>' +
                '<td' + cls + '><a onclick="getNextList(this)" data-path="' + esc(item.path) + '">' +
                '<img class="file-icon" src="' + ICON_FOLDER + '" alt="">' + name + '</a>' +
                '<a onclick="copyFolder(this)" data-path="' + esc(item.path) + '" class="action-link copy-link ms-2">复制地址</a></td>' +
                '<td' + cls + '></td>' +
                '<td' + cls + '>' + esc(item.mod_time) + '</td>' +
                '</tr>'
        }
        return '<tr>' +
            '<td' + cls + '><img class="file-icon" src="' + fileIcon(item.name) + '" alt="">' + name +
            '<a onclick="downloadFile(this)" data-pathname-key="' + esc(item.pathname_key) + '" class="action-link ms-2">下载</a>' +
            '<a onclick="copyDownload(this)" data-pathname-key="' + esc(item.pathname_key) + '" class="action-link copy-link ms-2">复制地址</a></td>' +
            '<td' + cls + '>' + esc(item.size) + ' ' + esc(item.size_unit) + '</td>' +
            '<td' + cls + '>' + esc(item.mod_time) + '</td>' +
            '</tr>'
    })
    $tbody.html(rows.join(''))
}

// 网格视图：卡片式，图标放大、名字最多两行，大小与时间并到一行里。
// 操作与列表一致（目录进目录、文件下载 / 复制地址），只是换了排布。
function renderGrid(list) {
    const $grid = $('#grid-view')
    if (list.length === 0) {
        $grid.html('<div class="grid-empty">文件夹空空如也！</div>')
        return
    }

    const items = list.map(item => {
        const cls = item.highlight ? ' hl' : ''
        const name = esc(item.name)
        const title = ' title="' + name + '"'
        if (item.is_dir) {
            return '<div class="grid-item' + cls + '">' +
                '<a class="grid-main" onclick="getNextList(this)" data-path="' + esc(item.path) + '"' + title + '>' +
                '<img class="grid-icon" src="' + ICON_FOLDER + '" alt="">' +
                '<span class="grid-name">' + name + '</span></a>' +
                '<div class="grid-meta">' + esc(item.mod_time) + '</div>' +
                '<div class="grid-actions">' +
                '<a onclick="copyFolder(this)" data-path="' + esc(item.path) + '" class="action-link copy-link">复制地址</a>' +
                '</div></div>'
        }
        return '<div class="grid-item' + cls + '">' +
            '<div class="grid-main"' + title + '>' +
            '<img class="grid-icon" src="' + fileIcon(item.name) + '" alt="">' +
            '<span class="grid-name">' + name + '</span></div>' +
            '<div class="grid-meta">' + esc(item.size) + ' ' + esc(item.size_unit) + '</div>' +
            '<div class="grid-meta">' + esc(item.mod_time) + '</div>' +
            '<div class="grid-actions">' +
            '<a onclick="downloadFile(this)" data-pathname-key="' + esc(item.pathname_key) + '" class="action-link">下载</a>' +
            '<a onclick="copyDownload(this)" data-pathname-key="' + esc(item.pathname_key) + '" class="action-link copy-link">复制地址</a>' +
            '</div></div>'
    })
    $grid.html(items.join(''))
}

function renderError(message) {
    $('.nav').children().remove()
    // 表格与网格两套 DOM 都在，都要写上，否则切到另一视图时看不到错误
    $('tbody').html('<tr><td class="cell-center err" colspan="3">' + esc(message) + '</td></tr>')
    $('#grid-view').html('<div class="grid-empty err">' + esc(message) + '</div>')
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

// 函数名里不要出现 download：<a> 元素自带 download 属性（HTMLAnchorElement.download），
// 而内联 onclick 的作用域链是「元素 → document → window」，元素上的同名属性会盖住全局函数，
// 于是 onclick="download(this)" 拿到的是那个空字符串属性，直接报 download is not a function。
function downloadFile(obj) {
    location.href = '/api/download?data=' + encodeURIComponent(attr(obj, 'data-pathname-key'))
}

// ---------------------------------------------------------------- 复制地址

// 复制出去的地址要能直接粘到浏览器、wget、下载工具里用，所以一律拼绝对地址。
function downloadUrl(key) {
    return location.origin + '/api/download?data=' + encodeURIComponent(key)
}

// 目录没有下载地址，给的是本页定位到该目录的地址（对方打开后看到的就是这个目录）。
function folderUrl(path) {
    return location.origin + location.pathname + '?s=' + encodeURIComponent(path)
}

// copyText 写剪贴板。navigator.clipboard 只在安全上下文（https / localhost）可用，
// 而本服务通常跑在 http://192.168.x.x:9801 这类局域网地址上，必须准备 execCommand 兜底。
function copyText(text) {
    if (window.isSecureContext && navigator.clipboard) {
        return navigator.clipboard.writeText(text)
    }
    return new Promise(function (resolve, reject) {
        const $ta = $('<textarea readonly>').val(text)
            .css({position: 'fixed', top: '-1000px', opacity: '0'}).appendTo(document.body)
        const el = $ta[0]
        el.select()
        el.setSelectionRange(0, el.value.length)   // iOS 上只调 select() 复制不到内容
        let ok = false
        try {
            ok = document.execCommand('copy')
        } catch (err) {
            ok = false
        }
        $ta.remove()
        if (ok) resolve()
        else reject(new Error('execCommand copy failed'))
    })
}

// 复制成功时链接上临时显示的文案。
const COPIED_TEXT = '已复制'

// copyLink 复制并在链接上给出反馈：文字临时变成「已复制」并转绿，1.5 秒后复原。
// 原始文案在第一次点击时记到 data-label 上，而不是每次去读链接当前文本：1.5 秒内连点两次
// 时，第二次读到的会是「已复制」，复原就会永久卡在「已复制」上。页面上复制链接的文案
// 并不统一（「复制地址」/「复制路径」），所以只能逐元素记，不能用一个常量。
// 实在复制不了时（个别浏览器 / 权限策略）退化成弹窗让用户手动复制，不让功能静默失效。
function copyLink(obj, text) {
    const $link = $(obj)
    let label = attr(obj, 'data-label')
    if (!label) {
        label = $link.text()
        $link.attr('data-label', label)
    }
    copyText(text).then(function () {
        $link.text(COPIED_TEXT).addClass('copied')
    }).catch(function () {
        window.prompt('自动复制失败，请手动复制：', text)
    }).finally(function () {
        setTimeout(function () {
            $link.text(label).removeClass('copied')
        }, 1500)
    })
}

function copyDownload(obj) {
    copyLink(obj, downloadUrl(attr(obj, 'data-pathname-key')))
}

function copyFolder(obj) {
    copyLink(obj, folderUrl(attr(obj, 'data-path')))
}

// 复制下载库路径本身（就是那串本地绝对路径，不是链接），方便粘到文件管理器 / 命令行里用。
function copyLibPath(obj) {
    copyLink(obj, attr(obj, 'data-path'))
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
            // 弹窗里的操作链接统一放进 .action-links 容器，间隔由容器的 gap 给：
            // 这里混着 btn-link 与 action-link，逐个加 ms-2 很容易漏（原来「定位到文件目录」
            // 就没带外边距，紧贴在「复制地址」后面）。
            if (item.is_dir) {
                img = '<img class="file-icon" src="' + ICON_FOLDER + '" alt=""> '
                btn = '<span class="action-links">'
                btn += '<a class="btn-link" href="?s=' + encodeURIComponent(item.path) + '">定位到此目录</a>'
                btn += '<a onclick="copyFolder(this)" data-path="' + esc(item.path) + '" class="action-link copy-link">复制地址</a>'
                btn += '</span>'
            } else {
                img = '<img class="file-icon" src="' + fileIcon(item.name) + '" alt=""> '
                size = '<i class="text-info">' + esc(item.size) + ' ' + esc(item.size_unit) + '</i>'
                btn = '<span class="action-links">'
                btn += '<a onclick="downloadFile(this)" data-pathname-key="' + esc(item.pathname_key) + '" class="action-link">下载</a>'
                btn += '<a onclick="copyDownload(this)" data-pathname-key="' + esc(item.pathname_key) + '" class="action-link copy-link">复制地址</a>'
                btn += '<a class="btn-link" href="?s=' + encodeURIComponent(item.parent_path) + '&f=' + encodeURIComponent(item.name) + '">定位到文件目录</a>'
                btn += '</span>'
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
