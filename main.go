// bountyhunter/main.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	toolsList = []struct {
		name      string
		installer string
	}{
		{"subfinder", "go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest"},
		{"httpx", "go install -v github.com/projectdiscovery/httpx/cmd/httpx@latest"},
		{"katana", "go install -v github.com/projectdiscovery/katana/cmd/katana@latest"},
		{"nuclei", "go install -v github.com/projectdiscovery/nuclei/v2/cmd/nuclei@latest"},
		{"gf", "go install -v github.com/tomnomnom/gf@latest"},
		{"dirsearch", "pip3 install dirsearch"},
		{"bxss", "go install -v github.com/PortSwigger/bxss@latest"},
		{"subzy", "go install -v github.com/LukaSikic/subzy@latest"},
		{"corsy", "pip3 install corsy"},
		{"openredirex", "git clone https://github.com/DevShaft/OpenRedireX"},
	}
	
	toolsDir  = filepath.Join(os.Getenv("HOME"), ".bounty_tools")
	outputDir = filepath.Join(os.Getenv("HOME"), "bounty_output")
)

type Scanner struct {
	Domain      string
	ToolsPath   string
	Output      string
	StartTime   time.Time
	SuccessCnt  int
	FailedCnt   int
	WebhookURL  string
	Parallel    bool
}

func main() {
	var webhook string
	var parallel bool
	
	rootCmd := &cobra.Command{
		Use:   "bountyhunter <domain>",
		Short: "Automated Bug Bounty Scanner",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				color.Red("✖ Please provide a domain")
				os.Exit(1)
			}
			
			s := &Scanner{
				Domain:     args[0],
				ToolsPath:  toolsDir,
				Output:     filepath.Join(outputDir, fmt.Sprintf("%s_%s", args[0], time.Now().Format("2006-01-02"))),
				StartTime:  time.Now(),
				WebhookURL: webhook,
				Parallel:   parallel,
			}
			
			s.Init()
			s.RunScan()
			s.ShowSummary()
			s.Cleanup()
		},
	}

	rootCmd.Flags().StringVarP(&webhook, "webhook", "w", "", "Webhook URL for notifications")
	rootCmd.Flags().BoolVarP(&parallel, "parallel", "p", false, "Enable parallel scanning")

	if err := rootCmd.Execute(); err != nil {
		color.Red("✖ Error: %v", err)
		os.Exit(1)
	}
}

func (s *Scanner) Init() {
	s.checkUpdates()
	s.createDirs()
	s.installTools()
	s.setupGFPatterns()
	s.addToPath()
}

func (s *Scanner) setupGFPatterns() {
	cmd := exec.Command("gf", "-install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func (s *Scanner) createDirs() {
	os.MkdirAll(s.ToolsPath, 0755)
	os.MkdirAll(s.Output, 0755)
}

// Full scan steps implementation
func (s *Scanner) RunScan() {
	color.Cyan("\n🚀 Starting Full Scan for %s", s.Domain)
	
	steps := []func(){
		s.subdomainEnum,
		s.aliveCheck,
		s.urlCollection,
		s.sensitiveFiles,
		s.jsAnalysis,
		s.dirsearchScan,
		s.xssScan,
		s.takeoverCheck,
		s.corsScan,
		s.nucleiFullScan,
		s.lfiScan,
		s.redirectScan,
	}
	
	for i, step := range steps {
		color.Blue("\n🔹 Step %d/%d", i+1, len(steps))
		step()
	}
}

func (s *Scanner) subdomainEnum() {
	cmd := exec.Command("subfinder", 
		"-d", s.Domain, 
		"-all", 
		"-recursive",
		"-o", filepath.Join(s.Output, "subs.txt"))
	s.runCmd(cmd, "Subfinder")
}

func (s *Scanner) aliveCheck() {
	cmd := exec.Command("httpx",
		"-l", filepath.Join(s.Output, "subs.txt"),
		"-ports", "80,443,8000,8008,8888",
		"-threads", "200",
		"-o", filepath.Join(s.Output, "alive.txt"))
	s.runCmd(cmd, "HTTPX")
}

// Additional scan steps implementations omitted for space
// Full code would include all steps from the workflow

func (s *Scanner) runCmd(cmd *exec.Cmd, tool string) {
	color.Yellow("\n▶ Running %s...", tool)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		color.Red("✖ %s failed: %v", tool, err)
		s.FailedCnt++
		return
	}
	s.SuccessCnt++
}

func (s *Scanner) ShowSummary() {
	duration := time.Since(s.StartTime).Round(time.Second)
	
	summary := fmt.Sprintf(`
📊 Scan Summary for %s
---------------------------------
⏱  Duration:       %s
✅ Successful:     %d
❌ Failed:         %d

📂 Output Directory: %s
	`, s.Domain, duration, s.SuccessCnt, s.FailedCnt, s.Output)
	
	color.Cyan(summary)
	
	if s.WebhookURL != "" {
		s.sendWebhook()
	}
}

func (s *Scanner) sendWebhook() {
	// Webhook implementation
}
