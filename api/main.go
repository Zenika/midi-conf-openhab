package main

import (
    "fmt"
    "flag"
    "github.com/faiface/beep/mp3"
    "github.com/faiface/beep/speaker"
    "os"
    "time"
    "math/rand"
    "github.com/faiface/beep"
    "net/http"
    "log"
    "io/ioutil"
)

func init() {
    rand.Seed(time.Now().UTC().UnixNano())
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "image/x-icon")
    w.Header().Set("Cache-Control", "public, max-age=7776000")
    fmt.Fprintln(w, "data:image/x-icon;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQEAYAAABPYyMiAAAABmJLR0T///////8JWPfcAAAACXBIWXMAAABIAAAASABGyWs+AAAAF0lEQVRIx2NgGAWjYBSMglEwCkbBSAcACBAAAeaR9cIAAAAASUVORK5CYII=\n")
}

func soundHandler(dir string) func(http.ResponseWriter,*http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Println("Sound play asked")
        path, err := choose(dir)
        if err != nil {
            log.Println("Retrieve sound failed: ", err)
            w.WriteHeader(http.StatusInternalServerError)
            w.Write([]byte("Play Sound failed :-("))
            return
        }
        stream, format, err := sound(path)
        if err != nil {
            log.Println("Stream sound failed: ", err)
            w.WriteHeader(http.StatusInternalServerError)
            w.Write([]byte("Play Sound failed :-("))
            return
        }
        err = play(stream, format)
        if err != nil {
            log.Println("Play sound failed: ", err)
            w.WriteHeader(http.StatusInternalServerError)
            w.Write([]byte("Play Sound failed :-("))
            return
        }
        log.Println("Sound played")
    }
}


func defaultHandler(w http.ResponseWriter, r *http.Request) {
    log.Printf("URI:%s PATH:%s\n", r.RequestURI, r.URL.Path[1:])
    fmt.Fprintf(w, "There's Nothing here :-), go to /sound", )
}

func main() {
    port := flag.Int("p", 8080, "Specifies the listen's port")
    dir := flag.String("s", "./sounds/", "Specifies the sound's dir")
    flag.Parse()

    log.Println("Sound directory is", *dir)
    log.Println("Listen on port ", *port)

    http.HandleFunc("/favicon.ico", faviconHandler)
    http.HandleFunc("/sound", soundHandler(*dir))
    http.HandleFunc("/", defaultHandler)
    http.ListenAndServe(fmt.Sprintf(":%v",*port), nil)
}

func sound(path string) (s beep.StreamSeekCloser, format beep.Format, err error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, beep.Format{}, fmt.Errorf("Error in path %s: ", path, err)
    }
    s, format, err = mp3.Decode(f)
    if err != nil {
        return nil, beep.Format{}, fmt.Errorf("Decode sound %s failed: %v", path, err)
    }
    return s, format, nil
}

func play(s beep.StreamSeekCloser, format beep.Format) error {
    err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
    if err != nil {
        return fmt.Errorf("Speaker initialisation failed: %v", err)
    }

    done := make(chan struct{})

    speaker.Play(beep.Seq(s, beep.Callback(func() {
        close(done)
    })))

    <-done
    return nil
}

func choose(dir string) (string, error) {
    files, err := ioutil.ReadDir(dir)
    if err != nil {
        return "", fmt.Errorf("Read dir %s failed: %v", dir, err)
    }
    index := rand.Intn(len(files))
    return fmt.Sprintf("%s/%s", dir, files[index].Name()), nil
}