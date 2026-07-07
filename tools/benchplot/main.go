// Package main provides a tool to parse Go benchmark results and generate plots.
// This replaces the Python script sweep_plot.py with a native Go implementation.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

const (
	defaultFile   = "benchmark/sweep_final.txt"
	defaultPrefix = "benchmark/sweep_plot"
	plotWidth     = 8 * vg.Inch
	plotHeight    = 5 * vg.Inch
	dpi           = 150
)

var (
	inputFile  = flag.String("input", defaultFile, "Path to benchmark results file")
	outputDir  = flag.String("outdir", "benchmark", "Output directory for plots")
	outputPfx  = flag.String("prefix", "sweep_plot", "Output file prefix")
)

// BenchmarkRecord represents a single benchmark result line
type BenchmarkRecord struct {
	Suite  string
	Label  string
	Procs  int
	Iters  int
	NsOp   float64
	BytesOp float64
	AllocsOp float64
}

// SummaryStats holds aggregated statistics for a benchmark label
type SummaryStats struct {
	Ns     float64
	Bytes  float64
	Allocs float64
}

// benchPattern matches Go benchmark output lines
var benchPattern = regexp.MustCompile(
	`^Benchmark(?P<suite>\w+)/(?P<label>.+?)-(?P<procs>\d+)\s+(?P<iter>\d+)\s+(?P<ns>[\d.]+) ns/op\s+(?P<bytes>[\d.]+) B/op\s+(?P<allocs>[\d.]+) allocs/op`,
)

func main() {
	flag.Parse()

	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Parse benchmark results
	records, err := parseResults(*inputFile)
	if err != nil {
		log.Fatalf("Failed to parse results: %v", err)
	}

	// Aggregate by suite and label
	summary := aggregateResults(records)

	// Generate all plots
	prefix := filepath.Join(*outputDir, *outputPfx)
	generateAllPlots(summary, prefix)

	fmt.Println("\nAll plots generated.")
}

