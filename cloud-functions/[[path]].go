package handler

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dsnet/compress/bzip2"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

const (
	faviconPath   = "/favicon.ico"
	faviconBase64 = "AAABAAEAICAAAAEAIACoEAAAFgAAACgAAAAgAAAAQAAAAAEAIAAAAAAAABAAAMMOAADDDgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADpzLQA6cy0AOnMtADpzLQA6cy0AE5HOQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA6cy0AOnMtADpzLQA6cy0AunMtADpzLQA6cy0AGhcUAAAAAAAAAAAAAAAAAAAAAAAir7mAIq/5QCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIm95wD//wAA6cy0AOnMtCXpzLSf6cy0denMtAfpzLQA6cy0AHhqXAAAAAAAir7mAIq+5gCKvuYAir7mAIq+5gCJv+UAib/lAIm/5QCJv+UAib/lAIm/5QCJv+UAib/lAIm/5QCJv+UAib/lAIm/5QCJv+UAib/lAIm/5QCJv+UAhr7nAP7PqQDnzLUl6cy0wOnMtP/pzLT66cy0gunMtAbpzLQA6cy0AIq+5gCKvuYAir7mAIq+5huKvuZlir7mkoq+5pmKvuaYir7mmIq+5piKvuaYir7mmIq+5piKvuaYir7mmIq+5piKvuaYir7mmIq+5piKvuaYir7mmIq+5piIvueYocHamOPLt8zqzLT/6cy0/+nMtP/pzLT66cy0gunMtAfpzLQAir7mAIq+5gCKvuY3ir7myoq+5v+Kvub/ir7m/4q+5v+Kvub/ir7m/4q+5v+Kvub/ir7m/4q+5v+Kvub/ir7m/4q+5v+Kvub/ir7m/4q+5v+Kvub/ib7m/5XA4P/QyMH/6sy0/+XLtv/ozLT/5sy1/+fMtf/pzLT66cy0fenMtASKvuYAir7mFoq+5sWKvub/ir7m/4q+5v+Kvub8ir7m/Iq+5vyKvub8ir7m/Iq+5vyKvub8ir7m/Iq+5vyKvub8ir7m/Iq+5vyKvub8ir7m/4q+5v+Ivuf+qMLW/OjMtP/Yyr39ucXN/efMtf/Xyb7/usXN/+XLtv/qzLTN6cy0FYq+5gCKvuZXir7m/Iq+5v+KvubYir7mZ4q+5lGKvuZSir7mUoq+5lKKvuZSir7mUoq+5lKKvuZSir7mUoq+5lKKvuZSir7mUIq+5nGKvubyir7m/4q+5tScwd1Z48u3YrnFzVS8xcyB6cy0+tTJv/+Rv+L/p8LX/73Fy30ApP8Air7mAIq+5n2Kvub/ir7m/4q+5naKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mKoq+5uyKvub/ir7mwHm77wXQyMEAz8jCAO7NskPqzLP41cm//4+/4/+Ivuf/iL7ne42/5ACKvuYAir7mgoq+5v+Kvub/ir7mZoq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYsir7m44q+5vqKvua8ir7mE4q+5gkAqP8D6cy0RurMs/jVyb//j7/j/4q+5v+KvuaCir7mAIq+5gCKvuaBir7m/4q+5v+KvuZnir7mAIq+5gSKvuZCir7mPYq+5gGKvuYAir7mAIq+5h6KvuZSir7mF4q+5gCKvuYAir7mAIq+5guKvuY6ir7mPYq+5l6Kvua7ir7mxIe+6H/mzLZI6syz+NXJv/+Pv+P/ir7m/4q+5oGKvuYAir7mAIq+5oGKvub/ir7m/4q+5meKvuYAir7mOoq+5rGKvuamir7mX4q+5gCKvuYmir7mooq+5qWKvuaDir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mPIq+5vqKvub/iL7nrOXLtkrqzLP61cm+/4+/4/+Kvub/ir7mgYq+5gCKvuYAir7mgYq+5v+Kvub/ir7mZ4q+5gCKvuYuir7mqYq+5kuKvuadir7meYq+5qGKvuZbir7mhYq+5nOKvuYAir7mAIq+5gCKvuYDir7mEYq+5hCKvuZIir7m5Iq+5vGIvuee48u3MuvMs9rLyMTwjL7l/4q+5v+KvuaBir7mAIq+5gCKvuaBir7m/4q+5v+KvuZnir7mAIq+5gCKvuZLir7mp4q+5kmKvuawir7mn4q+5oyKvuaair7mCoq+5gCKvuYAir7mAIq+5iWKvubCir7m1Yq+5qmKvuY0ir7mLom+5h/Xyb4D9c6uIpvB3X6Kvub/ir7m/4q+5oGKvuYAir7mAIq+5oGKvub/ir7m/4q+5meKvuYAir7mAIq+5gCKvuZLir7mpYq+5kmKvuaIir7mo4q+5qmKvuZDir7mAIq+5gCKvuYAir7mLoq+5vGKvub/ir7mw4q+5gWKvuYAir7mANDIwQCsw9QAib7nZYq+5v+Kvub/ir7mgYq+5gCKvuYAir7mgYq+5v+Kvub/ir7mZ4q+5gCKvuYAir7mAIq+5hSKvua/ir7mpIq+5gWKvuY6ir7mf4q+5qeKvuYQir7mAIq+5gCKvuYoir7m0Iq+5uSKvuaxir7mJoq+5h+KvuYVir7mAIq+5gCKvuZnir7m/4q+5v+KvuaBir7mAIq+5gCKvuaBir7m/4q+5v+KvuZnir7mAIq+5hmKvuZjir7mn4q+5paKvuagir7mKIq+5riKvuazir7m0Yq+5iCKvuYAir7mAIq+5gWKvuYcir7mHIq+5k6KvubZir7m5Iq+5piKvuYDir7mAIq+5meKvub/ir7m/4q+5oGKvuYAir7mAIq+5oGKvub/ir7m/4q+5meKvuYAir7mbIq+5syKvuayir7mGIq+5qCKvuZfir7mp4q+5oaKvuZpir7mC4q+5gCKvuYAir7mAIq+5gCKvuYAir7mPIq+5vqKvub/ir7msYq+5gOKvuYAir7mZ4q+5v+Kvub/ir7mgYq+5gCKvuYAir7mgYq+5v+Kvub/ir7mZ4q+5gCKvuZiir7mtYq+5mSKvuYAir7mM4q+5puKvua2ir7mlIq+5geKvuYAir7mAIq+5gCKvuYIir7mKoq+5iuKvuZWir7my4q+5tWKvuaOir7mA4q+5gCKvuZnir7m/4q+5v+KvuaBir7mAIq+5gCKvuaBir7m/4q+5v+KvuZnir7mAIq+5gWKvuYPib/mAIq+5gCKvuYAir7mCYq+5hmKvuYIir7mAIrG5QCKvuYAir7mAIq+5iqKvubbir7m8Iq+5reKvuYbir7mE4q+5g2KvuYAir7mAIq+5maKvub/ir7m/4q+5oKKvuYAir7mAIq+5oGKvub/ir7m/4q+5meKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mK4q+5u2Kvub/ir7mwIq+5gWKvuYAir7mAIq+5gCKvuYAir7mcYq+5v+Kvub/ir7mfoq+5gCKvuYAir7mgYq+5v+Kvub/ir7mZ4q+5gAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAir7mAIq+5gWKvuY0ir7mP4q+5j2KvuZhir7m8Yq+5v+KvubQir7mRoq+5j6KvuY/ir7mPoq+5lOKvubOir7m/4q+5v6KvuZcir7mAIq+5gCKvuaCir7m/4q+5v+KvuZmir7mAIq+5gAAAAAAAAAAAAAAAAAAAAAAAAAAAIq+5gCKvuYAir7mLoq+5uKKvub3ir7m9Yq+5veKvub+ir7m/4q+5v2Kvub2ir7m9Yq+5vWKvub1ir7m+oq+5v+Kvub/ir7mz4q+5huKvuYAir7mAIq+5n2Kvub/ir7m/4q+5naKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuZ3ir7m/4q+5v+Kvub/ir7m/4q+5v+Kvub/ir7m/4q+5v+Kvub/ir7m/4q+5v+Kvub/ir7m/4q+5tiKvuZEir7mAIq+5QCKvuYAir7mV4q+5vyKvub/ir7m2Iq+5meKvuZRir7mUoq+5lKKvuZSir7mUoq+5lKKvuZRir7mbIq+5t+Kvub/ir7m+oq+5r2Kvuasir7mrYq+5q2Kvuatir7mrYq+5q2Kvuatir7mroq+5qiKvuZ7ir7mJoq+5gCKvuYAir7mAIq+5gCKvuYWir7mxYq+5v+Kvub/ir7m/4q+5vyKvub8ir7m/Iq+5vyKvub8ir7m/Iq+5vyKvub/ir7m/4q+5v+Kvua7ir7mE4q/5gKKvuYDir7mA4q+5gOKvuYDir7mA4q+5gOKvuYDir7mAoq+5gCKvuYAir7mAIq+5gAAAAAAir7mAIq+5gCKvuY3ir7myoq+5v+Kvub/ir7m/4q+5v+Kvub/ir7m/4q+5v+Kvub/ir7m/4q+5v+Kvub+ir7mw4q+5i+KvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gAAAAAAAAAAAAAAAACKvuYAir7mAIq+5gCKvuYbir7mZoq+5pKKvuaZir7mmIq+5piKvuaYir7mmIq+5piKvuaZir7mkYq+5mGKvuYXir7mAIq+5gCKvuYAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACKvuYAir7mAIq+5gCKvuYAir7mAIm/5QCJv+UAib/lAIm/5QCJv+UAib/lAIq/5gCKvuYAir7mAIq+5gCKvuYAir7mAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAir7mAIq/5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCKvuYAir7mAIq+5gCJv+UAir7mAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA////A////gHgAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA/gAAAHwAAAAAAAAAAAAAAAAAAEAAAAHAAAf/4AAP//gAP////////////8="

	baiduVerificationPath    = "/baidu_verify_codeva-12lSsugOP8.html"
	baiduVerificationContent = "4cc8f0df575059784be959e98981a89f"
)

