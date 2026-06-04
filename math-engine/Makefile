CXX ?= g++
CXXFLAGS ?= -std=c++17 -O2 -Wall -Wextra -Wpedantic
LDFLAGS ?=
OUT ?= ../bin/logarift-math-engine

.PHONY: all clean test

all: $(OUT)

$(OUT): src/main.cpp
	mkdir -p $(dir $(OUT))
	$(CXX) $(CXXFLAGS) $(LDFLAGS) -o $(OUT) src/main.cpp

clean:
	rm -f $(OUT)

test: $(OUT)
	$(OUT) < samples/scoring-request.sample.json >/tmp/logarift-math-engine-output.json
	grep '"cla"' /tmp/logarift-math-engine-output.json
