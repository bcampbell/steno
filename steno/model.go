package main

import (
	"github.com/bcampbell/badger"
	"github.com/bcampbell/badger/query"
	"regexp"
	"sort"
)

type byPublished []*Article

func (s byPublished) Len() int {
	return len(s)
}

func (s byPublished) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byPublished) Less(i, j int) bool {
	return s[i].Published > s[j].Published
}

var defaultField string = "content"

func getPublications() ([]string, error) {
	var arts []*Article
	coll.Find(badger.NewAllQuery(), &arts)
	pubSet := make(map[string]struct{})
	for _, art := range arts {
		pubSet[art.Pub] = struct{}{}
	}
	var pubs []string
	for pub, _ := range pubSet {
		pubs = append(pubs, pub)
	}
	sort.Strings(pubs)

	return pubs, nil
}

// return an individual article by id
func getArtByID(objID string) *Article {
	var results []*Article
	q := badger.NewExactQuery("id", objID)
	coll.Find(q, &results)

	if len(results) == 0 {
		return nil
	}
	return results[0]
}

// search performs a search and returns the results
func search(queryString string) ([]*Article, error) {
	q, err := query.Parse(queryString, coll.ValidFields(), defaultField)
	if err != nil {
		return nil, err
	}

	var arts []*Article
	coll.Find(q, &arts)

	sort.Sort(byPublished(arts))

	return arts, nil
}

func buildQuery(queryString string) (badger.Query, error) {
	return query.Parse(queryString, coll.ValidFields(), defaultField)
}

func buildQueryFromIDs(ids []string) (badger.Query, error) {
	q := badger.NewExactQuery("id", ids...)
	return q, nil
}

func addTag(q badger.Query, tag string) (int, error) {

	changed := coll.Update(q, func(doc interface{}) {
		art := doc.(*Article)
		for _, t := range art.Tags {
			if t == tag {
				return // already got it
			}
		}
		art.Tags = append(art.Tags, tag)
	})

	return changed, nil
}

func removeTag(q badger.Query, tag string) (int, error) {

	changed := coll.Update(q, func(doc interface{}) {
		art := doc.(*Article)
		newTags := []string{}
		for _, t := range art.Tags {
			if t == tag {
				continue
			}
			newTags = append(newTags, t)
		}
		art.Tags = newTags
	})
	return changed, nil
}

func zap(q badger.Query) int {
	var arts []*Article
	coll.Find(q, &arts)

	for _, art := range arts {
		coll.Remove(art)
	}
	return len(arts)
}

func fileNameFromQuery(q string) string {
	colon := regexp.MustCompile(`:\s*`)
	spc := regexp.MustCompile(`\s+`)
	chars := regexp.MustCompile(`[^-\w]`)
	f := q
	f = colon.ReplaceAllString(f, "-")
	f = spc.ReplaceAllString(f, "_")
	f = chars.ReplaceAllString(f, "")
	return f
}