type archiveKind uint8

const (
	archiveZIP archiveKind = iota
	archiveTAR
	archiveGZip
	archiveBZip2
	archiveZstd
	archiveXZ
)

type outputFormat struct {
	plainText   bool
	archiveKind archiveKind
	extension   string
	contentType string
}

var defaultFormat = outputFormat{
	archiveKind: archiveXZ,
	extension:   "tar.xz",
	contentType: "application/x-xz",
}

func Handler(writer http.ResponseWriter, req *http.Request) {
	serveEmbeddedFile := func(writer http.ResponseWriter, req *http.Request, data []byte, contentType string) {
		writer.Header().Set("Content-Type", contentType)
		writer.Header().Set("Content-Length", strconv.Itoa(len(data)))

		if req.Method != http.MethodHead {
			_, _ = writer.Write(data)
		}
	}

	switch req.URL.Path {
	case faviconPath:
		content, err := base64.StdEncoding.DecodeString(faviconBase64)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		serveEmbeddedFile(writer, req, []byte(content), "image/x-icon")
		return

	case baiduVerificationPath:
		serveEmbeddedFile(writer, req, []byte(baiduVerificationContent), "text/html; charset=utf-8")
		return
	}

	setNoStoreHeaders(writer)
	format := resolveFormat(req.URL.Query().Get("format"))
	headers := collectHeaders(req)

	if format.plainText {
		writeResponse(writer, req, "text/plain; charset=utf-8", "headers.txt", buildText(headers))
		return
	}

	payload, err := buildArchive(format.archiveKind, headers)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writeResponse(writer, req, format.contentType, "headers."+format.extension, payload)
}

