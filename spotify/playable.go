package spotify

//type Playable interface {
//	Id() string
//	Name() string
//	Image() string
//}

//func (a Playlist) Id() string {
//	return utils.ConvertTo62(a.spotifyArtist.GetGid())
//}
//
//func (a Playlist) Name() string {
//	return a.spotifyArtist.GetName()
//}
//
//func (a Playlist) Bio() string {
//	bio := a.spotifyArtist.GetBiography()
//	if len(bio) > 0 {
//		return bio[0].GetText()
//	}
//	return ""
//}
//
//func (a Playlist) Image() string {
//	image := a.spotifyArtist.GetPortraitGroup().GetImage()
//	if len(image) > 0 {
//		return fmt.Sprintf("https://i.scdn.co/image/%032s", hex.EncodeToString(image[0].GetFileId()))
//	}
//	return ""
//}
//
//func (a Playlist) TopTracks() []string {
//	topTracks := a.spotifyArtist.GetTopTrack()
//	if len(topTracks) == 0 {
//		return nil
//	}
//
//	var ids []string
//	for _, track := range topTracks[0].GetTrack() {
//		ids = append(ids, fmt.Sprintf("%s", utils.ConvertTo62(track.GetGid())))
//	}
//
//	return ids
//}
