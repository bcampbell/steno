package store

import ()

type ArtList []ArtID

func (arts *ArtList) Days() []string {
	return []string{}
	/*XYZZY*/
	/*
		found := map[string]struct{}{}
		for _, art := range *arts {
			found[art.Day()] = struct{}{}
		}
		out := make([]string, 0, len(found))
		for day, _ := range found {
			out = append(out, day)
		}
		sort.Strings(out)
		return out
	*/
}

func (arts *ArtList) Pubs() []string {
	return []string{}
	/*XYZZY*/
	/*
		found := map[string]struct{}{}
		for _, art := range *arts {
			found[art.Pub] = struct{}{}
		}
		out := make([]string, 0, len(found))
		for pub, _ := range found {
			out = append(out, pub)
		}
		sort.Strings(out)
		return out
	*/
}

func (arts ArtList) Subtract(other ArtList) ArtList {
	lookup := map[ArtID]struct{}{}
	for _, id := range other {
		lookup[id] = struct{}{}
	}
	out := make(ArtList, 0, len(arts)-len(other))
	for _, id := range arts {
		if _, got := lookup[id]; !got {
			out = append(out, id)
		}
	}
	return out
}

// Debug helper to gauge memory usage of displayed fields...
// TODO: missing publication date,twitter-specific fields
func (arts ArtList) DumpAverages() {
	return
	/*XYZZY*/
	/*
			if len(arts) == 0 {
				return
			}
			var headlineCnt, pubCnt, sectionCnt, tagsCnt, bylineCnt, urlCnt, kwCnt int

			for _, art := range arts {
				headlineCnt += len(art.Headline)
				pubCnt += len(art.Pub)
				sectionCnt += len(art.Section)
				tagsCnt += len(art.TagsString())
				bylineCnt += len(art.BylineString())
				urlCnt += len(art.URL())
				kwCnt += len(art.KeywordsString())
			}
			n := len(arts)
			headlineCnt /= n
			pubCnt /= n
			sectionCnt /= n
			tagsCnt /= n
			bylineCnt /= n
			urlCnt /= n
			kwCnt /= n

			fmt.Printf(`-----averages-----
		headline: %d
		pub:      %d
		section:  %d
		tags:     %d
		byline:   %d
		url:      %d
		kw:       %d
		TOTAL:    %d
		`,
				headlineCnt,
				pubCnt,
				sectionCnt,
				tagsCnt,
				bylineCnt,
				urlCnt,
				kwCnt,
				headlineCnt+pubCnt+sectionCnt+tagsCnt+bylineCnt+urlCnt+kwCnt)
	*/
}
