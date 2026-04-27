import { defineConfig } from 'vitepress'
export default defineConfig({
  appearance: 'dark',
  base: '/ulog/',
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
                        text: 'Reference', 
                        link: '/en/core_reference-examples' 
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
        footer: {
          message: 'Released under the Apache License 2.0.',
          copyright: 'Copyright © 2026 Mikhail Dadaev'
        }
      }
    },
    ru: {
      description: 'Высокопроизводительная платформа без зависимостей для логов, метрик и трейсов.',
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
            link: '/ru/api' 
          },
        ],
        sidebar: [
          {
            items: [
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
                    text: 'Форматы', 
                    link: '/ru/core_reference-examples' 
                  }
                ] 
              },
              { 
                text: 'Файловый приёмник', 
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
                text: 'HTTP приёмник', 
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
        footer: {
          message: 'Распространяется под лицензией Apache 2.0.',
          copyright: '© 2026 Михаил Дадаев'
        },
      }
    },
    zh: {
      description: '高性能、零依赖的日志、指标和追踪平台。',
      label: '简体中文',
      lang: 'zh-CN',
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
            link: '/zh/api' 
          },
        ],
        sidebar: [
          {
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
                    text: '格式', 
                    link: '/zh/core_reference-examples' 
                  }
                ] 
              },
              { 
                text: '接收器 File', 
                collapsed: true,
                items: [
                  { 
                    text: '主要', 
                    link: '/zh/sinkfile_main-examples' 
                  },
                  { 
                    text: '参数', 
                    link: '/zh/sinkfile_params-examples' 
                  }
                ] 
              },
              { 
                text: '接收器 Http', 
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
                    text: '参数', 
                    link: '/zh/sinkhttp_params-examples' 
                  }
                ] 
              }
            ]
          }
        ],
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