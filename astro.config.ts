import sitemap from '@astrojs/sitemap'
import tailwindcss from '@tailwindcss/vite'
import htmlMinifier, { type HTMLMinifierOptions } from 'astro-html-minifier-next'
import robots from 'astro-robots'
import { defineConfig } from 'astro/config'
import topLevelAwait from 'vite-plugin-top-level-await'
import wasm from 'vite-plugin-wasm'

const site = 'https://tar.xz.cn'

const minifyHtmlOptions = {
	collapseBooleanAttributes: true,
	collapseInlineTagWhitespace: true,
	collapseWhitespace: true,
	decodeEntities: true,
	keepClosingSlash: false,
	mergeScripts: true,
	minifyCSS: true,
	minifyJS: true,
	minifySVG: true,
	minifyURLs: true,
	removeAttributeQuotes: true,
	removeComments: true,
	removeOptionalTags: true,
	removeRedundantAttributes: true,
	removeScriptTypeAttributes: true,
	removeStyleLinkTypeAttributes: true,
	useShortDoctype: true
} satisfies HTMLMinifierOptions

export default defineConfig({
	site,
	build: { assets: '_shiro2ying' },
	integrations: [
		sitemap({
			lastmod: new Date(),
			priority: 1
		}),
		robots({
			policy: [{ userAgent: '*', allow: '/' }],
			host: new URL(site).hostname
		}),
		htmlMinifier(minifyHtmlOptions)
	],
	vite: {
		plugins: [tailwindcss(), wasm(), topLevelAwait()]
	}
})