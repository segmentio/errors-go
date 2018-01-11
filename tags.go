package errors

import "sort"

// Tag is a key/value type used to represent a single error tag.
type Tag struct {
	Name  string
	Value string
}

// T returns a Tag value with the given name and value.
func T(name string, value string) Tag {
	return Tag{
		Name:  name,
		Value: value,
	}
}

func makeTagsMap(tags ...Tag) map[string]string {
	if len(tags) == 0 {
		return nil
	}
	m := make(map[string]string, len(tags))
	for _, t := range tags {
		if _, exists := m[t.Name]; !exists {
			m[t.Name] = t.Value
		}
	}
	return m
}

func makeTagsFromMap(m map[string]string) []Tag {
	if len(m) == 0 {
		return nil
	}
	tags := make([]Tag, 0, len(m))
	for k, v := range m {
		tags = append(tags, T(k, v))
	}
	sortTags(tags)
	return tags
}

func makeTags(tags ...Tag) []Tag {
	tags = copyTags(tags)
	sortTags(tags)
	return tags
}

func copyTags(tags []Tag) []Tag {
	tcpy := make([]Tag, len(tags))
	copy(tcpy, tags)
	return tcpy
}

func deepAppendTags(tags []Tag, err error) []Tag {
	walk(err, func(err error) {
		tags = appendTags(tags, err)
	})
	return tags
}

func appendTags(tags []Tag, err error) []Tag {
	if e, ok := err.(errorTags); ok {
		tags = append(tags, e.Tags()...)
	}
	sortTags(tags)
	return tags
}

func sortTags(tags []Tag) {
	sort.Sort(tagsByNameAndValue(tags))
}

type tagsByNameAndValue []Tag

func (t tagsByNameAndValue) Len() int {
	return len(t)
}

func (t tagsByNameAndValue) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t tagsByNameAndValue) Less(i, j int) bool {
	ti := t[i]
	tj := t[j]
	if ti.Name < tj.Name {
		return true
	}
	if ti.Name > tj.Name {
		return false
	}
	return t[i].Value < t[j].Value
}
