package categories

import (
	"sort"
	"strings"
	"sync"

	"github.com/Nv7-Github/Nv7Haven/eod/eodsort"
	"github.com/Nv7-Github/Nv7Haven/eod/types"
	"github.com/bwmarrin/discordgo"
)

func (b *Categories) CategoriesCmd(elem string, m types.Msg, rsp types.Rsp) {
	db, res := b.GetDB(m.GuildID)
	if !res.Exists {
		return
	}

	rsp.Acknowledge()

	el, res := db.GetElementByName(elem)
	if !res.Exists {
		rsp.ErrorMessage(res.Message)
		return
	}

	// Get Categories
	catsMap := make(map[catSortInfo]types.Empty)
	db.RLock()
	for _, cat := range db.Cats() {
		_, exists := cat.Elements[el.ID]
		if exists {
			catsMap[catSortInfo{
				Name: cat.Name,
				Cnt:  len(cat.Elements),
			}] = types.Empty{}
		}
	}
	db.RUnlock()
	cats := make([]catSortInfo, len(catsMap))
	i := 0
	for k := range catsMap {
		cats[i] = k
		i++
	}

	// Sort categories by count
	sort.Slice(cats, func(i, j int) bool {
		return cats[i].Cnt > cats[j].Cnt
	})

	// Convert to array
	out := make([]string, len(cats))
	for i, cat := range cats {
		out[i] = cat.Name
	}

	b.base.NewPageSwitcher(types.PageSwitcher{
		Kind: types.PageSwitchInv,
		Title: db.Config.LangProperty("ElemCategories", map[string]interface{}{
			"Element": el.Name,
			"Count":   len(out),
		}),
		PageGetter: b.base.InvPageGetter,
		Items:      out,
		User:       m.Author.ID,
		Thumbnail:  el.Image,
	}, m, rsp)
}

func (b *Categories) DownloadCatCmd(catName string, sort string, postfix bool, m types.Msg, rsp types.Rsp) {
	db, res := b.GetDB(m.GuildID)
	if !res.Exists {
		return
	}

	var els map[int]types.Empty
	var lock *sync.RWMutex
	catv, res := db.GetCat(catName)
	if !res.Exists {
		vcat, res := db.GetVCat(catName)
		if !res.Exists {
			rsp.ErrorMessage(res.Message)
			return
		}
		catName = vcat.Name
		els, res = b.base.CalcVCat(vcat, db)
		if !res.Exists {
			rsp.ErrorMessage(res.Message)
			return
		}
	} else {
		lock = catv.Lock
		els = catv.Elements
		catName = catv.Name
	}

	type catSortVal struct {
		id   int
		name string
	}
	db.RLock()
	elems := make([]catSortVal, len(els))
	i := 0
	if lock != nil {
		lock.RLock()
	}
	for elem := range els {
		el, _ := db.GetElement(elem, true)
		elems[i] = catSortVal{elem, el.Name}
		i++
	}
	if lock != nil {
		lock.RUnlock()
	}
	db.RUnlock()

	eodsort.Sort(elems, len(elems), func(index int) int {
		return elems[index].id
	}, func(index int) string {
		return elems[index].name
	}, func(index int, val string) {
		elems[index].name = val
	}, sort, m.Author.ID, db, postfix)

	out := &strings.Builder{}
	for _, elem := range elems {
		out.WriteString(elem.name + "\n")
	}
	buf := strings.NewReader(out.String())

	channel, err := b.dg.UserChannelCreate(m.Author.ID)
	if rsp.Error(err) {
		return
	}

	_, err = b.dg.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Content: db.Config.LangProperty("NameDownloadedCat", catName),
		Files: []*discordgo.File{
			{
				Name:        "cat.txt",
				ContentType: "text/plain",
				Reader:      buf,
			},
		},
	})
	if rsp.Error(err) {
		return
	}
	rsp.Message(db.Config.LangProperty("CatSentToDMs", nil))
}
