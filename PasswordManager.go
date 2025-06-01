// ----------------------------------------------------------------------
// main.go
// Author: Zarak Khan
//
// A simple command‑line password manager. Stores credentials in a map
// keyed by website, each value being a slice of Entry structs. The program
// supports:
//   • Optional initialization from a whitespace‑separated file (site user pass)
//   • Listing (L) – display all stored credentials
//   • Adding  (A) – add a new (site, user, pass) triple, rejecting duplicates
//   • Removing (R) – delete a whole site (single user) or a specific user
//   • Exit     (X)
//
// Prompts, error messages, and output format match the assignment’s sample
// The data structure is fixed as map[string]EntrySlice.
// ----------------------------------------------------------------------

package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

// Entry represents one credential record.
type Entry struct {
    site, user, password string
}

// EntrySlice is a helper alias for slices of Entry.
type EntrySlice []Entry

// passwordMap maps a website → its stored credentials.
var passwordMap map[string]EntrySlice

// addEntry inserts a credential if (site,user) is not already present.
// Returns true on success, false if duplicate.
func addEntry(site, user, pass string, reportDup bool) bool {
    slice := passwordMap[site]
    for _, e := range slice {
        if e.user == user {
            if reportDup {
                fmt.Println("**Error: Attempting to add a duplicate entry. Try again.")
            }
            return false
        }
    }
    passwordMap[site] = append(slice, Entry{site: site, user: user, password: pass})
    return true
}

// listAll prints the entire password map.
func listAll() {
    for site, slice := range passwordMap {
        fmt.Printf("Website: %s\n", site)
        for _, e := range slice {
            fmt.Printf("\t %s \t %s\n", e.user, e.password)
        }
        fmt.Println()
    }
}

// removeEntry handles R‑command logic according to spec.
func removeEntry(line string) {
    fields := strings.Fields(line)
    if len(fields) == 0 {
        return
    }

    site := fields[0]
    slice, ok := passwordMap[site]
    if !ok {
        fmt.Println("**Error: Attempt to remove a website that does not exist in the map. Try again.")
        return
    }

    // Only website provided
    if len(fields) == 1 {
        if len(slice) > 1 {
            fmt.Println("**Error: Attempt to remove multiple users. Try again.")
            return
        }
        delete(passwordMap, site)
        return
    }

    // Website + username provided
    username := fields[1]
    idx := -1
    for i, e := range slice {
        if e.user == username {
            idx = i
            break
        }
    }
    if idx == -1 {
        fmt.Println("**Error: Attempt to remove a username that does not exist in the map. Try again.")
        return
    }
    slice = append(slice[:idx], slice[idx+1:]...)
    if len(slice) == 0 {
        delete(passwordMap, site)
    } else {
        passwordMap[site] = slice
    }
}

// readFile initializes the map from a given file path.
func readFile(path string) {
    fmt.Println("Initializing map using file...")
    f, err := os.Open(path)
    if err != nil {
        fmt.Println("**Error opening file. Exiting program...")
        os.Exit(1)
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        parts := strings.Fields(scanner.Text())
        if len(parts) == 3 {
            addEntry(parts[0], parts[1], parts[2], false)
        }
    }
    fmt.Println("Done reading in file.")
}

// printMenu shows the main command menu.
func printMenu() {
    fmt.Println()
    fmt.Println("Select a menu option: ")
    fmt.Println("\t L to list the contents of the map")
    fmt.Println("\t A to add a new entry to the map")
    fmt.Println("\t R to remove a website and/or user")
    fmt.Println(" or X to exit the program.")
    fmt.Print("Your choice --> ")
}

// ----------------------------------------------------------------------
// main starts the interactive loop.
func main() {
    passwordMap = make(map[string]EntrySlice)
    reader := bufio.NewReader(os.Stdin)

    // Optional file initialization.
    fmt.Print("Enter a filename if you would like to initialize the map using a file\n")
    fmt.Print("(or enter N/A if the map should start as empty): ")
    firstLine, _ := reader.ReadString('\n')
    firstLine = strings.TrimSpace(firstLine)
    // ensure next output starts on a new line (matches sample I/O)
    fmt.Println()
    if strings.ToUpper(firstLine) != "N/A" && firstLine != "" {
        readFile(firstLine)
    }

    // Command loop.
    for {
        printMenu()
        cmdLine, _ := reader.ReadString('\n')
        cmd := strings.TrimSpace(cmdLine)

        switch cmd {
        case "L":
            listAll()
        case "A":
            fmt.Print("Enter the site, username, and password (separated by spaces): ")
            entryLine, _ := reader.ReadString('\n')
            parts := strings.Fields(entryLine)
            if len(parts) == 3 {
                addEntry(parts[0], parts[1], parts[2], true)
            }
        case "R":
            fmt.Print("Enter the site and username (separated by spaces, username optional): ")
            remLine, _ := reader.ReadString('\n')
            removeEntry(remLine)
        case "X":
            fmt.Println("Exiting program.")
            return
        default:
            fmt.Println("**Error, unknown command. Try again.")
        }
    }
}
