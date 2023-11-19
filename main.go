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

const favoriteKey = "favorites"

var a fyne.App

func main() {
	scheduleUrl := "https://golab.io/_next/data/eKsW0aSFaA1iGmNQIYfTS/schedule.json"
	scheduleReq, err := http.Get(scheduleUrl)
	if err != nil {
		panic(err)
	}
	scheduleJson, err := ioutil.ReadAll(scheduleReq.Body)
	if err != nil {
		panic(err)
	}
	scheduleReq.Body.Close()

	var data pageData
	if err := json.Unmarshal(scheduleJson, &data); err != nil {
		panic(err)
	}

	a = app.NewWithID("de.thielepaul.golab2023")
	w := a.NewWindow("Hello World")
	a.Preferences()

	favorites := a.Preferences().StringListWithFallback(favoriteKey, []string{})

	dayViews := []*widget.AccordionItem{}
	for _, day := range data.PageProps.Edition.Days {
		day := day
		list := widget.NewList(
			func() int {
				return len(day.Schedule)
			},
			func() fyne.CanvasObject {
				return widget.NewLabel("template")
			},
			func(i widget.ListItemID, o fyne.CanvasObject) {
				text := fmt.Sprint(day.Schedule[i].Time.Local().Format("15:04"), " - ", day.Schedule[i].Title)
				if slices.Contains(favorites, day.Schedule[i].Id) {
					text = fmt.Sprint("⭐ ", text)
				}
				o.(*widget.Label).SetText(text)
			})
		list.OnSelected = func(id widget.ListItemID) {
			recordId := day.Schedule[id].Id
			if slices.Contains(favorites, recordId) {
				favorites = slices.Delete(favorites, slices.Index(favorites, recordId), slices.Index(favorites, recordId)+1)
			} else {
				favorites = append(favorites, recordId)
			}
			a.Preferences().SetStringList(favoriteKey, favorites)
			list.RefreshItem(id)
		}

		dayViews = append(dayViews, widget.NewAccordionItem(day.Title, list))
	}

	daysView := widget.NewAccordion(dayViews...)

	w.SetContent(daysView)
	w.ShowAndRun()
}