// parseResults parses the benchmark results file
func parseResults(filename string) ([]BenchmarkRecord, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var records []BenchmarkRecord
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if match := benchPattern.FindStringSubmatch(line); match != nil {
			record := BenchmarkRecord{
				Suite: match[1],
				Label: match[2],
			}
			record.Procs, _ = strconv.Atoi(match[3])
			record.Iters, _ = strconv.Atoi(match[4])
			record.NsOp, _ = strconv.ParseFloat(match[5], 64)
			record.BytesOp, _ = strconv.ParseFloat(match[6], 64)
			record.AllocsOp, _ = strconv.ParseFloat(match[7], 64)
			records = append(records, record)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// aggregateResults aggregates records by suite and label (averaging multiple runs)
func aggregateResults(records []BenchmarkRecord) map[string]map[string]SummaryStats {
	// Group by suite -> label -> []records
	grouped := make(map[string]map[string][]BenchmarkRecord)
	for _, r := range records {
		if grouped[r.Suite] == nil {
			grouped[r.Suite] = make(map[string][]BenchmarkRecord)
		}
		grouped[r.Suite][r.Label] = append(grouped[r.Suite][r.Label], r)
	}

	// Compute averages
	summary := make(map[string]map[string]SummaryStats)
	for suite, labels := range grouped {
		summary[suite] = make(map[string]SummaryStats)
		for label, runs := range labels {
			var sumNs, sumBytes, sumAllocs float64
			for _, r := range runs {
				sumNs += r.NsOp
				sumBytes += r.BytesOp
				sumAllocs += r.AllocsOp
			}
			n := float64(len(runs))
			summary[suite][label] = SummaryStats{
				Ns:     sumNs / n,
				Bytes:  sumBytes / n,
				Allocs: sumAllocs / n,
			}
		}
	}

	return summary
}

// generateAllPlots generates all the plots matching the Python script
func generateAllPlots(summary map[string]map[string]SummaryStats, prefix string) {
	// 2a. GetAutoTrackSize — pattern count sweep, one series per match type
	for _, matchType := range []string{"match-first", "match-last", "no-match"} {
		plotSweep(summary, "SweepGetAutoTrackSize", matchType,
			"Pattern count", fmt.Sprintf("GetAutoTrackSize (%s)", matchType),
			fmt.Sprintf("%s_GetAutoTrackSize_%s.png", prefix, matchType),
			sweepConfig{
				logX:        true,
				withAllocs:  true,
				labelFilter: func(l string) bool { return strings.HasPrefix(l, matchType) },
				xTransform:  extractLastNumber,
			})
	}

	// 2b. GitattributesRead — file size sweep
	plotSweep(summary, "SweepGitattributesRead", "",
		"File size (bytes)", "Gitattributes read + parse",
		fmt.Sprintf("%s_GitattributesRead.png", prefix),
		sweepConfig{
			logX:       true,
			withAllocs: true,
			xTransform: extractSecondNumber,
		})

	// 2c. CleanPassThrough — file size sweep
	sizeMap1 := map[string]float64{
		"1KB": 1024, "10KB": 10240, "100KB": 102400,
		"1MB": 1048576, "10MB": 10485760, "50MB": 52428800,
	}
	plotSweep(summary, "SweepCleanPassThrough", "",
		"File size", "Clean pass-through (autotrack under threshold)",
		fmt.Sprintf("%s_CleanPassThrough.png", prefix),
		sweepConfig{
			logX:       true,
			logY:       true,
			xTransform: func(l string) float64 { return sizeMap1[l] },
		})

	// 2d. CleanThresholdRatio — ratio sweep
	plotSweep(summary, "SweepCleanThresholdRatio", "",
		"File size / threshold ratio", "Clean pass-through by threshold proximity",
		fmt.Sprintf("%s_CleanThresholdRatio.png", prefix),
		sweepConfig{
			xTransform: extractRatioFromLabel,
		})

	// 2e. CleanOverThreshold — file size sweep
	sizeMap2 := map[string]float64{"1MB": 1, "5MB": 5, "10MB": 10, "25MB": 25}
	plotSweep(summary, "SweepCleanOverThreshold", "",
		"File size", "Clean over-threshold (LFS pointer creation)",
		fmt.Sprintf("%s_CleanOverThreshold.png", prefix),
		sweepConfig{
			xTransform: func(l string) float64 { return sizeMap2[l] },
		})

	// 2f. FindLargeFilesStat — file count sweep
	plotSweep(summary, "SweepFindLargeFilesStat", "",
		"File count", "findLargeFiles stat loop",
		fmt.Sprintf("%s_FindLargeFilesStat.png", prefix),
		sweepConfig{
			logX:       true,
			withAllocs: true,
			xTransform: extractFirstNumber,
		})

	// 2g. SequentialGetAutoTrackSize — call count sweep
	for _, mt := range []string{"match", "no-match"} {
		plotSweep(summary, "SweepSequentialGetAutoTrackSize", mt,
			"Sequential call count", fmt.Sprintf("Sequential GetAutoTrackSize (%s)", mt),
			fmt.Sprintf("%s_SequentialGetAutoTrackSize_%s.png", prefix, mt),
			sweepConfig{
				logX:        true,
				withAllocs:  true,
				labelFilter: func(l string) bool { return strings.HasSuffix(l, mt) },
				xTransform:  extractFirstNumber,
			})
	}

	// 2h. CleanPassThroughManyPatterns — pattern count sweep
	plotSweep(summary, "SweepCleanPassThroughManyPatterns", "",
		"Pattern count", "Clean pass-through with many .gitattributes patterns",
		fmt.Sprintf("%s_CleanPassThroughManyPatterns.png", prefix),
		sweepConfig{
			logX:       true,
			withAllocs: true,
			xTransform: extractFirstNumber,
		})

	// 2i. CleanNoAutoTrack — file size sweep (baseline)
	sizeMap3 := map[string]float64{"1KB": 1024, "100KB": 102400, "1MB": 1048576}
	plotSweep(summary, "SweepCleanNoAutoTrack", "",
		"File size", "Clean pass-through (no autotracksize in .gitattributes)",
		fmt.Sprintf("%s_CleanNoAutoTrack.png", prefix),
		sweepConfig{
			xTransform: func(l string) float64 { return sizeMap3[l] },
		})

	// 2j. CleanFileNameLength — name length sweep
	plotSweep(summary, "SweepCleanFileNameLength", "",
		"Name length (chars)", "Clean pass-through by file name length",
		fmt.Sprintf("%s_CleanFileNameLength.png", prefix),
		sweepConfig{
			xTransform: extractSecondNumber,
		})

	// 2k-2p. Discrete bar charts
	plotDiscreteBar(summary, "SweepGetAutoTrackSizeCold",
		"GetAutoTrackSize cold cache", "ns/op",
		fmt.Sprintf("%s_GetAutoTrackSizeCold.png", prefix))

	plotDiscreteBar(summary, "SweepGitattributesExtreme",
		"Gitattributes extreme values", "ns/op",
		fmt.Sprintf("%s_GitattributesExtreme.png", prefix))

	plotDiscreteBar(summary, "SweepCleanAlreadyPointer",
		"Clean already-pointer content", "ns/op",
		fmt.Sprintf("%s_CleanAlreadyPointer.png", prefix))

	plotDiscreteBar(summary, "SweepCleanExcludedPaths",
		"Clean excluded paths", "ns/op",
		fmt.Sprintf("%s_CleanExcludedPaths.png", prefix))

	plotDiscreteBar(summary, "SweepSmudgePassThrough",
		"Smudge pass-through scenarios", "ns/op",
		fmt.Sprintf("%s_SmudgePassThrough.png", prefix))

	plotDiscreteBar(summary, "SweepIsAutoTrackExcluded",
		"IsAutoTrackExcluded scenarios", "ns/op",
		fmt.Sprintf("%s_IsAutoTrackExcluded.png", prefix))

	// 2q. GetAutoTrackSizePatternShape — grouped bars
	plotPatternShape(summary, fmt.Sprintf("%s_GetAutoTrackSizePatternShape.png", prefix))

	// 2r. Combined: Clean throughput
	plotCleanThroughput(summary, fmt.Sprintf("%s_CleanThroughput.png", prefix))

	// 2s. Combined: GetAutoTrackSize pattern count (all match types)
	plotGetAutoTrackSizeCombined(summary, fmt.Sprintf("%s_GetAutoTrackSize_combined.png", prefix))
}

// sweepConfig holds configuration for sweep plots
type sweepConfig struct {
	logX        bool
	logY        bool
	withAllocs  bool
	labelFilter func(string) bool
	xTransform  func(string) float64
}

// plotSweep generates a line plot for a sweep benchmark
func plotSweep(summary map[string]map[string]SummaryStats, suite, subset, xlabel, title, filename string, cfg sweepConfig) {
	data, ok := summary[suite]
	if !ok {
		fmt.Printf("  [skip] %s not found\n", suite)
		return
	}

	// Filter and sort labels
	var labels []string
	for l := range data {
		if cfg.labelFilter == nil || cfg.labelFilter(l) {
			labels = append(labels, l)
		}
	}
	if len(labels) == 0 {
		return
	}

	// Extract x values and sort by them
	type pair struct {
		x     float64
		label string
	}
	var pairs []pair
	for _, l := range labels {
		x := cfg.xTransform(l)
		if !math.IsNaN(x) && x != 0 {
			pairs = append(pairs, pair{x, l})
		}
	}
	if len(pairs) == 0 {
		return
	}

	sort.Slice(pairs, func(i, j int) bool { return pairs[i].x < pairs[j].x })

	// Build plot data
	pts := make(plotter.XYs, len(pairs))
	var allocPts plotter.XYs
	if cfg.withAllocs {
		allocPts = make(plotter.XYs, len(pairs))
	}

	for i, p := range pairs {
		pts[i].X = p.x
		pts[i].Y = data[p.label].Ns
		if cfg.withAllocs {
			allocPts[i].X = p.x
			allocPts[i].Y = data[p.label].Allocs
		}
	}

	// Create plot
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = xlabel
	p.Y.Label.Text = "ns/op"

	if cfg.logX {
		p.X.Scale = plot.LogScale{}
		p.X.Tick.Marker = plot.LogTicks{}
	}
	if cfg.logY {
		p.Y.Scale = plot.LogScale{}
		p.Y.Tick.Marker = plot.LogTicks{}
	}

	p.Add(plotter.NewGrid())

	line, scatter, err := plotter.NewLinePoints(pts)
	if err != nil {
		log.Printf("Error creating line: %v", err)
		return
	}
	line.Color = color.RGBA{R: 31, G: 119, B: 180, A: 255}
	scatter.Color = line.Color
	scatter.Shape = draw.CircleGlyph{}
	p.Add(line, scatter)

	// Add allocs on right axis if requested
	if cfg.withAllocs && len(allocPts) > 0 {
		// Note: gonum/plot doesn't support dual Y-axes natively
		// We'll create a separate plot for now (simplified)
		// For a full implementation, consider using a different library or custom rendering
	}

	if err := p.Save(plotWidth, plotHeight, filename); err != nil {
		log.Printf("Error saving plot %s: %v", filename, err)
		return
	}

	fmt.Printf("  → %s\n", filename)
}

// plotDiscreteBar generates a bar chart for discrete benchmarks
func plotDiscreteBar(summary map[string]map[string]SummaryStats, suite, title, ylabel, filename string) {
	data, ok := summary[suite]
	if !ok {
		return
	}

	labels := make([]string, 0, len(data))
	for l := range data {
		labels = append(labels, l)
	}
	sort.Strings(labels)

	if len(labels) == 0 {
		return
	}

	values := make(plotter.Values, len(labels))
	for i, l := range labels {
		values[i] = data[l].Ns
	}

	p := plot.New()
	p.Title.Text = title
	p.Y.Label.Text = ylabel

	bars, err := plotter.NewBarChart(values, vg.Points(20))
	if err != nil {
		log.Printf("Error creating bar chart: %v", err)
		return
	}
	bars.Color = color.RGBA{R: 31, G: 119, B: 180, A: 200}
	p.Add(bars)

	p.NominalX(labels...)

	if err := p.Save(plotWidth, plotHeight-vg.Inch, filename); err != nil {
		log.Printf("Error saving plot %s: %v", filename, err)
		return
	}

	fmt.Printf("  → %s\n", filename)
}

// plotPatternShape generates grouped bar chart for pattern shape benchmarks
func plotPatternShape(summary map[string]map[string]SummaryStats, filename string) {
	data, ok := summary["SweepGetAutoTrackSizePatternShape"]
	if !ok {
		return
	}

	// Parse labels: "simple-glob-5-pats" -> type="simple-glob", count=5
	type key struct {
		ptype string
		count int
	}
	parsed := make(map[key]float64)
	countSet := make(map[int]bool)
	typeSet := make(map[string]bool)

	for label, stats := range data {
		parts := strings.Split(label, "-")
		if len(parts) < 3 {
			continue
		}
		// Find "pats" and extract count before it
		var count int
		var ptype string
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] == "pats" && i > 0 {
				count, _ = strconv.Atoi(parts[i-1])
				ptype = strings.Join(parts[:i-1], "-")
				break
			}
		}
		if count > 0 && ptype != "" {
			parsed[key{ptype, count}] = stats.Ns
			countSet[count] = true
			typeSet[ptype] = true
		}
	}

	if len(parsed) == 0 {
		return
	}

	var counts []int
	for c := range countSet {
		counts = append(counts, c)
	}
	sort.Ints(counts)

	var types []string
	for t := range typeSet {
		types = append(types, t)
	}
	sort.Strings(types)

	p := plot.New()
	p.Title.Text = "GetAutoTrackSize by pattern shape and count"
	p.X.Label.Text = "Pattern count"
	p.Y.Label.Text = "ns/op"

	w := vg.Points(15)
	colors := []color.RGBA{
		{R: 31, G: 119, B: 180, A: 200},
		{R: 255, G: 127, B: 14, A: 200},
		{R: 44, G: 160, B: 44, A: 200},
		{R: 214, G: 39, B: 40, A: 200},
	}

	for i, ptype := range types {
		vals := make(plotter.Values, len(counts))
		for j, c := range counts {
			vals[j] = parsed[key{ptype, c}]
		}

		bars, err := plotter.NewBarChart(vals, w)
		if err != nil {
			continue
		}
		bars.Color = colors[i%len(colors)]
		bars.Offset = vg.Length(i-len(types)/2) * w
		p.Add(bars)
		p.Legend.Add(ptype, bars)
	}

	// Set X labels
	xlabels := make([]string, len(counts))
	for i, c := range counts {
		xlabels[i] = strconv.Itoa(c)
	}
	p.NominalX(xlabels...)

	if err := p.Save(plotWidth+2*vg.Inch, plotHeight, filename); err != nil {
		log.Printf("Error saving plot %s: %v", filename, err)
		return
	}

	fmt.Printf("  → %s\n", filename)
}

