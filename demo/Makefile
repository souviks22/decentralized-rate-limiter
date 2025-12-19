# Makefile for running Vegeta benchmarks

RATE ?= 1000
DURATION ?= 30s
TARGETS ?= targets.txt
OUTPUT ?= results.bin
PLOT ?= plot.html

.PHONY: all attack report plot clean

all: attack report plot

attack:
	@echo "🔥 Running vegeta attack with rate=$(RATE), duration=$(DURATION)..."
	vegeta attack -rate=$(RATE) -duration=$(DURATION) -targets=$(TARGETS) | tee $(OUTPUT) > /dev/null

report:
	@echo "📊 Generating vegeta report..."
	vegeta report < $(OUTPUT)

plot:
	@echo "🖼️  Generating plot at $(PLOT)..."
	vegeta plot < $(OUTPUT) > $(PLOT)
	@echo "✅ Open the plot in browser with: open $(PLOT)"

clean:
	@echo "🧹 Cleaning results..."
	rm -f $(OUTPUT) $(PLOT)
