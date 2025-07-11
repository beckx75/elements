package elements

type CompareData struct {
	ElementIdsOnly1 []string
	ElementIdsOnly2 []string
	ElementIdsInter []string
}

func CompareElementBevies(eb1, eb2 ElementBevy) *CompareData {
	cd := &CompareData{}

	return cd
}

func compareElementIds(eb1, eb2 ElementBevy) ([]string, []string, []string) {
	id1 := make(map[string]bool)
	id2 := make(map[string]bool)
	inter := make(map[string]bool)

	for k, _ := range eb2.Elements {
		id2[k] = true
	}

	for k, _ := range eb1.Elements {
		_, ok := eb2.Elements[k]
		if ok {
			inter[k] = true
			delete(id2, k)
		} else {
			id1[k] = true
		}
	}

	id1list := []string{}
	id2list := []string{}
	interlist := []string{}

	for k, _ := range id1 {
		id1list = append(id1list, k)
	}
	for k, _ := range id2 {
		id2list = append(id2list, k)
	}
	for k, _ := range inter {
		interlist = append(interlist, k)
	}

	return id1list, id2list, interlist
}

func compareUnionkeys(eb1, eb2 ElementBevy) ([]string, []string, []string) {
	id1 := make(map[string]bool)
	id2 := make(map[string]bool)
	inter := make(map[string]bool)

	union1, _ := eb1.GetUnionKeys()
	union2, _ := eb1.GetUnionKeys()

	for k, _ := range union2 {
		id2[k] = true
	}

	for k, _ := range union1 {
		_, ok := union2[k]
		if ok {
			inter[k] = true
			delete(id2, k)
		} else {
			id1[k] = true
		}
	}

	id1list := []string{}
	id2list := []string{}
	interlist := []string{}

	for k, _ := range id1 {
		id1list = append(id1list, k)
	}
	for k, _ := range id2 {
		id2list = append(id2list, k)
	}
	for k, _ := range inter {
		interlist = append(interlist, k)
	}

	return id1list, id2list, interlist
}