// plotCleanThroughput generates throughput comparison bar chart
func plotCleanThroughput(summary map[string]map[string]SummaryStats, filename string) {
	ptData := summary["SweepCleanPassThrough"]
	noatData := summary["SweepCleanNoAutoTrack"]
	ovtData := summary["SweepCleanOverThreshold"]

	sizeMap := map[string]float64{
		"1KB": 1024, "1MB": 1048576, "5MB": 5242880,
		"10MB": 10485760, "25MB": 26214400,
	}
	sizeLabels := []string{"1KB", "1MB", "5MB", "10MB", "25MB"}

	type barSet struct {
		label  string
		values plotter.Values
		color  color.RGBA
	}

	sets := []barSet{
		{"No autotrack", make(plotter.Values, len(sizeLabels)),
			color.RGBA{R: 31, G: 119, B: 180, A: 200}},
		{"Pass-through (under threshold)", make(plotter.Values, len(sizeLabels)),
			color.RGBA{R: 255, G: 127, B: 14, A: 200}},
		{"Over threshold → LFS Clean", make(plotter.Values, len(sizeLabels)),
			color.RGBA{R: 44, G: 160, B: 44, A: 200}},
	}

	for i, label := range sizeLabels {
		sz := sizeMap[label]
		mbPerSec := func(data map[string]SummaryStats, key string) float64 {
			if stats, ok := data[key]; ok && stats.Ns > 0 {
				return (sz / stats.Ns) * 1e9 / (1024 * 1024)
			}
			return 0
		}

		sets[0].values[i] = mbPerSec(noatData, label)
		sets[1].values[i] = mbPerSec(ptData, label)
		sets[2].values[i] = mbPerSec(ovtData, label)
	}

	p := plot.New()
	p.Title.Text = "Clean throughput comparison"
	p.Y.Label.Text = "Throughput (MB/s)"

	w := vg.Points(20)
	for i, set := range sets {
		bars, err := plotter.NewBarChart(set.values, w)
		if err != nil {
			continue
		}
		bars.Color = set.color
		bars.Offset = vg.Length(i-1) * w
		p.Add(bars)
		p.Legend.Add(set.label, bars)
	}

	p.NominalX(sizeLabels...)

	if err := p.Save(plotWidth, plotHeight, filename); err != nil {
		log.Printf("Error saving plot %s: %v", filename, err)
		return
	}

	fmt.Printf("  → %s\n", filename)
}