func setNoStoreHeaders(writer http.ResponseWriter) {
	writer.Header().Set("Cache-Control", "no-store, max-age=0")
	writer.Header().Set("Pragma", "no-cache")
	writer.Header().Set("Vary", "*")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
}

func writeResponse(writer http.ResponseWriter, req *http.Request, contentType, filename string, payload []byte) {
	writer.Header().Set("Content-Type", contentType)
	writer.Header().Set("Content-Length", strconv.Itoa(len(payload)))
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))

	if req.Method != http.MethodHead {
		_, _ = writer.Write(payload)
	}
}

func resolveFormat(value string) outputFormat {
	switch value {
	case "txt", "text", "plaintext":
		return outputFormat{plainText: true}
	case "zip":
		return outputFormat{archiveKind: archiveZIP, extension: "zip", contentType: "application/zip"}
	case "tar":
		return outputFormat{archiveKind: archiveTAR, extension: "tar", contentType: "application/x-tar"}
	case "tgz":
		return outputFormat{archiveKind: archiveGZip, extension: "tgz", contentType: "application/gzip"}
	case "tar.gz":
		return outputFormat{archiveKind: archiveGZip, extension: "tar.gz", contentType: "application/gzip"}
	case "tar.bz2":
		return outputFormat{archiveKind: archiveBZip2, extension: "tar.bz2", contentType: "application/x-bzip2"}
	case "tar.zst":
		return outputFormat{archiveKind: archiveZstd, extension: "tar.zst", contentType: "application/zstd"}
	case "tar.xz":
		return defaultFormat
	default:
		return defaultFormat
	}
}

