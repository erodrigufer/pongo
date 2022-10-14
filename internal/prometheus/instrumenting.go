package prometheus

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// counter, holds all the data required to handle a Prometheus counter.
type counter struct {
	// name, of counter.
	name string
	// description, of counter. It is equivalent to the 'Help' field of
	// prometheus.CounterOpts.
	description string
	// labelsNames, names of the labels used by a particular counter.
	labelsNames []string
	// promObject, is the actual Prometheus object that hosts a counter
	// with labels.
	promObject *prometheus.CounterVec
}

// counters, holds all the counters registered for a particular instrumentation.
type counters struct {
	counters map[string]*counter
}

// newCounter, returns a new counter pointer.
func newCounter(name, description string, labels []string) *counter {
	c := new(counter)
	c.name = name
	c.description = description
	c.labelsNames = labels

	c.promObject = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: c.name,
			Help: c.description,
		},
		c.labelsNames,
	)
	return c
}

// insertCounter, adds counter to counters map.
func (cs counters) insertCounter(c *counter) error {

	if cs.counterExists(c.name) {
		return fmt.Errorf("a counter could not be inserted to the counters map, since another counter already exists with the same name (name='%s')", c.name)
	}

	// Add counter to map.
	cs.counters[c.name] = c

	return nil
}

// registerCounter, register a counter with the default Prometheus register.
// If the counter does not exist within the counters map this function returns
// an error.
func (cs counters) registerCounter(name string) error {
	if !cs.counterExists(name) {
		return fmt.Errorf("could not register counter %s, since it does not exist within the counters map.", name)
	}

	c := cs.counters[name]
	// Register the counter.
	prometheus.Register(c.promObject)

	return nil
}

// counterExists, checks if a counter with a given `name` exists within the
// counters map of the counters struct. It returns false if the counter does
// not exist.
func (cs counters) counterExists(name string) bool {
	_, ok := cs.counters[name]
	return ok
}

// incrementCounter, interface-specific method used to increment a counter.
// If the counter does not exist within the counters map this method returns
// an error.
func (cs counters) incrementCounter(name string, labels []string) error {
	if !cs.counterExists(name) {
		return fmt.Errorf("could not increment counter, since the counter does not exist within counters map")
	}

	coun := cs.counters[name]

	// Increment counter using the labels passed as parameters.
	// For more information about 'exploding' the labels []string slice, read
	// the following article on variadic functions:
	// https://www.digitalocean.com/community/tutorials/how-to-use-variadic-functions-in-go
	coun.promObject.WithLabelValues(labels...).Inc()

	return nil

}

// IncrementCounter, is the exported function responsible for incrementing a
// counter defined in an InstrumentationAPI interface. If the InstrumentationAPI
// method noOp() returns true, then no instrumentation is performed by
// this function.
func IncrementCounter(i InstrumentationAPI, name string, labels ...string) error {
	// Do not perform any instrumentation if the noOp method of the interface
	// returns true.
	if i.noOp() {
		return nil
	}

	if err := i.incrementCounter(name, labels); err != nil {
		return err
	}

	return nil
}

// gauge, holds all the data required to handle a Prometheus gauge.
type gauge struct {
	// name, of gauge.
	name string
	// description, of gauge. It is equivalent to the 'Help' field of
	// prometheus.GaugeOpts.
	description string
	// labelsNames, names of the labels used by a particular gauge.
	labelsNames []string
	// promObject, is the actual Prometheus object that hosts a gauge
	// with labels.
	promObject *prometheus.GaugeVec
}

// gauges, holds all the gauges registered for a particular instrumentation.
type gauges struct {
	gauges map[string]*gauge
}

// newGauge, returns a new gauge pointer.
func newGauge(name, description string, labels []string) *gauge {
	g := new(gauge)
	g.name = name
	g.description = description
	g.labelsNames = labels

	g.promObject = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: g.name,
			Help: g.description,
		},
		g.labelsNames,
	)
	return g
}

// gaugeExists, checks if a gauge with a given `name` exists within the
// gauges map of the gauges struct. It returns false if the gauge does
// not exist.
func (gs gauges) gaugeExists(name string) bool {
	_, ok := gs.gauges[name]
	return ok
}

// registerGauge, register a gauge with the default Prometheus register.
// If the gauge does not exist within the gauges map this function returns
// an error.
func (gs gauges) registerGauge(name string) error {
	if !gs.gaugeExists(name) {
		return fmt.Errorf("could not register gauge %s, since it does not exist within the gauges map.", name)
	}

	g := gs.gauges[name]
	// Register the gauge.
	prometheus.Register(g.promObject)

	return nil
}

