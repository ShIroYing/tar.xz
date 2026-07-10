use std::io::{Cursor, Write};

use anyhow::{Context, Result, anyhow};
use js_sys::{Array, Uint8Array};
use leptos::mount::mount_to;
use leptos::prelude::*;
use leptos_use::use_document;
use lzma_rust2::{XzOptions, XzWriter};
use obfstr::obfstr;
use wasm_bindgen::JsCast;
use wasm_bindgen::prelude::*;
use web_sys::{Blob, BlobPropertyBag, HtmlAnchorElement, Url};

fn get_love() -> String { obfstr!("白彩恋超爱枫莹的！").into() }

#[wasm_bindgen(start)]
pub fn main() -> std::result::Result<(), JsValue> {
	console_error_panic_hook::set_once();
	mount_app().map_err(|err| JsValue::from_str(&format!("{err:?}")))
}

fn mount_app() -> Result<()> {
	let body = use_document().body().context("missing body")?;
	body.set_inner_html("");
	mount_to(body, || view! { <App /> }).forget();
	Ok(())
}

#[component]
fn App() -> impl IntoView {
	let selected = RwSignal::new(ArchiveKind::Empty);

	view! {
		<main class="page">
			<div class="controls">
				<select
					aria-label="选择 tar.xz"
					prop:value=move || selected.get().value()
					on:change=move |event| selected.set(ArchiveKind::from_value(&event_target_value(&event)))
				>
					<option value="empty">"空空如也的 tar.xz"</option>
					<option value="love">"充满爱意的 tar.xz"</option>
				</select>
				<button type="button" on:click=move |_| generate_and_download(selected.get_untracked()).unwrap_or_default()>
					"来个它！"
				</button>
			</div>
			<footer>"© tar.xz.cn"</footer>
		</main>
	}
}

#[derive(Clone, Copy, PartialEq, Eq)]
enum ArchiveKind {
	Empty,
	Love,
}

impl ArchiveKind {
	const fn file_name(self) -> &'static str {
		match self {
			Self::Empty => "empty.tar.xz",
			Self::Love => "love.tar.xz",
		}
	}

	const fn value(self) -> &'static str {
		match self {
			Self::Empty => "empty",
			Self::Love => "love",
		}
	}

	fn from_value(value: &str) -> Self {
		match value {
			"love" => Self::Love,
			_ => Self::Empty,
		}
	}
}

fn generate_and_download(kind: ArchiveKind) -> Result<()> {
	let mut tar_bytes = vec![];
	{
		let mut builder = tar::Builder::new(&mut tar_bytes);
		match kind {
			ArchiveKind::Empty => append_empty(&mut builder)?,
			ArchiveKind::Love => append_txt(&mut builder, "love.txt", &get_love())?,
		}
		builder.finish()?;
	}

	let mut xz_bytes = vec![];
	{
		let mut encoder = XzWriter::new(&mut xz_bytes, XzOptions::with_preset(9))?;
		encoder.write_all(&tar_bytes)?;
		encoder.finish()?;
	}

	download_bytes(kind.file_name(), &xz_bytes)
}

fn append_empty(builder: &mut tar::Builder<impl Write>) -> Result<()> {
	let mut header = tar::Header::new_gnu();
	header.set_entry_type(tar::EntryType::Directory);
	header.set_size(0);
	header.set_mode(0o555);
	header.set_mtime(0);
	header.set_cksum();

	builder
		.append_data(&mut header, "./", Cursor::new(Vec::<u8>::new()))
		.context("append root directory")
}

fn append_txt(builder: &mut tar::Builder<impl Write>, name: &str, content: &str) -> Result<()> {
	let content = content.as_bytes();
	let mut header = tar::Header::new_gnu();
	header.set_size(content.len() as _);
	header.set_mode(0o444);
	header.set_mtime(0);
	header.set_cksum();

	builder
		.append_data(&mut header, name, Cursor::new(content))
		.with_context(|| format!("append {name}"))
}

fn download_bytes(file_name: &str, bytes: &[u8]) -> Result<()> {
	let bytes = Uint8Array::from(bytes);
	let parts = Array::new();
	parts.push(&bytes.buffer());

	let props = BlobPropertyBag::new();
	props.set_type("application/x-xz");
	let blob =
		Blob::new_with_u8_array_sequence_and_options(&parts, &props).map_err(|err| anyhow!("create blob failed: {err:?}"))?;
	let url = Url::create_object_url_with_blob(&blob).map_err(|err| anyhow!("create url failed: {err:?}"))?;

	let result = trigger_download(file_name, &url);
	Url::revoke_object_url(&url).ok();
	result
}

fn trigger_download(file_name: &str, url: &str) -> Result<()> {
	let document = use_document();
	let document = document.as_ref().context("missing document")?;
	let body = document.body().context("missing body")?;
	let anchor = document
		.create_element("a")
		.map_err(|err| anyhow!("create anchor failed: {err:?}"))?
		.dyn_into::<HtmlAnchorElement>()
		.map_err(|_| anyhow!("create anchor failed"))?;

	anchor.set_href(url);
	anchor.set_download(file_name);
	anchor
		.set_attribute("style", "display: none")
		.map_err(|err| anyhow!("hide anchor failed: {err:?}"))?;

	body.append_child(&anchor)
		.map_err(|err| anyhow!("attach anchor failed: {err:?}"))?;
	anchor.click();
	body.remove_child(&anchor).ok();

	Ok(())
}
