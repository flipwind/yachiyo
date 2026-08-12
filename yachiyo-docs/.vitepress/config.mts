import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
    title: "Project Yachiyo (开发文档)",
    description: "to build a life runtime",
    themeConfig: {
        // https://vitepress.dev/reference/default-theme-config
        nav: [
            { text: "主页", link: "/" },
        ],

        sidebar: [
            {
                text: "总览",
                items: [
                    { text: "何为 Project Yachiyo？", link: "/overview/introduce.md" }
                ]
            }
        ],

        socialLinks: [
            { icon: "github", link: "https://github.com/flipwind/yachiyo" }
        ],

        search: {
            provider: "local",
        }
    },

    cleanUrls: true,
    lastUpdated: true,
})
