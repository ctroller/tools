package httpapi

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"syscall"
	"time"
	"uuid"

	"github.com/dustin/go-humanize"
	"trox.dev/file-converter/internal/common"
)

// cgnatBlock is the carrier-grade NAT range (RFC 6598), commonly used by cloud providers
// for internal-only addresses. net.IP.IsPrivate() does not cover it.
var cgnatBlock = func() *net.IPNet {
	_, n, err := net.ParseCIDR("100.64.0.0/10")
	if err != nil {
		panic(err)
	}
	return n
}()

// downloadClient rejects connections to loopback, private, CGNAT, link-local, unspecified,
// and multicast addresses so DownloadUrl cannot be used to reach internal network services
// (SSRF), including via a redirect to such an address. It also bounds the whole request so a
// slow or unresponsive remote server cannot hold the handling goroutine open indefinitely.
var downloadClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Control: func(_, address string, _ syscall.RawConn) error {
				host, _, err := net.SplitHostPort(address)
				if err != nil {
					return err
				}
				ip := net.ParseIP(host)
				if ip == nil {
					return fmt.Errorf("could not parse IP: %s", host)
				}
				if isDisallowedDownloadTarget(ip) {
					return common.NotAllowedErr{
						Msg: fmt.Sprintf("connections to %s are not allowed", ip),
					}
				}
				return nil
			},
		}).DialContext,
	},
}

func isDisallowedDownloadTarget(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsMulticast() ||
		cgnatBlock.Contains(ip)
}

func (a *API) DownloadUrl(w http.ResponseWriter, r *http.Request) {
	rawUrl := r.URL.Query().Get("url")
	if rawUrl == "" {
		HttpProblem(w, "", "Bad Request", http.StatusBadRequest, "Missing url parameter")
		return
	}

	slog.Info("Downloading url", "url", rawUrl)

	out := a.internalDownload(rawUrl, w)
	if out == nil {
		return
	}
	defer func(out *os.File) {
		if err := out.Close(); err != nil {
			slog.Error("Failed to close file", "file", out.Name(), "err", err)
		}
		a.fs.DeleteFile(out)
	}(out)

	result, ok := a.detectAndSubmit(w, out, func(t string) string {
		return "The requested url has an unsupported media type (" + t + ")"
	})
	if !ok {
		return
	}

	RenderJSON(w, Response[Result]{Data: result})
}

// internalDownload fetches rawUrl into a new store file and returns it open and seeked to the
// start. On any failure it writes the response, cleans up the file, and returns nil.
func (a *API) internalDownload(rawUrl string, w http.ResponseWriter) *os.File {
	_, err := url.ParseRequestURI(rawUrl)
	if err != nil {
		HttpProblem(w, "", "Bad Request", http.StatusBadRequest, "Invalid url parameter")
		return nil
	}

	id := uuid.New().String()

	out, err := a.fs.Create(id)
	if err != nil {
		HttpProblemISE(w, "Failed to create file", err)
		return nil
	}

	success := false
	defer func() {
		if !success {
			if err := out.Close(); err != nil {
				slog.Error("Failed to close file", "file", out.Name(), "err", err)
			}
			a.fs.Delete(id)
		}
	}()

	resp, err := downloadClient.Get(rawUrl)
	if err != nil {
		if notAllowedErr, ok := errors.AsType[common.NotAllowedErr](err); ok {
			HttpProblem(w, "", "Bad Request", http.StatusBadRequest, notAllowedErr.Error())
			return nil
		}

		HttpProblemISE(w, "Failed to download file", err)
		return nil
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			slog.Error("Failed to close response body", "err", err)
		}
	}(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		HttpProblem(w, "", "Bad Request", http.StatusBadRequest, fmt.Sprintf("The requested url returned status %d", resp.StatusCode))
		return nil
	}

	if cl := resp.Header.Get("Content-Length"); cl != "" {
		l, err := strconv.ParseUint(cl, 10, 64)
		if err != nil {
			HttpProblemISE(w, "Failed to parse content length", err)
			return nil
		}
		if l > uint64(a.config.MaxFileSizeBytes) {
			HttpProblem(w, "", "Request Entity Too Large", http.StatusRequestEntityTooLarge, "The requested file is too large ("+humanize.IBytes(l)+"). Allowed is a maximum of "+humanize.IBytes(uint64(a.config.MaxFileSizeBytes)))
			return nil
		}
	}

	// The Content-Length check above is only a fast path: a server can omit or lie about
	// the header, so the limit is enforced again here regardless of what was declared.
	limited := &io.LimitedReader{R: resp.Body, N: a.config.MaxFileSizeBytes + 1}
	n, err := out.ReadFrom(limited)
	if err != nil {
		HttpProblemISE(w, "Failed to write file", err)
		return nil
	}
	if n > a.config.MaxFileSizeBytes {
		HttpProblem(w, "", "Request Entity Too Large", http.StatusRequestEntityTooLarge, "The requested file is too large ("+humanize.IBytes(uint64(n))+"). Allowed is a maximum of "+humanize.IBytes(uint64(a.config.MaxFileSizeBytes)))
		return nil
	}

	if _, err = out.Seek(0, io.SeekStart); err != nil {
		HttpProblemISE(w, "Failed to seek file", err)
		return nil
	}

	success = true
	return out
}
