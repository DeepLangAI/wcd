let navConfig = [
    {
        label: '解析配置',
        submenu: [
            {
                'text': '站点规则',
                'link': '/public/html/rule.html'
            },
            {
                'text': '解析实验',
                'link': '/public/html/parse_test.html'
            },
        ]
    },
    // {
    //     label: '解析配置',
    //     submenu: [
    //         {
    //             'text': '站点规则2-1',
    //             'link': '/public/html/rule.html'
    //         },
    //         {
    //             'text': '站点规则2-2',
    //             'link': '/public/html/rule.html'
    //         },
    //         {
    //             'text': '站点规则2-3',
    //             'link': '/public/html/rule.html'
    //         },
    //     ]
    // },
];

function generateNav(activeMenu = null, config = null) {
    if (config === null) {
        config = navConfig; // 默认配置
    }

    // 创建侧栏元素
    let $sidebar = $(`
<div class="sidebar" id="sidebar">
    <button class="toggle-button">☰</button>
    <h2>Web Content Distill</h2>
    <ul id="nav-list"></ul>
</div>`);

    let $navList = $sidebar.find('#nav-list');
    let $toggleButton = $sidebar.find('.toggle-button');

    // 切换侧栏
    $toggleButton.on('click', function() {
        $sidebar.toggleClass('collapsed');
    });

    // 生成导航栏
    $.each(config, function(index, item) {
        let $li = $('<li class="toggle"></li>').text(item.label);
        let $icon = $('<span class="icon">▶</span>');

        $navList.append($li);

        if (item.submenu.length > 0) {
            $li.append($icon);
            let $submenu = $('<ul class="submenu"></ul>');
            $.each(item.submenu, function(i, subItem) {
                let $subMentItem = $('<li></li>')
                    .addClass('submenu-item')
                    .text(subItem.text);
                if (subItem.text === activeMenu) {
                    $subMentItem.addClass('active');
                }
                $subMentItem.click(() => {
                    window.location.href = subItem.link;
                })
                $submenu.append($subMentItem);
            });
            $navList.append($submenu);

            // 默认展开子菜单
            $submenu.css('display', 'block');
            $icon.addClass('collapsed'); // 更新箭头状态
        }

        // 点击事件
        $li.on('click', function() {
            $(this).next('.submenu').slideToggle();
            $icon.toggleClass('collapsed'); // 切换旋转效果
        });
    });

    // 将侧栏添加到 body 或特定容器中
    $('body').append($sidebar); // 或者替换为特定容器，例如 $('#container').append($sidebar);
    return $sidebar;
}