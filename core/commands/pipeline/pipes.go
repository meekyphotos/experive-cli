package pipeline

type Stage func() error

type Pipeline struct {
	stages []Stage
}

func (p *Pipeline) Add(stage ...Stage) {
	p.stages = append(p.stages, stage...)
}

func (p *Pipeline) RunPipe() error {
	var err error
	for _, s := range p.stages {
		err = s()
		if err != nil {
			return err
		}
	}
	return nil
}
