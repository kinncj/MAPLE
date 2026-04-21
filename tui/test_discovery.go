package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ─── Multi-framework test discovery ──────────────────────────────────────────

func loadTestEntries() []testEntry {
	var out []testEntry
	out = append(out, detectGherkinEntries()...)
	out = append(out, detectGoTestEntries()...)
	out = append(out, detectNodeTestEntries()...)
	out = append(out, detectPythonTestEntries()...)
	out = append(out, detectRubyTestEntries()...)
	out = append(out, detectJavaTestEntries()...)
	out = append(out, detectPHPTestEntries()...)
	out = append(out, detectRustTestEntries()...)
	return out
}

func detectGherkinEntries() []testEntry {
	globs := []string{"tests/features/*.feature", "features/*.feature", "test/features/*.feature"}
	seen := map[string]bool{}
	var out []testEntry
	for _, g := range globs {
		paths, _ := filepath.Glob(g)
		for _, p := range paths {
			if seen[p] {
				continue
			}
			seen[p] = true
			count := 0
			if data, err := os.ReadFile(p); err == nil {
				for _, l := range strings.Split(string(data), "\n") {
					t := strings.TrimSpace(l)
					if strings.HasPrefix(t, "Scenario:") || strings.HasPrefix(t, "Scenario Outline:") {
						count++
					}
				}
			}
			out = append(out, testEntry{
				path:      p,
				framework: "gherkin",
				runCmd:    gherkinRunCmd(p),
				count:     count,
			})
		}
	}
	return out
}

func gherkinRunCmd(path string) []string {
	if data, err := os.ReadFile("Makefile"); err == nil {
		if strings.Contains(string(data), "test-features") {
			return []string{"make", "test-features", "FEATURE=" + path}
		}
	}
	if _, err := exec.LookPath("npx"); err == nil {
		return []string{"npx", "cucumber-js", path}
	}
	if _, err := exec.LookPath("bundle"); err == nil {
		return []string{"bundle", "exec", "cucumber", path}
	}
	return []string{"cucumber", path}
}

func detectGoTestEntries() []testEntry {
	if _, err := os.Stat("go.mod"); err != nil {
		return nil
	}
	var paths []string
	_ = filepath.WalkDir(".", func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && (d.Name() == "vendor" || d.Name() == ".git") {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(p, "_test.go") {
			paths = append(paths, p)
		}
		return nil
	})
	var out []testEntry
	for _, p := range paths {
		dir := filepath.Dir(p)
		pkg := "./" + dir + "/..."
		if dir == "." {
			pkg = "./..."
		}
		out = append(out, testEntry{
			path:      p,
			framework: "go",
			runCmd:    []string{"go", "test", "-v", pkg},
		})
	}
	return out
}

func detectNodeTestEntries() []testEntry {
	if _, err := os.Stat("package.json"); err != nil {
		return nil
	}
	fw := detectNodeFramework()
	if fw == "" {
		return nil
	}
	globs := []string{
		"src/**/*.test.ts", "src/**/*.test.js",
		"src/**/*.spec.ts", "src/**/*.spec.js",
		"test/**/*.test.ts", "test/**/*.test.js",
		"tests/**/*.test.ts", "tests/**/*.test.js",
		"__tests__/**/*.ts", "__tests__/**/*.js",
		"*.test.ts", "*.test.js", "*.spec.ts", "*.spec.js",
	}
	seen := map[string]bool{}
	var out []testEntry
	for _, g := range globs {
		paths, _ := filepath.Glob(g)
		for _, p := range paths {
			if seen[p] || strings.Contains(p, "node_modules") {
				continue
			}
			seen[p] = true
			out = append(out, testEntry{
				path:      p,
				framework: fw,
				runCmd:    nodeRunCmd(fw, p),
			})
		}
	}
	return out
}

func detectNodeFramework() string {
	data, err := os.ReadFile("package.json")
	if err != nil {
		return ""
	}
	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return ""
	}
	hasDep := func(name string) bool {
		for _, key := range []string{"dependencies", "devDependencies"} {
			if deps, ok := pkg[key].(map[string]interface{}); ok {
				if _, ok := deps[name]; ok {
					return true
				}
			}
		}
		return false
	}
	for _, f := range []string{"vitest.config.ts", "vitest.config.js", "vitest.config.mts"} {
		if _, err := os.Stat(f); err == nil {
			return "vitest"
		}
	}
	if hasDep("vitest") {
		return "vitest"
	}
	if hasDep("jest") {
		return "jest"
	}
	if _, ok := pkg["jest"]; ok {
		return "jest"
	}
	if hasDep("mocha") {
		return "mocha"
	}
	if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
		if _, ok := scripts["test"]; ok {
			return "npm"
		}
	}
	return ""
}

