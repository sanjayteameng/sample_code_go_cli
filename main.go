package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

type VersionInfo struct {
	Hostname  string `json:"hostname"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	Uptime    string `json:"uptime"`
}

type InterfaceInfo struct {
	Name      string   `json:"name"`
	State     string   `json:"state"`
	MTU       int      `json:"mtu"`
	MAC       string   `json:"mac"`
	Addresses []string `json:"addresses"`
}

func main() {
	httpAddr := getenv("HTTP_ADDR", ":8081")
	telnetAddr := getenv("TELNET_ADDR", ":2324")

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/version", handleVersion)
		mux.HandleFunc("/api/interfaces", handleInterfaces)
		mux.HandleFunc("/", handleRoot)

		log.Printf("HTTP server listening on %s", httpAddr)
		if err := http.ListenAndServe(httpAddr, mux); err != nil {
			log.Fatalf("http server error: %v", err)
		}
	}()

	go func() {
		log.Printf("Telnet server listening on %s", telnetAddr)
		if err := serveTelnet(telnetAddr); err != nil {
			log.Fatalf("telnet server error: %v", err)
		}
	}()

	runLocalCLI(os.Stdin, os.Stdout)
}

func handleRoot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "sample_go_code is running")
	fmt.Fprintln(w, "Available endpoints:")
	fmt.Fprintln(w, "  /api/version")
	fmt.Fprintln(w, "  /api/interfaces")
}

func handleVersion(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, versionInfo())
}

func handleInterfaces(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, interfaceInfo())
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func serveTelnet(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		go handleTelnet(conn)
	}
}

func handleTelnet(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	fmt.Fprint(conn, "Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	fmt.Fprint(conn, "Password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	if strings.TrimSpace(username) != "admin" || strings.TrimSpace(password) != "admin" {
		fmt.Fprint(conn, "\r\nAuthentication failed.\r\n")
		return
	}

	fmt.Fprint(conn, "\r\nWelcome to sample_go_code over Telnet\r\n")
	runSession(reader, conn, "TE-Telnet =>> ", "\r\n")
}

func runLocalCLI(in io.Reader, out io.Writer) {
	fmt.Fprintln(out, "sample_go_code local CLI")
	runSession(bufio.NewReader(in), out, "TE-CLI> ", "\n")
}

func runSession(reader *bufio.Reader, out io.Writer, prompt, newline string) {
	fmt.Fprintf(out, "Type 'help' for commands.%s", newline)
	for {
		fmt.Fprint(out, prompt)
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		response, exit := handleCommand(line, newline)
		if response != "" {
			fmt.Fprint(out, response)
		}
		if exit {
			return
		}
	}
}

func handleCommand(line, newline string) (string, bool) {
	switch strings.TrimSpace(strings.ToLower(line)) {
	case "help":
		return strings.Join([]string{
			"Commands:",
			"  help",
			"  show version",
			"  show interface",
			"  exit",
			"",
		}, newline), false
	case "show version":
		return formatVersion(versionInfo(), newline), false
	case "show interface":
		return formatInterfaces(interfaceInfo(), newline), false
	case "exit", "quit":
		return "Bye." + newline, true
	case "":
		return "", false
	default:
		return "% Unknown command" + newline, false
	}
}

func versionInfo() VersionInfo {
	hostname, _ := os.Hostname()
	return VersionInfo{
		Hostname:  hostname,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Uptime:    readUptime(),
	}
}

func interfaceInfo() []InterfaceInfo {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil
	}

	items := make([]InterfaceInfo, 0, len(ifs))
	for _, ifc := range ifs {
		addrs, _ := ifc.Addrs()
		addrStrings := make([]string, 0, len(addrs))
		for _, addr := range addrs {
			addrStrings = append(addrStrings, addr.String())
		}
		state := "down"
		if ifc.Flags&net.FlagUp != 0 {
			state = "up"
		}
		items = append(items, InterfaceInfo{
			Name:      ifc.Name,
			State:     state,
			MTU:       ifc.MTU,
			MAC:       ifc.HardwareAddr.String(),
			Addresses: addrStrings,
		})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items
}

func formatVersion(info VersionInfo, newline string) string {
	lines := []string{
		fmt.Sprintf("Hostname   : %s", info.Hostname),
		fmt.Sprintf("Go Version : %s", info.GoVersion),
		fmt.Sprintf("OS         : %s", info.OS),
		fmt.Sprintf("Arch       : %s", info.Arch),
		fmt.Sprintf("Uptime     : %s", info.Uptime),
		"",
	}
	return strings.Join(lines, newline)
}

func formatInterfaces(items []InterfaceInfo, newline string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-12s %-6s %-6s %-18s %s%s", "Name", "State", "MTU", "MAC", "Addresses", newline)
	fmt.Fprintf(&b, "%-12s %-6s %-6s %-18s %s%s", strings.Repeat("-", 12), strings.Repeat("-", 6), strings.Repeat("-", 6), strings.Repeat("-", 18), strings.Repeat("-", 24), newline)
	for _, item := range items {
		fmt.Fprintf(&b, "%-12s %-6s %-6d %-18s %s%s", item.Name, item.State, item.MTU, item.MAC, strings.Join(item.Addresses, ", "), newline)
	}
	return b.String()
}

func readUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "unknown"
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "unknown"
	}
	secondsText := fields[0]
	whole, _, _ := strings.Cut(secondsText, ".")
	seconds, err := time.ParseDuration(whole + "s")
	if err != nil {
		return secondsText + " seconds"
	}
	return seconds.String()
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
