package spotify

import (
	"fmt"
	"net/url"
	"strings"
)

type ResourceType int

const (
	TrackResourceType ResourceType = iota
	ArtistResourceType
	AlbumResourceType
	PlaylistResourceType
	UnknownResourceType
)

const expectedLinkHost = "open.spotify.com"

type Uri struct {
	Scheme    string
	Authority ResourceType
	Path      string
}

func StringToResourceType(s string) ResourceType {
	switch strings.ToLower(s) {
	case "track":
		return TrackResourceType
	case "artist":
		return ArtistResourceType
	case "album":
		return AlbumResourceType
	case "playlist":
		return PlaylistResourceType
	default:
		return UnknownResourceType
	}
}

func (r ResourceType) String() string {
	return []string{"track", "artist", "album", "playlist", "unknown"}[r]
}

func NewUri(s string) Uri {
	var uri Uri
	strSplit := strings.Split(s, ":")
	switch len(strSplit) {
	case 3:
		uri.Scheme = strSplit[0]
		uri.Authority = StringToResourceType(strSplit[1])
		uri.Path = strSplit[2]
	case 5:
		uri.Scheme = strSplit[0]
		uri.Authority = StringToResourceType(strSplit[3])
		uri.Path = strSplit[4]
	default:
	}

	return uri
}

func (u Uri) String() string {
	return fmt.Sprintf("%s:%s:%s", u.Scheme, u.Authority, u.Path)
}

func ConvertLinkToUri(link string) (Uri, bool) {
	var uri Uri

	parsedUrl, err := url.Parse(link)
	if err != nil {
		return uri, false
	}

	if parsedUrl.Host != expectedLinkHost {
		return uri, false
	}

	pathSplit := strings.Split(parsedUrl.Path, "/")
	if len(pathSplit) != 3 {
		return uri, false
	}

	uri.Scheme = "spotify"
	uri.Authority = StringToResourceType(pathSplit[1])
	uri.Path = pathSplit[2]

	return uri, true
}