func nodeRunCmd(fw, path string) []string {
	switch fw {
	case "vitest":
		return []string{"npx", "vitest", "run", path}
	case "jest":
		return []string{"npx", "jest", "--no-coverage", path}
	case "mocha":
		return []string{"npx", "mocha", path}
	default:
		return []string{"npm", "test", "--", path}
	}
}

func detectPythonTestEntries() []testEntry {
	hasPytest := false
	for _, f := range []string{"pytest.ini", "setup.cfg", "pyproject.toml"} {
		if _, err := os.Stat(f); err == nil {
			hasPytest = true
			break
		}
	}
	globs := []string{
		"test_*.py", "tests/test_*.py", "test/**/test_*.py",
		"**/test_*.py", "**/*_test.py",
	}
	seen := map[string]bool{}
	var out []testEntry
	for _, g := range globs {
		paths, _ := filepath.Glob(g)
		for _, p := range paths {
			if seen[p] || strings.Contains(p, ".venv") || strings.Contains(p, "node_modules") {
				continue
			}
			seen[p] = true
			fw := "pytest"
			if !hasPytest {
				fw = "unittest"
			}
			out = append(out, testEntry{
				path:      p,
				framework: fw,
				runCmd:    []string{"python", "-m", "pytest", "-v", p},
			})
		}
	}
	return out
}

func detectRubyTestEntries() []testEntry {
	if _, err := os.Stat("spec"); err != nil {
		return nil
	}
	globs := []string{"spec/**/*_spec.rb", "spec/*_spec.rb"}
	seen := map[string]bool{}
	var out []testEntry
	for _, g := range globs {
		paths, _ := filepath.Glob(g)
		for _, p := range paths {
			if seen[p] {
				continue
			}
			seen[p] = true
			out = append(out, testEntry{
				path:      p,
				framework: "rspec",
				runCmd:    []string{"bundle", "exec", "rspec", p},
			})
		}
	}
	return out
}

func detectJavaTestEntries() []testEntry {
	hasMaven := false
	hasGradle := false
	if _, err := os.Stat("pom.xml"); err == nil {
		hasMaven = true
	}
	for _, f := range []string{"build.gradle", "build.gradle.kts"} {
		if _, err := os.Stat(f); err == nil {
			hasGradle = true
		}
	}
	if !hasMaven && !hasGradle {
		return nil
	}
	var paths []string
	_ = filepath.WalkDir("src/test", func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(p, ".java") {
			base := filepath.Base(p)
			if strings.HasSuffix(base, "Test.java") || strings.HasSuffix(base, "Tests.java") ||
				strings.HasSuffix(base, "IT.java") || strings.HasPrefix(base, "Test") {
				paths = append(paths, p)
			}
		}
		return nil
	})
	fw := "maven"
	if hasGradle && !hasMaven {
		fw = "gradle"
	}
	var out []testEntry
	for _, p := range paths {
		className := javaClassName(p)
		var runCmd []string
		if fw == "gradle" {
			runCmd = []string{"./gradlew", "test", "--tests", className}
		} else {
			runCmd = []string{"mvn", "test", "-Dtest=" + className, "-pl", "."}
		}
		out = append(out, testEntry{path: p, framework: fw, runCmd: runCmd})
	}
	return out
}

func javaClassName(path string) string {
	idx := strings.Index(path, "src/test/java/")
	if idx < 0 {
		return strings.TrimSuffix(filepath.Base(path), ".java")
	}
	rel := path[idx+len("src/test/java/"):]
	return strings.ReplaceAll(strings.TrimSuffix(rel, ".java"), string(os.PathSeparator), ".")
}

func detectPHPTestEntries() []testEntry {
	found := false
	for _, f := range []string{"phpunit.xml", "phpunit.xml.dist"} {
		if _, err := os.Stat(f); err == nil {
			found = true
		}
	}
	if !found {
		return nil
	}
	runner := "./vendor/bin/phpunit"
	if _, err := os.Stat(runner); err != nil {
		runner = "phpunit"
	}
	globs := []string{"tests/**/*Test.php", "test/**/*Test.php", "tests/*Test.php"}
	seen := map[string]bool{}
	var out []testEntry
	for _, g := range globs {
		paths, _ := filepath.Glob(g)
		for _, p := range paths {
			if seen[p] {
				continue
			}
			seen[p] = true
			out = append(out, testEntry{
				path:      p,
				framework: "phpunit",
				runCmd:    []string{runner, p},
			})
		}
	}
	return out
}

func detectRustTestEntries() []testEntry {
	if _, err := os.Stat("Cargo.toml"); err != nil {
		return nil
	}
	return []testEntry{{
		path:      "Cargo.toml",
		framework: "cargo",
		runCmd:    []string{"cargo", "test"},
	}}
}
