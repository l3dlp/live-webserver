package livews

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

var (
	homeDir     string
	lastFolder  string
	server      *http.Server
	logBuilder  strings.Builder
	folderEntry *widget.Entry
	reader      io.ReadCloser
	startButton *widget.Button
	stopButton  *widget.Button
	w           fyne.Window
)

// The main function creates a GUI for a web server with options to choose a folder, set a domain and port,
// start and stop the server, and view logs.
func GUI() {
	a := app.NewWithID("my-webserver")
	w = a.NewWindow("Mon Serveur Web")

	pref := a.Preferences()
	homeDir, _ = os.UserHomeDir()
	lastFolder = pref.StringWithFallback("lastFolder", homeDir)

	folderEntry = widget.NewEntry()
	folderEntry.SetText(lastFolder)

	folderButton := widget.NewButton("Choisir un dossier", func() {
		folderDialog := dialog.NewFolderOpen(func(dir fyne.ListableURI, err error) {
			if err == nil && dir != nil {
				folderEntry.SetText(dir.Path())
				pref.SetString("lastFolder", dir.Path())
				lastFolder = dir.Path()
			}
		}, w)
		folderDialog.Show()
	})

	domainEntry := widget.NewEntry()
	domainEntry.SetText("localhost")

	portEntry := widget.NewEntry()
	portEntry.SetText("8080")

	startButton = widget.NewButton("ON", func() {
		ServerStart(portEntry.Text)
		startButton.Disable()
		stopButton.Enable()
	})

	stopButton = widget.NewButton("OFF", func() {
		ServerStop()
		stopButton.Disable()
		startButton.Enable()
	})

	stopButton.Disable()

	logButton := widget.NewButton("Logs", func() {
		logWindow := a.NewWindow("Logs")
		logWindow.Resize(fyne.NewSize(600, 400))

		logReader := widget.NewMultiLineEntry()
		logReader.SetText("")
		logReader.Disable()

		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := reader.Read(buf)
				if err != nil {
					break
				}
				logReader.SetText(logReader.Text + string(buf[:n]))
			}
		}()
		logWindow.SetContent(container.NewScroll(logReader))
		logWindow.Show()
	})

	content := container.NewVBox(
		folderButton,
		folderEntry,
		portEntry,
		domainEntry,
		startButton,
		stopButton,
		logButton,
	)

	w.SetContent(content)
	w.Resize(fyne.Size{
		Width:  480,
		Height: 240,
	})
	w.CenterOnScreen()

	if desk, ok := a.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Show", func() {
				w.Show()
			}))
		desk.SetSystemTrayMenu(m)
	}

	w.SetCloseIntercept(func() {
		w.Hide()
	})

	w.ShowAndRun()

}

// The function starts a server on a specified port and serves files from a specified folder with
// logging middleware.
func ServerStart(port string) {
	if _, err := os.Stat(folderEntry.Text); os.IsNotExist(err) {
		dialog.ShowInformation("Erreur", "Le dossier n'existe pas ou ne peut pas être lu.", w)
		return
	}

	mux := http.NewServeMux()
	mux.Handle("/", loggingMiddleware(&dynamicFileServerHandler{}))
	server = &http.Server{Addr: ":" + port, Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("listen: %s\n", err)
		}
	}()
}

// The function stops the server and exits the program.
func ServerStop() {
	if server != nil {
		server.Close()
	}
	os.Exit(1)
}
