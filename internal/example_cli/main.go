package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/osbuild/logging/pkg/strc"
)

func add(ctx context.Context, num int) int {
	span, _ := strc.Start(ctx, "add", "num", num)
	defer span.End()

	if num%2 == 0 {
		time.Sleep(time.Duration(num) * 2 * time.Millisecond)
	} else {
		time.Sleep(time.Duration(num) * time.Millisecond)
	}

	return num
}

func sum(ctx context.Context, nums ...int) int {
	span, ctx := strc.Start(ctx, "sum", "length", len(nums))
	defer span.End()

	sum := 0
	for _, num := range nums {
		sum += add(ctx, num)
	}
	return sum
}

func calculate(ctx context.Context) {
	span, ctx := strc.Start(ctx, "calculate")
	defer span.End()

	sum(ctx, 10, 20, 30)
	sum(ctx, 40, 50, 60)
}

func main() {
	buf := bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	strc.SetLogger(logger)
	calculate(context.Background())

	cwd, _ := os.Getwd()

	for _, f := range []string{"text", "flamegraph"} {
		fmt.Printf("Output format: %s\n", f)
		cmd := exec.Command("go", "run", "github.com/osbuild/logging/pkg/strc/stgraph", "-strip-source-prefix", cwd+"/", "-output-format", f)
		cmd.Stdin = bytes.NewReader(buf.Bytes())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			panic(err)
		}

		fmt.Println()
	}
}
