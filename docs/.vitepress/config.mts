import { defineConfig } from 'vitepress'
export default defineConfig({
  appearance: 'dark',
  locales: {
    root: {
      description: 'A high-performance, zero-dependency platform for logs, metrics, and traces.',
      label: 'English',
      lang: 'en',
      title: 'ULOG',
      themeConfig: {
        nav: [
          { 
            text: 'Home', 
            link: '/' 
          },
          { 
            text: 'API', 
            link: '/api' 
          },
          { 
            text: 'Benchmarks', 
            link: '/benchmarks' 
          },
        ],
        sidebar: [
          {
            items: [
              { 
                text: 'Core', 
                collapsed: true,
                items: [
                  { text: 'Main', link: '/core_main-examples' },
                  { text: 'Options', link: '/core_options-examples' },
                  { text: 'Reference', link: '/core_reference-examples' }
                ] 
              },
              { 
                text: 'FileSink',
                collapsed: true, 
                items: [
                  { 
                    text: 'Main', 
                    link: '/sinkfile_main-examples' 
                  },
                  { 
                    text: 'Params', 
                    link: '/sinkfile_params-examples' 
                  }
                ] 
              },
              { 
                text: 'HTTPSink', 
                collapsed: true,
                items: [
                  { 
                    text: 'Main', 
                    link: '/sinkhttp_main-examples' 
                  },
                  { 
                    text: 'Params', 
                    link: '/sinkhttp_params-examples' 
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
      description: 'Высокопроизводительная платформа для логов, метрик и трейсов без зависимостей.',
      label: 'Русский',
      lang: 'ru',
      link: '/ru/',
      title: 'ULOG',
      themeConfig: {
        nav: [
          { text: 'Главная', link: '/ru/' },
          { text: 'API', link: '/ru/api' },
          { text: 'Бенчмарки', link: '/ru/benchmarks' },
        ],
        sidebar: [
          {
            items: [
              { 
                text: 'Ядро', 
                collapsed: true,
                items: [
                  { text: 'Основное', link: '/ru/core_main-examples' },
                  { text: 'Опции', link: '/ru/core_options-examples' },
                  { text: 'Форматы', link: '/ru/core_reference-examples' }
                ] 
              },
              { 
                text: 'Приёмники', 
                collapsed: true,
                items: [
                  { text: 'Файл', link: '/ru/sink_file-examples' },
                  { text: 'HTTP', link: '/ru/sink_http-examples' }
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
          { text: '首页', link: '/zh/' },
          { text: 'API', link: '/zh/api' },
          { text: '基准测试', link: '/zh/benchmarks' },
        ],
        sidebar: [
          {
            items: [
              { 
                text: '核心', 
                collapsed: true,
                items: [
                  { text: '主要', link: '/zh/core_main-examples' },
                  { text: '类型', link: '/zh/core_options-examples' },
                  { text: '字段', link: '/zh/core_reference-examples' }
                ] 
              },
              { 
                text: '接收器', 
                collapsed: true,
                items: [
                  { text: '文件', link: '/zh/sink_file-examples' },
                  { text: 'HTTP', link: '/zh/sink_http-examples' }
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
      { icon: 'github', link: 'https://github.com/mikhaildadaev/ulog' }
    ],
  }
})