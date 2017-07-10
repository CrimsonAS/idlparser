package idl

func (p *parser) parseUnion() {
	// ### TODO
	p.pushContext(contextUnion, "")
	for p.tok().Id != TokenCloseBrace {
		p.advance()
	}
}
