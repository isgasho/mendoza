package mendoza

import (
	"errors"
	"sort"
)

// HashList stores a document as a flat list of entries. Each entry contains a hash of its contents, allowing you
// to quickly find equivalent sub trees.
type HashList struct {
	Entries []HashEntry
}

func HashListFor(doc interface{}) (*HashList, error) {
	hashList := &HashList{}
	err := hashList.AddDocument(doc)
	if err != nil {
		return nil, err
	}
	return hashList, nil
}

type Reference struct {
	Index int
	Key   string
}

func MapEntryReference(idx int, key string) Reference {
	return Reference{Index: idx, Key: key}
}

func SliceEntryReference(idx int) Reference {
	return Reference{Index: idx}
}

func (ref Reference) IsMapEntry() bool {
	return len(ref.Key) != 0
}

func (ref Reference) IsSliceEntry() bool {
	return len(ref.Key) == 0
}

type HashEntry struct {
	Hash      Hash
	Value     interface{}
	Size      int
	Parent    int
	Sibling   int
	Reference Reference
}

func (hashList *HashList) AddDocument(obj interface{}) error {
	_, _, err := hashList.process(-1, Reference{}, obj)
	return err
}

func (hashList *HashList) IsNonEmptyMap(idx int) bool {
	if len(hashList.Entries) <= idx+1 {
		return false
	}

	nextEntry := hashList.Entries[idx+1]
	return nextEntry.Parent == idx && nextEntry.Reference.IsMapEntry()
}

func (hashList *HashList) IsNonEmptySlice(idx int) bool {
	if len(hashList.Entries) <= idx+1 {
		return false
	}

	nextEntry := hashList.Entries[idx+1]
	return nextEntry.Parent == idx && nextEntry.Reference.IsSliceEntry()
}

func (hashList *HashList) process(parent int, ref Reference, obj interface{}) (result Hash, size int, err error) {
	current := len(hashList.Entries)

	hashList.Entries = append(hashList.Entries, HashEntry{
		Parent:    parent,
		Value:     obj,
		Reference: ref,
		Sibling:   -1,
	})

	switch obj := obj.(type) {
	case nil:
		result = HashNull
		size = 1
	case bool:
		if obj {
			result = HashTrue
		} else {
			result = HashFalse
		}
		size = 1
	case float64:
		result = HashFloat64(obj)
		size = 8
	case string:
		result = HashString(obj)
		size = len(obj) + 1
	case map[string]interface{}:
		hasher := HasherMap()
		keys := sortedKeys(obj)

		prevIdx := -1

		for idx, key := range keys {
			value := obj[key]
			entryIdx := len(hashList.Entries)
			valueHash, valueSize, err := hashList.process(current, MapEntryReference(idx, key), value)
			if err != nil {
				return result, size, err
			}

			size += len(key) + valueSize + 1

			if prevIdx != -1 {
				prevEntry := &hashList.Entries[prevIdx]
				prevEntry.Sibling = entryIdx
			}

			prevIdx = entryIdx

			hasher.WriteField(key, valueHash)
		}

		result = hasher.Sum()
	case []interface{}:
		hasher := HasherSlice()

		prevIdx := -1

		for idx, value := range obj {
			entryIdx := len(hashList.Entries)

			valueHash, valueSize, err := hashList.process(current, SliceEntryReference(idx), value)
			if err != nil {
				return result, size, err
			}

			size += valueSize + 1

			if prevIdx != -1 {
				prevEntry := &hashList.Entries[prevIdx]
				prevEntry.Sibling = entryIdx
			}

			prevIdx = entryIdx

			hasher.WriteElement(valueHash)
		}

		result = hasher.Sum()
	default:
		return result, size, errors.New("unsupported type")
	}

	hashList.Entries[current].Hash = result
	hashList.Entries[current].Size = size

	return result, size, nil
}

func (hashList *HashList) Iter(idx int) *Iter {
	return &Iter{
		hashList: hashList,
		idx:      idx + 1,
	}
}

type Iter struct {
	hashList *HashList
	idx      int
}

func (it *Iter) GetIndex() int {
	return it.idx
}

func (it *Iter) GetEntry() HashEntry {
	return it.hashList.Entries[it.idx]
}

func (it *Iter) GetKey() string {
	return it.GetEntry().Reference.Key
}

func (it *Iter) IsDone() bool {
	return it.idx == -1
}

func (it *Iter) Next() {
	it.idx = it.GetEntry().Sibling
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}