// insertGauge, adds gauge to gauges map.
func (gs gauges) insertGauge(g *gauge) error {

	if gs.gaugeExists(g.name) {
		return fmt.Errorf("a gauge could not be inserted to the gauges map, since another gauge already exists with the same name (name='%s')", g.name)
	}

	// Add gauge to map.
	gs.gauges[g.name] = g

	return nil
}

// incrementGauge, method used to increment the gauge's value by 1.
// If the gauge does not exist within the gauges map this method returns
// an error.
func (gs gauges) incrementGauge(name string, labels []string) error {
	if !gs.gaugeExists(name) {
		return fmt.Errorf("could not increase gauge, since the gauge does not exist within gauges map")
	}

	gaug := gs.gauges[name]

	// Increment gauge using the labels passed as parameters.
	// For more information about 'exploding' the labels []string slice, read
	// the following article on variadic functions:
	// https://www.digitalocean.com/community/tutorials/how-to-use-variadic-functions-in-go
	gaug.promObject.WithLabelValues(labels...).Inc()

	return nil

}

// decreaseGauge, method used to decrease the gauge's value by 1.
// If the gauge does not exist within the gauges map this method returns
// an error.
func (gs gauges) decreaseGauge(name string, labels []string) error {
	if !gs.gaugeExists(name) {
		return fmt.Errorf("could not decrease gauge, since the gauge does not exist within gauges map")
	}

	gaug := gs.gauges[name]

	// Decrease gauge using the labels passed as parameters.
	// For more information about 'exploding' the labels []string slice, read
	// the following article on variadic functions:
	// https://www.digitalocean.com/community/tutorials/how-to-use-variadic-functions-in-go
	gaug.promObject.WithLabelValues(labels...).Dec()

	return nil

}

// IncrementGauge, is the exported function responsible for incrementing a
// gauge by 1 using an InstrumentationAPI interface. If the
// InstrumentationAPI method noOp() returns true, then no instrumentation is
// performed by this function.
func IncrementGauge(i InstrumentationAPI, name string, labels ...string) error {
	// Do not perform any instrumentation if the noOp method of the interface
	// returns true.
	if i.noOp() {
		return nil
	}

	if err := i.incrementGauge(name, labels); err != nil {
		return err
	}

	return nil
}

// DecrementGauge, is the exported function responsible for decrementing a
// gauge by 1 using a InstrumentationAPI interface. If the
// InstrumentationAPI method noOp() returns true, then no instrumentation is
// performed by this function.
func DecrementGauge(i InstrumentationAPI, name string, labels ...string) error {
	// Do not perform any instrumentation if the noOp method of the interface
	// returns true.
	if i.noOp() {
		return nil
	}

	if err := i.decreaseGauge(name, labels); err != nil {
		return err
	}

	return nil
}

// histogram, holds all the data required to handle a Prometheus histogram.
type histogram struct {
	// name, of histogram.
	name string
	// description, of histogram. It is equivalent to the 'Help' field of
	// prometheus.CounterOpts.
	description string
	// labelsNames, names of the labels used by a particular histogram.
	labelsNames []string
	// buckets, in which the different values measured are placed.
	buckets []float64
	// promObject, is the actual Prometheus object that hosts a histogram
	// with labels.
	promObject *prometheus.HistogramVec
}

// histograms, holds all the histograms registered for a particular
// instrumentation.
type histograms struct {
	histograms map[string]*histogram
}

// newHistogram, returns a new histogram pointer.
func newHistogram(name, description string, labels []string, buckets []float64) *histogram {
	h := new(histogram)
	h.name = name
	h.description = description
	h.labelsNames = labels
	h.buckets = buckets

	h.promObject = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    h.name,
			Help:    h.description,
			Buckets: h.buckets,
		},
		h.labelsNames,
	)
	return h
}

// insertHistogram, adds histogram to histograms map.
func (hs histograms) insertHistogram(h *histogram) error {

	if hs.histogramExists(h.name) {
		return fmt.Errorf("a histogram could not be inserted to the histograms map, since another histogram already exists with the same name (name='%s')", h.name)
	}

	// Add histogram to map.
	hs.histograms[h.name] = h

	return nil
}

// registerHistogram, register a histogram with the default Prometheus register.
// If the histogram does not exist within the histograms map this function
// returns an error.
func (hs histograms) registerHistogram(name string) error {
	if !hs.histogramExists(name) {
		return fmt.Errorf("could not register histogram %s, since it does not exist within the histograms map.", name)
	}

	h := hs.histograms[name]
	// Register the histogram.
	prometheus.Register(h.promObject)

	return nil
}