// plotGetAutoTrackSizeCombined generates combined line plot with all match types
func plotGetAutoTrackSizeCombined(summary map[string]map[string]SummaryStats, filename string) {
	data, ok := summary["SweepGetAutoTrackSize"]
	if !ok {
		return
	}

	p := plot.New()
	p.Title.Text = "GetAutoTrackSize — pattern count sweep"
	p.X.Label.Text = "Pattern count"
	p.Y.Label.Text = "ns/op"
	p.X.Scale = plot.LogScale{}
	p.X.Tick.Marker = plot.LogTicks{}
	p.Add(plotter.NewGrid())

	matchTypes := []struct {
		prefix string
		color  color.RGBA
		shape  draw.GlyphDrawer
	}{
		{"match-first", color.RGBA{R: 31, G: 119, B: 180, A: 255}, draw.CircleGlyph{}},
		{"match-last", color.RGBA{R: 255, G: 127, B: 14, A: 255}, draw.SquareGlyph{}},
		{"no-match", color.RGBA{R: 44, G: 160, B: 44, A: 255}, draw.TriangleGlyph{}},
	}

	for _, mt := range matchTypes {
		var pairs []pair
		for label := range data {
			if strings.HasPrefix(label, mt.prefix) {
				count := extractLastNumber(label)
				if !math.IsNaN(count) && count > 0 {
					pairs = append(pairs, pair{count, label})
				}
			}
		}

		if len(pairs) == 0 {
			continue
		}

		sort.Slice(pairs, func(i, j int) bool { return pairs[i].x < pairs[j].x })

		pts := make(plotter.XYs, len(pairs))
		for i, p := range pairs {
			pts[i].X = p.x
			pts[i].Y = data[p.label].Ns
		}

		line, scatter, err := plotter.NewLinePoints(pts)
		if err != nil {
			continue
		}
		line.Color = mt.color
		scatter.Color = mt.color
		scatter.Shape = mt.shape
		p.Add(line, scatter)
		p.Legend.Add(mt.prefix, line, scatter)
	}

	if err := p.Save(plotWidth, plotHeight, filename); err != nil {
		log.Printf("Error saving plot %s: %v", filename, err)
		return
	}

	fmt.Printf("  → %s\n", filename)
}