func collectHeaders(req *http.Request) map[string][]string {
	headers := make(map[string][]string, len(req.Header)+1)
	for rawName, values := range req.Header {
		name := http.CanonicalHeaderKey(rawName)
		headers[name] = append(headers[name], values...)
	}
	headers["Host"] = append(headers["Host"], req.Host)
	return headers
}

func buildText(headers map[string][]string) []byte {
	var output bytes.Buffer
	for name, values := range headers {
		for _, value := range values {
			output.WriteString(name)
			output.WriteString(": ")
			output.WriteString(value)
			output.WriteByte('\n')
		}
	}
	return output.Bytes()
}

func buildArchive(kind archiveKind, headers map[string][]string) ([]byte, error) {
	var output bytes.Buffer

	switch kind {
	case archiveZIP:
		writer := zip.NewWriter(&output)
		if err := writeZIP(writer, headers); err != nil {
			return nil, err
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}
	case archiveTAR:
		if err := writeTAR(&output, headers); err != nil {
			return nil, err
		}
	case archiveGZip:
		writer := gzip.NewWriter(&output)
		if err := writeTAR(writer, headers); err != nil {
			return nil, err
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}
	case archiveBZip2:
		writer, err := bzip2.NewWriter(&output, nil)
		if err != nil {
			return nil, err
		}
		if err := writeTAR(writer, headers); err != nil {
			return nil, err
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}
	case archiveZstd:
		writer, err := zstd.NewWriter(&output)
		if err != nil {
			return nil, err
		}
		if err := writeTAR(writer, headers); err != nil {
			return nil, err
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}
	case archiveXZ:
		writer, err := xz.NewWriter(&output)
		if err != nil {
			return nil, err
		}
		if err := writeTAR(writer, headers); err != nil {
			return nil, err
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}
	}

	return output.Bytes(), nil
}

func writeZIP(writer *zip.Writer, headers map[string][]string) error {
	for name, values := range headers {
		header := &zip.FileHeader{
			Name:     name,
			Method:   zip.Deflate,
			Modified: time.Now(),
		}
		file, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := file.Write(headerContent(values)); err != nil {
			return err
		}
	}
	return nil
}

func writeTAR(dst io.Writer, headers map[string][]string) error {
	writer := tar.NewWriter(dst)
	for name, values := range headers {
		content := headerContent(values)
		header := &tar.Header{
			Name:    name,
			Mode:    0o444,
			Size:    int64(len(content)),
			ModTime: time.Now(),
			Format:  tar.FormatUSTAR,
		}
		if err := writer.WriteHeader(header); err != nil {
			return err
		}
		if _, err := writer.Write(content); err != nil {
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}

func headerContent(values []string) []byte {
	return []byte(strings.Join(values, "\n"))
}
