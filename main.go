package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/exp/slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

const favoriteKey = "favorites"

func main() {
	data, err := getData()
	if err != nil {
		panic(err)
	}

	a := app.NewWithID("de.thielepaul.golab2023")
	w := a.NewWindow("GoLab 2023 Schedule")

	state := State{
		app:       a,
		favorites: a.Preferences().StringListWithFallback(favoriteKey, []string{}),
	}

	daysView := widget.NewAccordion(state.buildDaysView(data)...)

	w.SetContent(daysView)
	w.ShowAndRun()
}

type State struct {
	app       fyne.App
	favorites []string
}

func (s State) buildDaysView(data []day) []*widget.AccordionItem {
	dayViews := []*widget.AccordionItem{}
	for _, day := range data {
		day := day
		list := widget.NewList(
			func() int {
				return len(day.Schedule)
			},
			func() fyne.CanvasObject {
				return widget.NewLabel("loading...")
			},
			func(id widget.ListItemID, o fyne.CanvasObject) {
				text := fmt.Sprint(day.Schedule[id].Time.Local().Format("15:04"), " - ", day.Schedule[id].Title)
				if s.isFavorite(day.Schedule[id].Id) {
					text = fmt.Sprint("⭐ ", text)
				}
				o.(*widget.Label).SetText(text)
			})
		list.OnSelected = s.toggleFavorite(day, list)

		dayViews = append(dayViews, widget.NewAccordionItem(day.Title, list))
	}
	return dayViews
}

func (s State) isFavorite(id string) bool {
	return slices.Contains(s.favorites, id)
}

func (s State) toggleFavorite(d day, list *widget.List) func(id widget.ListItemID) {
	return func(id widget.ListItemID) {
		recordId := d.Schedule[id].Id
		if slices.Contains(s.favorites, recordId) {
			s.favorites = slices.Delete(s.favorites, slices.Index(s.favorites, recordId), slices.Index(s.favorites, recordId)+1)
		} else {
			s.favorites = append(s.favorites, recordId)
		}
		s.app.Preferences().SetStringList(favoriteKey, s.favorites)
		list.RefreshItem(id)
		time.Sleep(100 * time.Millisecond)
		list.Unselect(id)
	}
}

type pageData struct {
	PageProps pageProps `json:"pageProps"`
}

type pageProps struct {
	Edition edition `json:"edition"`
}

type edition struct {
	Days []day `json:"days"`
}

type day struct {
	Title    string   `json:"title"`
	Schedule []Record `json:"schedule"`
}

type Record struct {
	Id                string    `json:"id"`
	Title             string    `json:"title"`
	Time              time.Time `json:"time"`
	DurationInMinutes int       `json:"durationInMinutes"`
	Text              string    `json:"text"`
}

func getData() ([]day, error) {
	scheduleURL := "https://golab.io/_next/data/eKsW0aSFaA1iGmNQIYfTS/schedule.json"
	resp, err := http.Get(scheduleURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	scheduleJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data pageData
	if err := json.Unmarshal(scheduleJSON, &data); err != nil {
		return nil, err
	}

	return data.PageProps.Edition.Days, nil
}
