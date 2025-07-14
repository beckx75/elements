package elements

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	pathSeparator       string = "/"
	valsJoinString      string = "|>"
	LINE_BREAK_IN_ATTRS string = "<-'"
	PARENT_SEP          string = "||"
)

var keyMapTrueVals = []string{"j", "y", "ja", "Ja", "yes"}

type ElementBevy struct {
	Elementindex []string
	Elements     map[string]Element
	Levels       [][]string
	ElementTree  map[string]string
}

type Element struct {
	Id      string
	Name    string
	Path    string
	Level   int
	Parent  string
	Childs  []string
	KeyVals map[string][]string
}

type KeyMapRow struct {
	SrcName            string
	DstName            string
	UseSrcNameInTarget bool
}

func (eb *ElementBevy) MakePaths() {
	eb.ElementTree = make(map[string]string)
	for _, e := range eb.Elements {
		path := e.Id
		walkParents(eb, e.Id, &path)
		fmt.Println("Path:", path)
		fmt.Println("<--------- DONE --------------->")
		e.Path = path
	}
}

func walkParents(eb *ElementBevy, eid string, path *string) {
	fmt.Println("current EID:", eid)
	if eb.Elements[eid].Parent != "" {
		parents := strings.Split(eb.Elements[eid].Parent, PARENT_SEP)
		for _, parent := range parents {
			fmt.Println("\tparent:", parent)
			fmt.Println("\tpath:", *path)
			*path = parent + pathSeparator + eid
			walkParents(eb, parent, path)
		}
	} else {
		return
	}
}

func ImportSpreadsheet(rec [][]string, headerrow int, idcol, namecol, parentcol, childcol string,
	createOrigin bool) (*ElementBevy, error) {
	eb := NewElementBevy(createOrigin)
	noHeaderIdx := []int{}
	headernames := []string{}
	for i, headername := range rec[headerrow] {
		if headername == "" {
			noHeaderIdx = append(noHeaderIdx, i)
			continue
		}
		idx := findIdx(headernames, headername)
		if idx >= 0 {
			return new(ElementBevy), fmt.Errorf("ERROR non-unique Headername '%s'", headername)
		}
		headernames = append(headernames, headername)
	}

	idcolidx, namecolidx, parentcolidx, childcolidx := colStringToIdx(
		headernames, idcol, namecol, parentcol, childcol,
	)
	if idcolidx < 0 {
		return nil, fmt.Errorf("ID-Col is not a valid headername nor an index: %s", idcol)
	}

	for i, row := range rec {
		if i <= headerrow {
			continue
		}
		if idcolidx >= len(row) {
			return nil, fmt.Errorf("index out of range: id-col-idx %d is greater than row length %d (no ID-Value Column)",
				idcolidx, len(row),
			)
		}
		id := row[idcolidx]
		parent := ""
		name := ""
		childs := []string{}
		if (parentcolidx >= 0) && (parentcolidx < len(row)) {
			parent = row[parentcolidx]
		}
		if (namecolidx >= 0) && (namecolidx < len(row)) {
			name = row[namecolidx]
		}
		if (childcolidx >= 0) && (childcolidx < len(row)) {
			childs = append(childs, row[childcolidx])
		}

		e := NewElement(id, name, parent, "", childs, -1)
		for k, val := range row {
			if isIn(noHeaderIdx, k) {
				continue
			}
			if k >= len(headernames) {
				return nil, fmt.Errorf("ERROR - index out of range: cell-val-index %d greater than headernames %d with value %s",
					k, len(headernames), val)
			}
			e.AddKeyVal(headernames[k], val)
		}
		eb.AddElement(e, false)
	}
	return eb, nil
}

func NewElement(id, name, parent, path string, childs []string, level int) Element {
	var e = Element{
		Id: id, Name: name,
		Parent: parent, Path: path,
		Level: level,
	}
	if childs == nil {
		e.Childs = []string{}
	} else {
		e.Childs = childs
	}
	e.KeyVals = make(map[string][]string)
	return e
}

func NewElementBevy(makeroot bool) *ElementBevy {
	var eb = new(ElementBevy)
	eb.Elements = make(map[string]Element)
	eb.Levels = [][]string{}
	if makeroot {
		var e = NewElement("/", "root", "", "/", nil, 0)
		eb.Elements[e.Id] = e
		eb.Elementindex = []string{e.Id}
	} else {
		eb.Elementindex = []string{}
	}
	return eb
}

