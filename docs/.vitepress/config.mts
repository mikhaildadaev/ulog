import { defineConfig } from 'vitepress'
export default defineConfig({
  appearance: 'dark',
  base: '/ulog/',
  head: [
    [
      'script',
      {},
      `(function() {
        const lang = document.documentElement.lang
        if (lang) {
          localStorage.setItem('vitepress-lang', lang)
        }
        const observer = new MutationObserver(() => {
          const newLang = document.documentElement.lang
          if (newLang) {
            localStorage.setItem('vitepress-lang', newLang)
          }
        })
        observer.observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] })
      })()`
    ]
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
            link: '/en/core_main-examples' 
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
                        text: 'Main', 
                        link: '/en/core_main-examples' 
                      },
                      { 
                        text: 'Options', 
                        link: '/en/core_options-examples' 
                      },
                      { 
                        text: 'Types', 
                        link: '/en/core_types-examples' 
                      }
                    ] 
                  },
                  {
                    text: 'SinkFile',
                    collapsed: true, 
                    items: [
                      { 
                        text: 'Main', 
                        link: '/en/sinkfile_main-examples' 
                      },
                      { 
                        text: 'Params', 
                        link: '/en/sinkfile_params-examples' 
                      }
                    ]
                  },
                  {
                    text: 'SinkHttp', 
                    collapsed: true,
                    items: [
                      { 
                        text: 'Main', 
                        link: '/en/sinkhttp_main-examples' 
                      },
                      { 
                        text: 'Factories', 
                        link: '/en/sinkhttp_factories-examples' 
                      },
                      { 
                        text: 'Params', 
                        link: '/en/sinkhttp_params-examples' 
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
          message: 'Released under the Apache License 2.0.',
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
            link: '/ru/core_main-examples' 
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
                text: 'Ядро', 
                collapsed: true,
                items: [
                  { 
                    text: 'Основное', 
                    link: '/ru/core_main-examples' 
                  },
                  { 
                    text: 'Опции', 
                    link: '/ru/core_options-examples' 
                  },
                  { 
                    text: 'Типы', 
                    link: '/ru/core_types-examples' 
                  }
                ] 
              },
              { 
                text: 'Запись в файл', 
                collapsed: true,
                items: [
                  { 
                    text: 'Основное', 
                    link: '/ru/sinkfile_main-examples' 
                  },
                  { 
                    text: 'Параметры', 
                    link: '/ru/sinkfile_params-examples' 
                  }
                ] 
              },
              { 
                text: 'Запись по сети', 
                collapsed: true,
                items: [
                  { 
                    text: 'Основное', 
                    link: '/ru/sinkhttp_main-examples' 
                  },
                  { 
                    text: 'Фабрики', 
                    link: '/ru/sinkhttp_factories-examples' 
                  },
                  { 
                    text: 'Параметры', 
                    link: '/ru/sinkhttp_params-examples' 
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
          message: 'Распространяется под лицензией Apache 2.0.',
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
            link: '/zh/core_main-examples' 
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
                        text: '主要', 
                        link: '/zh/core_main-examples' 
                      },
                      { 
                        text: '选项', 
                        link: '/zh/core_options-examples' 
                      },
                      { 
                        text: '类别', 
                        link: '/zh/core_types-examples' 
                      }
                    ] 
                  },
                  { 
                    text: '写入文件', 
                    collapsed: true,
                    items: [
                      { 
                        text: '主要', 
                        link: '/zh/sinkfile_main-examples' 
                      },
                      { 
                        text: '帕拉姆斯', 
                        link: '/zh/sinkfile_params-examples' 
                      }
                    ] 
                  },
                  { 
                    text: '通过网络录制', 
                    collapsed: true,
                    items: [
                      { 
                        text: '主要', 
                        link: '/zh/sinkhttp_main-examples' 
                      },
                      { 
                        text: '工厂', 
                        link: '/zh/sinkhttp_factories-examples' 
                      },
                      { 
                        text: '帕拉姆斯', 
                        link: '/zh/sinkhttp_params-examples' 
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
          message: '根据 Apache 2.0 许可证发布。',
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