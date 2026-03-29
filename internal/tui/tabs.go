package tui

import (
	"strings"
	"time"

	"github.com/hawkaii/obia/internal/task"
)

type Tab int

const (
	TabTasks Tab = iota
	TabToday
	TabOverdue
	TabCalDAV
	tabCount
)

var tabNames = [tabCount]string{
	"Tasks",
	"Today",
	"Overdue",
	"CalDAV",
}

func (t Tab) String() string {
	return tabNames[t]
}

func renderTabBar(active Tab, width int) string {
	var tabs []string
	for i := Tab(0); i < tabCount; i++ {
		if i == active {
			tabs = append(tabs, activeTabStyle.Render("["+i.String()+"]"))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(" "+i.String()+" "))
		}
	}
	bar := strings.Join(tabs, "")
	return tabBarStyle.Width(width).Render(bar)
}

func filterTasksForTab(tasks []task.Task, tab Tab, dailyNotesFolder, dailyNotesFormat string) []task.Task {
	switch tab {
	case TabToday:
		return filterToday(tasks, dailyNotesFolder, dailyNotesFormat)
	case TabOverdue:
		return filterOverdue(tasks)
	case TabCalDAV:
		return filterCalDAV(tasks)
	default:
		return filterOpen(tasks)
	}
}

func filterOpen(tasks []task.Task) []task.Task {
	var out []task.Task
	for i := range tasks {
		if !tasks[i].IsDone() {
			out = append(out, tasks[i])
		}
	}
	return out
}

func filterToday(tasks []task.Task, dailyNotesFolder, dailyNotesFormat string) []task.Task {
	today := time.Now()
	todayStr := today.Format(dailyNotesFormat)
	var out []task.Task
	for i := range tasks {
		t := &tasks[i]
		// Tasks from today's daily note
		if strings.Contains(t.Source.FilePath, dailyNotesFolder+"/"+todayStr) {
			out = append(out, *t)
			continue
		}
		// Tasks due today
		if t.Due != nil && sameDay(*t.Due, today) {
			out = append(out, *t)
		}
	}
	return out
}

func filterOverdue(tasks []task.Task) []task.Task {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var out []task.Task
	for i := range tasks {
		t := &tasks[i]
		if t.Due != nil && t.Due.Before(today) && !t.IsDone() {
			out = append(out, *t)
		}
	}
	return out
}

func filterCalDAV(tasks []task.Task) []task.Task {
	var out []task.Task
	for i := range tasks {
		if tasks[i].CalDAVUID != "" {
			out = append(out, tasks[i])
		}
	}
	return out
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