func ParseKeyMap(rec [][]string, hasHeader bool) ([]KeyMapRow, []string, error) {
	km := []KeyMapRow{}
	srcColnames := []string{}

	for i, row := range rec {
		if hasHeader && (i == 0) {
			continue
		}
		dstname := ""
		if len(row) == 0 {
			return []KeyMapRow{}, []string{}, fmt.Errorf("ERROR - found row with no data in it: row-idx %d", i)
		}
		srcColnames = append(srcColnames, row[0])
		switch len(row) {
		case 1:
			continue
		case 2:
			if findIdx(keyMapTrueVals, row[1]) >= 0 {
				km = append(km, KeyMapRow{SrcName: row[0], DstName: row[0], UseSrcNameInTarget: false})
			}
		case 3:
			if findIdx(keyMapTrueVals, row[1]) >= 0 {
				if row[2] == "" {
					dstname = row[0]
				} else {
					dstname = row[2]
				}
				km = append(km, KeyMapRow{SrcName: row[0], DstName: dstname, UseSrcNameInTarget: false})
			}
		case 4:
			if findIdx(keyMapTrueVals, row[1]) >= 0 {
				if row[2] == "" {
					dstname = row[0]
				} else {
					dstname = row[2]
				}
				if findIdx(keyMapTrueVals, row[3]) >= 0 {
					km = append(km, KeyMapRow{SrcName: row[0], DstName: dstname, UseSrcNameInTarget: true})
				} else {
					km = append(km, KeyMapRow{SrcName: row[0], DstName: dstname, UseSrcNameInTarget: false})
				}
			}
		}
	}
	return km, srcColnames, nil
}

func (e *Element) AddKeyVal(key, val string) {
	_, ok := e.KeyVals[key]
	if ok {
		e.KeyVals[key] = append(e.KeyVals[key], val)
	} else {
		e.KeyVals[key] = []string{val}
	}
}

func (e *Element) AppendTextToKeyValue(key, text string) bool {
	_, ok := e.KeyVals[key]
	if !ok {
		return false
	}
	newvals := []string{}
	for _, val := range e.KeyVals[key] {
		newvals = append(newvals, (val + text))
	}
	e.KeyVals[key] = newvals
	return true
}

func (e *Element) DelSuffixFromKeyValue(key, text string) bool {
	_, ok := e.KeyVals[key]
	if !ok {
		return false
	}
	newvals := []string{}
	for _, val := range e.KeyVals[key] {
		newvals = append(newvals, strings.TrimSuffix(val, text))
	}
	e.KeyVals[key] = newvals
	return true
}

// GetValues returns concatenated slice of values from key or empty string
func (e Element) GetValues(key string) (string, error) {
	_, ok := e.KeyVals[key]
	if ok {
		return strings.Join(e.KeyVals[key], valsJoinString), nil
	} else {
		return "", fmt.Errorf("key %s not a member of Element.KeyVals", key)
	}
}

func (eb *ElementBevy) AddElement(elem Element, check_parent bool) error {
	_, ok := eb.Elements[elem.Id]
	if ok {
		return fmt.Errorf("ERROR: element wiht id %s already in element-bevy", elem.Id)
	} else {
		eb.Elements[elem.Id] = elem
		eb.Elementindex = append(eb.Elementindex, elem.Id)
	}
	return nil
}

func (eb *ElementBevy) AddElements(elems []Element, check_parent bool) error {
	var err error
	for _, elem := range elems {
		err = eb.AddElement(elem, check_parent)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Error - skip this element")
		}
	}
	return nil
}

// AddStringToKeyValValues adds given string 'pat' to 'key'-values
// returns slice of element.ids from elements, where key is not a member of
// Element.KeyVals
// if 'addAsSuffix' is true -> appends pat to values, otherwith adds pat at
// the begining of the values
func (eb *ElementBevy) AddStringToKeyValValues(key, pat string, addAsSuffix bool) []string {
	keyNotFoundElementIds := []string{}
	for id, e := range eb.Elements {
		_, ok := e.KeyVals[key]
		if ok {
			newvals := []string{}
			for _, val := range e.KeyVals[key] {
				if addAsSuffix {
					newvals = append(newvals, fmt.Sprintf("%s%s", val, pat))
				} else {
					newvals = append(newvals, fmt.Sprintf("%s%s", pat, val))
				}
			}
			e.KeyVals[key] = newvals
		} else {
			keyNotFoundElementIds = append(keyNotFoundElementIds, id)
		}
	}
	return keyNotFoundElementIds
}