// histogramExists, checks if a histogram with a given `name` exists within the
// histograms map of the histograms struct. It returns false if the histogram
// does not exist.
func (hs histograms) histogramExists(name string) bool {
	_, ok := hs.histograms[name]
	return ok
}

// observeHistogram, interface-specific method used to observe the values of a
// histogram.
// If the histogram does not exist within the histograms map this method returns
// an error.
func (hs histograms) observeHistogram(value float64, name string, labels []string) error {
	if !hs.histogramExists(name) {
		return fmt.Errorf("could not increment histogram, since the histogram does not exist within histograms map")
	}

	hist := hs.histograms[name]

	// Increment histogram using the labels passed as parameters.
	// For more information about 'exploding' the labels []string slice, read
	// the following article on variadic functions:
	// https://www.digitalocean.com/community/tutorials/how-to-use-variadic-functions-in-go
	hist.promObject.WithLabelValues(labels...).Observe(value)

	return nil

}

// ObserveHistogram, is the exported function responsible for observing a value
// in a histogram defined in an InstrumentationAPI interface.
// If the InstrumentationAPI method noOp() returns true, then no instrumentation
// is performed by this function.
func ObserveHistogram(i InstrumentationAPI, value float64, name string, labels ...string) error {
	// Do not perform any instrumentation if the noOp method of the interface
	// returns true.
	if i.noOp() {
		return nil
	}

	if err := i.observeHistogram(value, name, labels); err != nil {
		return err
	}

	return nil
}

type InstrumentationAPI interface {
	incrementCounter(string, []string) error
	noOp() bool
	incrementGauge(string, []string) error
	decreaseGauge(string, []string) error
	observeHistogram(float64, string, []string) error
}

// Instrumentation, holds all the components used for instrumenting an
// application: counters, gauges, histograms, etc.
type Instrumentation struct {
	counters
	gauges
	histograms
}

// StartInstrumentation, returns an Instrumentation object that fulfils the
// InstrumentationAPI interface with all the instrumentation (counters, gauges
// and histograms) as defined in application.go. All the instrumentation is
// already register in Prometheus default register.
func startInstrumentation() (*Instrumentation, error) {
	ins := new(Instrumentation)

	// Counters section. ------------------------------------------------------
	// Start map with counters.
	ins.counters.counters = make(map[string]*counter)
	// Insert all counters defined in the struct 'applicationCounters' in the
	// application.go file.
	for _, c := range applicationCounters {
		// Create counter and insert it into counters map, if a counter with the
		// same name already exists the function returns an error.
		if err := ins.insertCounter(newCounter(c.name, c.description, c.labels)); err != nil {
			return ins, fmt.Errorf("error while inserting counter to counters map: %w", err)
		}
		if err := ins.registerCounter(c.name); err != nil {
			return ins, err
		}
	}

	// Gauges section. --------------------------------------------------------
	// Start map with gauges.
	ins.gauges.gauges = make(map[string]*gauge)
	// Insert all gauges defined in the struct 'applicationGauges' in the
	// application.go file.
	for _, g := range applicationGauges {
		// Create gauge and insert it into gauges map, if a gauge with the
		// same name already exists the function returns an error.
		if err := ins.insertGauge(newGauge(g.name, g.description, g.labels)); err != nil {
			return ins, fmt.Errorf("error while inserting gauge to gauges map: %w", err)
		}
		if err := ins.registerGauge(g.name); err != nil {
			return ins, err
		}
	}

	// Histograms section. --------------------------------------------------------
	// Start map with histograms.
	ins.histograms.histograms = make(map[string]*histogram)
	// Insert all histograms defined in the struct 'applicationHistograms' in
	// the application.go file.
	for _, h := range applicationHistograms {
		// Create histogram and insert it into histograms map, if a histogram
		// with the same name already exists the function returns an error.
		if err := ins.insertHistogram(newHistogram(h.name, h.description, h.labels, h.buckets)); err != nil {
			return ins, fmt.Errorf("error while inserting gauge to gauges map: %w", err)
		}
		if err := ins.registerHistogram(h.name); err != nil {
			return ins, err
		}
	}
	return ins, nil
}

// noOp, is a method required to properly decoupled the Instrumentation at the
// application level. If this method would return true, then the application,
// would not try to collect metrics.
func (i Instrumentation) noOp() bool {
	return false
}
