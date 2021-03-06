package idl

import (
	"fmt"
)

// union LogServiceRequestData switch (DdsData::LogServiceRequestType) {
func (p *parser) parseUnion() {
	p.advance()

	unionName := p.parseIdentifier()

	if unionName == "" {
		p.reportError(fmt.Errorf("expected type in union"))
		return
	}

	switchKeyword := p.parseIdentifier()

	if switchKeyword != keywordSwitch {
		p.reportError(fmt.Errorf("expected switch after type in union"))
		return
	}

	if p.tok().ID != TokenOpenBracket {
		p.reportError(fmt.Errorf("expected open bracket before type in union"))
		return
	}

	p.advance()

	switchType := p.parseType()

	if p.hasError() {
		p.reportError(fmt.Errorf("expected switch on type in union"))
		return
	}

	if p.tok().ID != TokenCloseBracket {
		p.reportError(fmt.Errorf("expected close bracket after type in union"))
		return
	}

	p.advance()

	if p.tok().ID != TokenOpenBrace {
		p.reportError(fmt.Errorf("expected open brace after type in union"))
		return
	}

	if parseDebug {
		fmt.Printf("Read union %s switching on type %s\n", unionName, switchType)
	}

	p.pushContext(contextUnion, unionName)
	p.currentUnion.Discriminant = switchType
}

//    case (DdsData::AnalogTimeSeries):
//          DdsData::TimeSeriesRequest analogTimeSeries; //@ID 1
func (p *parser) parseUnionMember() {

	keywordName := p.parseIdentifier()

	if keywordName != keywordCase {
		p.reportError(fmt.Errorf("expected case in union member"))
		return
	}

	if p.tok().ID != TokenOpenBracket {
		p.reportError(fmt.Errorf("expected open bracket before type in union member"))
		return
	}

	p.advance()

	switchType := p.parseType()

	if p.hasError() {
		p.reportError(fmt.Errorf("expected type in union member"))
		return
	}

	if p.tok().ID != TokenCloseBracket {
		p.reportError(fmt.Errorf("expected close bracket after type in union member"))
		return
	}

	p.advance()

	if p.tok().ID != TokenColon {
		p.reportError(fmt.Errorf("expected colon after close bracket in union member"))
		return
	}

	p.advance()

	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("expected var type in union member"))
		return
	}

	varType := p.parseType()

	if p.tok().ID != TokenIdentifier {
		p.reportError(fmt.Errorf("expected var name in union member"))
		return
	}

	varName := p.parseIdentifier()

	if p.tok().ID != TokenSemicolon {
		p.reportError(fmt.Errorf("expected semicolon at the end of  union member"))
		return
	}

	p.advance()

	if parseDebug {
		fmt.Printf("Read union member of type %s with var name %s (%s)\n", switchType, varName, varType)
	}

	p.currentUnion.Members = append(p.currentUnion.Members, UnionMember{
		CaseValue:  switchType,
		MemberType: varType,
		MemberName: varName,
	})
}