func (eb *ElementBevy) AppendTextToKeyValues(key, text string) []string {
	keynotfoundElementIds := []string{}
	for id, element := range eb.Elements {
		if !element.AppendTextToKeyValue(key, text) {
			keynotfoundElementIds = append(keynotfoundElementIds, id)
		}
	}
	return keynotfoundElementIds
}

func (eb *ElementBevy) DelElement(eId string) error {
	_, ok := eb.Elements[eId]
	if !ok {
		return fmt.Errorf("ERROR: key-not-found in ElementBevy.Elements: %s", eId)
	}
	eIdx := findIdx(eb.Elementindex, eId)
	parent := eb.Elements[eId].Parent
	childs := eb.Elements[eId].Childs

	for _, child := range childs {
		childElem, ok := eb.Elements[child]
		if !ok {
			return fmt.Errorf("ERROR: child with id '{child}' is not a member of ElementBevy.Elements :(")
		} else {
			childElem.Parent = ""
		}
	}

	if parent != "" {
		parElement, ok := eb.Elements[parent]
		if !ok {
			return fmt.Errorf("ERROR: parent with id '{parent}' is not a member of ElementBevy.Elements :(")
		} else {
			newchilds := []string{}
			for _, pchild := range parElement.Childs {
				if pchild == parent {
					continue
				} else {
					newchilds = append(newchilds, pchild)
				}
			}
			parElement.Childs = newchilds
			eb.Elements[parent] = parElement
		}
	}

	if eIdx >= 0 {
		// TODO: for go 1.20 there is a func in slices-package slice = slices.Delete(slice, 1, 3)
		eb.Elementindex = append(eb.Elementindex[:eIdx], eb.Elementindex[eIdx+1:]...)
	}
	delete(eb.Elements, eId)
	return nil
}

func (eb *ElementBevy) DelElementByKeyVal(key, val string) ([]string, int) {
	// deletes Element if Element.keyVals has given string 'val' for 'key'
	deletedElementsCount := 0
	deletedElementIds := []string{}

	for eId, e := range eb.Elements {
		_, ok := e.KeyVals[key]
		if !ok {
			continue
		}
		if e.KeyVals[key][0] == val {
			deletedElementsCount++
			deletedElementIds = append(deletedElementIds, eId)
		}
	}

	for _, eId := range deletedElementIds {
		eb.DelElement(eId)
	}

	return deletedElementIds, deletedElementsCount
}

func (eb *ElementBevy) DelSuffixFromKeyValues(key, text string) []string {
	keyNotFoundElementIds := []string{}
	for id, element := range eb.Elements {
		if !element.DelSuffixFromKeyValue(key, text) {
			keyNotFoundElementIds = append(keyNotFoundElementIds, id)
		}
	}
	return keyNotFoundElementIds
}

// collects all used KeyVals-Keys from elements in ElementBevy
//
// returns Map with Key and it's occurence as well as a list with
// union keys
func (eb ElementBevy) GetAllKeyValsKeys() (map[string]int, []string) {
	keyOccurence := make(map[string]int)
	unionkeys := []string{}

	for _, elm := range eb.Elements {
		for key, _ := range elm.KeyVals {
			_, ok := keyOccurence[key]
			if ok {
				keyOccurence[key]++
			} else {
				keyOccurence[key] = 1
				unionkeys = append(unionkeys, key)
			}
		}
	}

	return keyOccurence, unionkeys
}

// collects all values from given key.
// returns
// - union-Table with values and occurences in ElementBevy
// - uniqueVals: this values are unique in ElementBevy for the key
// - nonuniqueVals: values, which are more than once in ElementBevy for the key
// e.g.: i get results from all the values in all elements for a key "Streetname"
func (eb ElementBevy) GetAllValues(key string) (map[string]int, []string, []string) {
	valOccurences := make(map[string]int)
	uniqueVals := []string{}
	nonUniqueVals := []string{}

	for _, e := range eb.Elements {
		_, ok := e.KeyVals[key]
		if !ok {
			continue
		}
		val, _ := e.GetValues(key)
		_, ok = valOccurences[val]
		if ok {
			valOccurences[val]++
			valIdx := findIdx(uniqueVals, val)
			if valIdx >= 0 {
				uniqueVals = append(uniqueVals[:valIdx], uniqueVals[valIdx+1:]...)
				if findIdx(nonUniqueVals, val) < 0 {
					nonUniqueVals = append(nonUniqueVals, val)
				}
			} else {
				uniqueVals = append(uniqueVals, val)
			}
		} else {
			valOccurences[val] = 1
			uniqueVals = append(uniqueVals, val)
		}
	}

	return valOccurences, uniqueVals, nonUniqueVals
}