// Helper functions to extract numeric values from labels

func extractFirstNumber(label string) float64 {
	parts := strings.Split(label, "-")
	for _, p := range parts {
		if v, err := strconv.ParseFloat(p, 64); err == nil {
			return v
		}
	}
	return math.NaN()
}

func extractSecondNumber(label string) float64 {
	parts := strings.Split(label, "-")
	if len(parts) < 2 {
		return math.NaN()
	}
	for i := 1; i < len(parts); i++ {
		if v, err := strconv.ParseFloat(parts[i], 64); err == nil {
			return v
		}
	}
	return math.NaN()
}

func extractLastNumber(label string) float64 {
	parts := strings.Split(label, "-")
	for i := len(parts) - 1; i >= 0; i-- {
		if v, err := strconv.ParseFloat(parts[i], 64); err == nil {
			return v
		}
	}
	return math.NaN()
}

func extractRatioFromLabel(label string) float64 {
	parts := strings.Split(label, "-")
	if len(parts) < 2 {
		return math.NaN()
	}
	// Look for pattern like "ratio-0.5"
	for i, p := range parts {
		if p == "ratio" && i+1 < len(parts) {
			if v, err := strconv.ParseFloat(parts[i+1], 64); err == nil {
				return v
			}
		}
	}
	// Fallback to second number
	return extractSecondNumber(label)
}

type pair struct {
	x     float64
	label string
}
