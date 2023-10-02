package magneturi

import (
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

const magnetPrefix = "magnet:?"

// Magnet represents a Magnet URI.
type Magnet struct {
	// ExactTopics is a "xt".
	ExactTopics []string

	// DisplayName is a "dn".
	// A filename to display to the user, for convenience.
	DisplayName string

	// ExactLength is a "xl".
	// The file size, in bytes
	ExactLength int64

	// Trackers is a "tr".
	Trackers []string

	// WebSeed is a "ws".
	// The payload data served over HTTP(S)
	// TODO(cristaloleg): clarify is this string or []string.

	// AcceptableSources is "as".
	// Refers to a direct download from a web server.
	AcceptableSources []string

	// ExactSource is a "xs".
	ExactSource []string

	// KeywordTopic is a "kt".
	// Specifies a string of search keywords to search for in P2P networks, rather than a particular file.
	KeywordTopic []string

	// ManifestTopic is a "mt".
	// Link to the metafile that contains a list of magneto.
	ManifestTopic string

	// Select Only
	// Lists specific files torrent clients should download.
	// TODO(cristaloleg): implement BEP53.

	// Extra is a "x." and other unparsed params.
	Extra map[string][]string
}

var errNoPrefix = errors.New("magnet URI prefix not found")

// Parse magnet URI.
func Parse(raw string) (*Magnet, error) {
	if !strings.HasPrefix(raw, magnetPrefix) {
		return nil, errNoPrefix
	}

	parts := strings.Split(raw, magnetPrefix)
	if len(parts) <= 1 {
		return nil, errNoPrefix
	}

	m := &Magnet{
		Extra: make(map[string][]string),
	}

	exactTopics := make(map[string]struct{})
	trackers := make(map[string]struct{})
	acceptableSources := make(map[string]struct{})

	params := strings.Split(parts[1], "&")

	for _, param := range params {
		params := strings.Split(param, "=")
		if len(params) < 2 || params[1] == "" {
			continue
		}

		switch key, value := params[0], params[1]; key {
		case "dn":
			decoded, err := url.QueryUnescape(value)
			if err != nil {
				return nil, err
			}
			m.DisplayName = decoded

		case "xt":
			exactTopics[value] = struct{}{}

		case "kt":
			m.KeywordTopic = append(m.KeywordTopic, strings.Split(value, "+")...)

		case "mt":
			m.ManifestTopic = value

		case "tr":
			v, err := url.QueryUnescape(value)
			if err != nil {
				return nil, err
			}
			trackers[v] = struct{}{}

		case "as":
			v, err := url.QueryUnescape(value)
			if err != nil {
				return nil, err
			}
			acceptableSources[v] = struct{}{}

		case "xl":
			size, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, err
			}
			m.ExactLength = size

		case "xs":
			v, err := url.QueryUnescape(value)
			if err != nil {
				return nil, err
			}
			m.ExactSource = append(m.ExactSource, v)

		default:
			v, err := url.QueryUnescape(value)
			if err != nil {
				return nil, err
			}
			m.Extra[key] = append(m.Extra[key], v)
		}
	}

	m.ExactTopics = make([]string, 0, len(exactTopics))
	for topic := range exactTopics {
		m.ExactTopics = append(m.ExactTopics, topic)
	}

	m.Trackers = make([]string, 0, len(trackers))
	for tracker := range trackers {
		m.Trackers = append(m.Trackers, tracker)
	}

	m.AcceptableSources = make([]string, 0, len(acceptableSources))
	for as := range acceptableSources {
		m.AcceptableSources = append(m.AcceptableSources, as)
	}
	return m, nil
}

// Normalize will sort all fields with multiple values.
func (m *Magnet) Normalize() {
	sort.Strings(m.ExactTopics)
	sort.Strings(m.Trackers)
	sort.Strings(m.AcceptableSources)
	sort.Strings(m.ExactSource)
	sort.Strings(m.KeywordTopic)
	for _, v := range m.Extra {
		sort.Strings(v)
	}
}

// Encode magnet URI.
func (m *Magnet) Encode() string {
	var b strings.Builder
	// TODO(cristaloleg): calculate this dynamically or estimate.
	b.Grow(512)

	b.WriteString(magnetPrefix)
	b.WriteString("dn=")
	b.WriteString(url.QueryEscape(m.DisplayName))

	for _, xt := range m.ExactTopics {
		b.WriteString("&xt=")
		b.WriteString(xt)
	}

	if m.ExactLength > 0 {
		b.WriteString("&xl=")
		b.WriteString(strconv.FormatInt(m.ExactLength, 10))
	}

	for _, tr := range m.Trackers {
		b.WriteString("&tr=")
		b.WriteString(url.QueryEscape(tr))
	}

	// TODO(cristaloleg): add ws.

	for _, as := range m.AcceptableSources {
		b.WriteString("&as=")
		b.WriteString(url.QueryEscape(as))
	}

	for _, xs := range m.ExactSource {
		b.WriteString("&xs=")
		b.WriteString(url.QueryEscape(xs))
	}

	if len(m.KeywordTopic) > 0 {
		b.WriteString("&kt=")
		b.WriteString(strings.Join(m.KeywordTopic, "+"))
	}

	if m.ManifestTopic != "" {
		b.WriteString("&mt=")
		b.WriteString(m.ManifestTopic)
	}

	// TODO(cristaloleg): add so.

	for key, param := range m.Extra {
		for _, paramValue := range param {
			b.WriteByte('&')
			b.WriteString(key)
			b.WriteByte('=')
			b.WriteString(paramValue)
		}
	}
	return b.String()
}