func (eb ElementBevy) GetChildInfos(e Element) [][]string {
	rec := [][]string{}

	for _, eid := range e.Childs {
		rec = append(rec, []string{eid, eb.Elements[eid].Name})
	}

	return rec
}

func (eb ElementBevy) GetChildInfosById(eId string) [][]string {
	return eb.GetChildInfos(eb.Elements[eId])
}

func (eb ElementBevy) GetElement(idx int) Element {
	return eb.Elements[eb.Elementindex[idx]]
}

func (eb ElementBevy) GetElementInfosByLevel(lvl int) [][]string {
	rec := [][]string{}

	if lvl < 0 {
		return rec
	}

	for _, e := range eb.Elements {
		if e.Level == lvl {
			rec = append(rec, []string{e.Id, e.Name})
		}
	}

	return rec
}

func (eb ElementBevy) GetElementLevel(e Element) int {
	if e.Parent == "" {
		return 0
	}
	foundParent := true
	level := 0
	curParent := e.Parent
	for foundParent {
		curElement, ok := eb.Elements[curParent]
		if ok {
			level++
			curParent = curElement.Parent
		} else {
			foundParent = false
		}
	}
	return level
}

func (eb ElementBevy) GetElementLevelById(eid string) int {
	return eb.GetElementLevel(eb.Elements[eid])
}

func (eb ElementBevy) GetRootElementInfo() [][]string {
	res := [][]string{}
	for _, e := range eb.Elements {
		if e.Parent == "" {
			res = append(res, []string{e.Id, e.Name})
		}
	}
	return res
}

func (eb ElementBevy) GetSameLevelElementInfos(e Element) [][]string {
	res := [][]string{}

	elevel := eb.GetElementLevel(e)
	for _, e := range eb.Elements {
		if e.Level == elevel {
			res = append(res, []string{e.Id, e.Name})
		}
	}
	return res
}

func (eb ElementBevy) GetSameLevelElementInfosById(eId string) [][]string {
	return eb.GetSameLevelElementInfos(eb.Elements[eId])
}

func (eb ElementBevy) GetUnionKeys() (map[string]int, []string) {
	keyOccurence := make(map[string]int)
	keys := []string{}
	for _, e := range eb.Elements {
		for key, _ := range e.KeyVals {
			_, ok := keyOccurence[key]
			if ok {
				keyOccurence[key]++
			} else {
				keyOccurence[key] = 1
				keys = append(keys, key)
			}
		}
	}
	return keyOccurence, keys
}

func (eb *ElementBevy) ReplaceKeyValuePattern(old, new string) int {
	replacedKeyVals := 0
	for _, e := range eb.Elements {
		for k, vals := range e.KeyVals {
			newvals := []string{}
			for _, val := range vals {
				if strings.Contains(val, old) {
					newval := strings.ReplaceAll(val, old, new)
					newvals = append(newvals, newval)
				} else {
					newvals = append(newvals, val)
				}
			}
			e.KeyVals[k] = newvals
		}
	}
	return replacedKeyVals
}

func (eb *ElementBevy) ReplaceKeyValueVals(srchpat, replpat string) int {
	replacedKeyVals := 0
	for _, e := range eb.Elements {
		for k, vals := range e.KeyVals {
			newvals := []string{}
			for _, val := range vals {
				if val == srchpat {
					newvals = append(newvals, replpat)
					replacedKeyVals++
				} else {
					newvals = append(newvals, val)
				}
			}
			e.KeyVals[k] = newvals
		}
	}
	return replacedKeyVals
}

func (eb *ElementBevy) ReplaceKeyValueValsByKey(key, oldval, newval string) (int, int) {
	replacedKeyVals := 0
	keyNotFound := 0
	for _, e := range eb.Elements {
		vals, ok := e.KeyVals[key]
		if !ok {
			keyNotFound++
			continue
		}
		newvals := []string{}
		for _, val := range vals {
			if val == oldval {
				newvals = append(newvals, newval)
				replacedKeyVals++
			} else {
				newvals = append(newvals, val)
			}
		}
		e.KeyVals[key] = newvals
	}
	return replacedKeyVals, keyNotFound
}

