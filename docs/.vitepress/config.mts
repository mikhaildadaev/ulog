import { defineConfig } from 'vitepress'
export default defineConfig({
  appearance: 'dark',
  base: '/ulog/',
  head: [
    ['link', { rel: 'stylesheet', href: '/ulog/styles.css' }],
    ['script', { src: '/ulog/scripts.js' }]
  ],
  lastUpdated: true,
  locales: {
    en: {
      description: 'A high-performance, zero-dependency platform for logs, metrics and traces.',
      label: 'English',
      lang: 'en',
      link: '/en/',
      title: 'ULOG',
      themeConfig: {
        nav: [
          { 
            text: 'Home', 
            link: '/en/' 
          },
          { 
            text: 'Go', 
            link: '/en/go' 
          },
          { 
            text: 'Benchmarks', 
            link: '/en/benchmarks' 
          },
          { 
            text: 'API', 
            link: '/en/core_constructors' 
          },
        ],
        sidebar: [
          {
            items: [
              { 
                text: 'Go', 
                link: '/en/go' 
              },
              { 
                text: 'Benchmarks', 
                link: '/en/benchmarks' 
              },
              { 
                text: 'API', 
                collapsed: true,
                items: [
                  { 
                    text: 'Core', 
                    collapsed: true,
                    items: [
                      { 
                        text: 'Constructors', 
                        link: '/en/core_constructors' 
                      },
                      { 
                        text: 'Options', 
                        link: '/en/core_options' 
                      },
                      { 
                        text: 'Types', 
                        link: '/en/core_types' 
                      }
                    ] 
                  },
                  {
                    text: 'SinkFile',
                    collapsed: true, 
                    items: [
                      { 
                        text: 'Constructors', 
                        link: '/en/sinkfile_constructors' 
                      },
                      { 
                        text: 'Params', 
                        link: '/en/sinkfile_params' 
                      }
                    ]
                  },
                  {
                    text: 'SinkHttp', 
                    collapsed: true,
                    items: [
                      { 
                        text: 'Constructors', 
                        link: '/en/sinkhttp_constructors' 
                      },
                      { 
                        text: 'Factories', 
                        link: '/en/sinkhttp_factories' 
                      },
                      { 
                        text: 'Params', 
                        link: '/en/sinkhttp_params' 
                      }
                    ]
                  }
                ] 
              }
            ]
          }
        ],
        darkModeSwitchLabel: "Appearance",
        darkModeSwitchTitle: "Switch to dark theme",
        lightModeSwitchTitle: "Switch to light theme",
        sidebarMenuLabel: "Menu",
        returnToTopLabel: "Return to top",
        outline: {
          label: "On this page"
        },
        lastUpdated: {
          text: "Last Updated",
          formatOptions: {
            dateStyle: "short",
            timeStyle: "short"
          }
        },
        docFooter: {
          prev: "Previous page",
          next: "Next page"
        },
        footer: {
          message: 'Released under the Apache License 2.0',
          copyright: '© 2026 Mikhail Dadaev'
        }
      }
    },
    ru: {
      description: 'Производительная платформа без зависимостей для логов, метрик и трейсов.',
      label: 'Русский',
      lang: 'ru',
      link: '/ru/',
      title: 'ULOG',
      themeConfig: {
        nav: [
          { 
            text: 'Главная', 
            link: '/ru/' 
          },
          { 
            text: 'Go', 
            link: '/ru/go' 
          },
          { 
            text: 'Бенчмарки', 
            link: '/ru/benchmarks' 
          },
          { 
            text: 'API', 
            link: '/ru/core_constructors' 
          },
        ],
        sidebar: [
          {
            items: [
              { 
                text: 'Go', 
                link: '/ru/go' 
              },
              { 
                text: 'Бенчмарки', 
                link: '/ru/benchmarks' 
              },
              {
                text: 'API', 
                collapsed: true,
                items: [
                  { 
                    text: 'Ядро', 
                    collapsed: true,
                    items: [
                      { 
                        text: 'Конструкторы', 
                        link: '/ru/core_constructors' 
                      },
                      { 
                        text: 'Опции', 
                        link: '/ru/core_options' 
                      },
                      { 
                        text: 'Типы', 
                        link: '/ru/core_types' 
                      }
                    ] 
                  },
                  { 
                    text: 'Запись в файл', 
                    collapsed: true,
                    items: [
                      { 
                        text: 'Конструкторы', 
                        link: '/ru/sinkfile_constructors' 
                      },
                      { 
                        text: 'Параметры', 
                        link: '/ru/sinkfile_params' 
                      }
                    ] 
                  },
                  { 
                    text: 'Запись по сети', 
                    collapsed: true,
                    items: [
                      { 
                        text: 'Конструкторы', 
                        link: '/ru/sinkhttp_constructors' 
                      },
                      { 
                        text: 'Фабрики', 
                        link: '/ru/sinkhttp_factories' 
                      },
                      { 
                        text: 'Параметры', 
                        link: '/ru/sinkhttp_params' 
                      }
                    ] 
                  }
                ]
              }
            ]
          }
        ],
        darkModeSwitchLabel: "Внешний вид",
        darkModeSwitchTitle: "Переключиться на тёмную тему",
        lightModeSwitchTitle: "Переключиться на светлую тему",
        sidebarMenuLabel: "Меню",
        returnToTopLabel: "Вернуться наверх",
        outline: {
          label: "Содержание страницы"
        },
        lastUpdated: {
          text: "Последние изменения",
          formatOptions: {
            dateStyle: "short",
            timeStyle: "short"
          }
        },
        docFooter: {
          prev: "Предыдущая страница",
          next: "Следующая страница"
        },
        footer: {
          message: 'Под лицензией Apache 2.0',
          copyright: '© 2026 Михаил Дадаев'
        },
      }
    },
    zh: {
      description: '一个高性能、零依赖性的日志、度量和跟踪平台。',
      label: '简体中文',
      lang: 'zh',
      link: '/zh/',
      title: 'ULOG',
      themeConfig: {
        nav: [
          { 
            text: '首页', 
            link: '/zh/' 
          },
          { 
            text: 'Go', 
            link: '/zh/go' 
          },
          { 
            text: '基准测试', 
            link: '/zh/benchmarks' 
          },
          { 
            text: 'API', 
            link: '/zh/core_constructors' 
          },
        ],
        sidebar: [
          {
            items: [
              { 
                text: 'Go', 
                link: '/zh/go' 
              },
              { 
                text: '基准', 
                link: '/zh/benchmarks' 
              },
              { 
                text: 'API', 
                collapsed: true,
                items: [
                  { 
                    text: '核心', 
                    collapsed: true,
                    items: [
                      { 
                        text: '构造函数', 
                        link: '/zh/core_constructors' 
                      },
                      { 
                        text: '选项', 
                        link: '/zh/core_options' 
                      },
                      { 
                        text: '类别', 
                        link: '/zh/core_types' 
                      }
                    ] 
                  },
                  { 
                    text: '文件接收器', 
                    collapsed: true,
                    items: [
                      { 
                        text: '构造函数', 
                        link: '/zh/sinkfile_constructors' 
                      },
                      { 
                        text: '参数 ', 
                        link: '/zh/sinkfile_params' 
                      }
                    ] 
                  },
                  { 
                    text: 'HTTP 接收器', 
                    collapsed: true,
                    items: [
                      { 
                        text: '构造函数', 
                        link: '/zh/sinkhttp_constructors' 
                      },
                      { 
                        text: '工厂', 
                        link: '/zh/sinkhttp_factories' 
                      },
                      { 
                        text: '参数', 
                        link: '/zh/sinkhttp_params' 
                      }
                    ] 
                  }
                ] 
              }
            ]
          }
        ],
        darkModeSwitchLabel: "深色模式",
        darkModeSwitchTitle: "切换至深色主题",
        lightModeSwitchTitle: "切换至浅色主题",
        sidebarMenuLabel: "目录",
        returnToTopLabel: "返回至顶部",
        outline: {
          label: "页面导航"
        },
        lastUpdated: {
          text: "最近更改",
          formatOptions: {
            dateStyle: "short",
            timeStyle: "short"
          }
        },
        docFooter: {
          prev: "上一页",
          next: "下一页"
        },
        footer: {
          message: '根据 Apache 2.0 许可证发布',
          copyright: '© 2026 Mikhail Dadaev'
        },
      }
    }
  },
  themeConfig: {
    search: {
      provider: 'local'
    },
    socialLinks: [
      { 
        icon: 'github', 
        link: 'https://github.com/mikhaildadaev/ulog' 
      }
    ],
  }
})