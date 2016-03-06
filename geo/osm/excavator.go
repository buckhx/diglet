package osm

import (
	"github.com/buckhx/diglet/util"
	"github.com/qedus/osmpbf"
	"io"
	"os"
	"sync"
)

type Excavator struct {
	pbfpath         string
	pbffile         *os.File
	pbf             *osmpbf.Decoder
	nodes           chan *Node
	ways            chan *Way
	relations       chan *Relation
	errs            chan error
	dredgers        *sync.WaitGroup
	couriers        *sync.WaitGroup
	NodeCourier     func(<-chan *Node)
	WayCourier      func(<-chan *Way)
	RelationCourier func(<-chan *Relation)
}

func NewExcavator(pbfpath string) (ex *Excavator, err error) {
	file, err := os.Open(pbfpath)
	if err != nil {
		return
	}
	/*
		stat, err := file.Stat()
		if err != nil {
			return
		}
		est := stat.Size() >> 8
	*/
	pbf := osmpbf.NewDecoder(file)
	ex = &Excavator{
		pbfpath:   pbfpath,
		pbffile:   file,
		pbf:       pbf,
		nodes:     make(chan *Node, BlockSize),
		ways:      make(chan *Way, BlockSize),
		relations: make(chan *Relation, BlockSize),
		errs:      make(chan error),
		dredgers:  &sync.WaitGroup{},
		couriers:  &sync.WaitGroup{},
		NodeCourier: func(c <-chan *Node) {
			for _ = range c {
			}
		},
		WayCourier: func(c <-chan *Way) {
			for _ = range c {
			}
		},
		RelationCourier: func(c <-chan *Relation) {
			for _ = range c {
			}
		},
	}
	return
}

func (ex *Excavator) Start(workers int) (err error) {
	util.Info("Excavating %s with %d workers", ex.pbfpath, workers)
	err = ex.pbf.Start(workers)
	if err != nil {
		return
	}
	ex.dredgers.Add(workers)
	for i := 0; i < workers; i++ {
		util.Info("GO DREDGE %d", i)
		go ex.dredge()
	}
	go func() {
		ex.dredgers.Wait()
		ex.Close()
		util.Info("Done excavating, closing channels")
	}()
	ex.couriers.Add(workers)
	for i := 0; i < workers; i++ {
		util.Info("GO COURIER %d", i)
		go ex.courier()
	}
	ex.couriers.Wait()
	return
}

func (ex *Excavator) Restart(workers int) (err error) {
	err = ex.Reset()
	if err != nil {
		return
	}
	err = ex.Start(workers)
	return
}

func (ex *Excavator) Reset() (err error) {
	ex.Close()
	ex, err = NewExcavator(ex.pbfpath)
	return
}

func (ex *Excavator) Errors() <-chan error {
	return ex.errs
}

func (ex *Excavator) courier() {
	defer ex.couriers.Done()
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		ex.NodeCourier(ex.nodes)
	}()
	go func() {
		defer wg.Done()
		ex.WayCourier(ex.ways)
	}()
	go func() {
		defer wg.Done()
		ex.RelationCourier(ex.relations)
	}()
	wg.Wait()
}

func (ex *Excavator) dredge() {
	util.Info("Starting dredger")
	defer ex.dredgers.Done()
	for {
		if val, err := ex.pbf.Decode(); err == io.EOF {
			break
		} else if err != nil {
			ex.errs <- err
			break
		} else {
			switch val := val.(type) {
			case *osmpbf.Node:
				o := &Node{val}
				if o.Valid() {
					ex.nodes <- o
				}
			case *osmpbf.Way:
				o := &Way{val}
				if o.Valid() {
					ex.ways <- o
				}
			case *osmpbf.Relation:
				ex.relations <- &Relation{val}
			default:
				ex.errs <- util.Errorf("Unknown OSM Type %T %v", val, val)
			}
		}
	}
	util.Info("Closing dredger")
}

func (ex *Excavator) Close() {
	close(ex.nodes)
	close(ex.ways)
	close(ex.relations)
	ex.pbffile.Close()
	// close errs?
}