func (eb ElementBevy) ToSpreadsheet() ([][]string, error) {
	// TODO Konzept for integration element-fields Id, Name, Path, Level, Parent, Childs
	rec := [][]string{}
	_, headerrow := eb.GetAllKeyValsKeys()
	sort.Strings(headerrow)

	rec = append(rec, headerrow)
	for _, eid := range eb.Elementindex {
		// for _, e := range eb.Elements{
		e, ok := eb.Elements[eid]
		if !ok {
			return nil, fmt.Errorf("found Element-ID in Elementindex which is not a member of Elements: %s", eid)
		}
		row := []string{}
		for _, col := range headerrow {
			vals, ok := e.KeyVals[col]
			if ok {
				val := strings.Join(vals, valsJoinString)
				row = append(row, val)
			} else {
				row = append(row, "")
			}
		}
		rec = append(rec, row)
	}
	return rec, nil
}

func (eb ElementBevy) ToSpreadsheetWithKeyMap(km []KeyMapRow) ([][]string, error) {
	// TODO Konzept for integration element-fields Id, Name, Path, Level, Parent, Childs
	rec := [][]string{}
	headerrow := []string{}
	for _, kmr := range km {
		if findIdx(headerrow, kmr.DstName) < 0 {
			headerrow = append(headerrow, kmr.DstName)
		}
	}
	sort.Strings(headerrow)

	rec = append(rec, headerrow)
	for _, eid := range eb.Elementindex {
		// for _, e := range eb.Elements{
		e, ok := eb.Elements[eid]
		if !ok {
			return nil, fmt.Errorf("found Element-ID in Elementindex which is not a member of Elements: %s", eid)
		}
		row := []string{}
		for i := 0; i < len(headerrow); i++ {
			row = append(row, "")
		}

		for _, kmr := range km {
			vals, ok := e.KeyVals[kmr.SrcName]
			if !ok {
				fmt.Printf("EEERRROOORRR---------- have Sourcename in Keymap which is not in Element.KeyVals: %s\n",
					kmr.SrcName)
				continue
			}
			val := strings.Join(vals, valsJoinString)
			if val == "" {
				continue
			}

			if kmr.UseSrcNameInTarget {
				val = fmt.Sprintf("\"%s:\" \"%s\"", kmr.SrcName, val)
			}

			headerIdx := findIdx(headerrow, kmr.DstName)
			if row[headerIdx] == "" {
				row[headerIdx] = val
			} else {
				row[headerIdx] = fmt.Sprintf("%s%s%s", row[headerIdx], LINE_BREAK_IN_ATTRS, val)
			}
		}
		rec = append(rec, row)
	}
	return rec, nil
}

// colStringtoidx returns column-index from column-strings
// if idcol is not in headernames => column-strings ar treated as int-indices for all
// the column-strings
func colStringToIdx(headernames []string, idcol, namecol, parentcol, childcol string) (
	int, int, int, int,
) {
	var err error
	idcolidx := findIdx(headernames, idcol)
	namecolidx := findIdx(headernames, namecol)
	parentcolidx := findIdx(headernames, parentcol)
	childcolidx := findIdx(headernames, childcol)

	if idcolidx < 0 {
		idcolidx, err = strconv.Atoi(idcol)
		if err != nil {
			idcolidx = -1
		}
		namecolidx, err = strconv.Atoi(namecol)
		if err != nil {
			namecolidx = -1
		}
		parentcolidx, err = strconv.Atoi(parentcol)
		if err != nil {
			parentcolidx = -1
		}
		childcolidx, err = strconv.Atoi(childcol)
		if err != nil {
			childcolidx = -1
		}
	}
	return idcolidx, namecolidx, parentcolidx, childcolidx
}

func findIdx(list []string, item string) int {
	for i, val := range list {
		if val == item {
			return i
		}
	}
	return -1
}

func hasSrcName(km []KeyMapRow, s string) (bool, []int) {
	// TODO here better use another konzept; perhaps a map!?
	idxs := []int{}
	for i, kmr := range km {
		if kmr.SrcName == s {
			idxs = append(idxs, i)
		}
	}
	return false, idxs
}

func isIn(l []int, i int) bool {
	for _, v := range l {
		if i == v {
			return true
		}
	}
	return false
}
