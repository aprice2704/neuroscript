// Generated from /home/aprice/dev/neuroscript/pkg/core/NeuroScript.g4 by ANTLR 4.13.1
import org.antlr.v4.runtime.atn.*;
import org.antlr.v4.runtime.dfa.DFA;
import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.misc.*;
import org.antlr.v4.runtime.tree.*;
import java.util.List;
import java.util.Iterator;
import java.util.ArrayList;

@SuppressWarnings({"all", "warnings", "unchecked", "unused", "cast", "CheckReturnValue"})
public class NeuroScriptParser extends Parser {
	static { RuntimeMetaData.checkVersion("4.13.1", RuntimeMetaData.VERSION); }

	protected static final DFA[] _decisionToDFA;
	protected static final PredictionContextCache _sharedContextCache =
		new PredictionContextCache();
	public static final int
		LINE_ESCAPE_GLOBAL=1, KW_ACOS=2, KW_AND=3, KW_AS=4, KW_ASIN=5, KW_ASK=6, 
		KW_ATAN=7, KW_BREAK=8, KW_CALL=9, KW_CLEAR=10, KW_CLEAR_ERROR=11, KW_CONTINUE=12, 
		KW_COS=13, KW_DO=14, KW_EACH=15, KW_ELSE=16, KW_EMIT=17, KW_ENDFOR=18, 
		KW_ENDFUNC=19, KW_ENDIF=20, KW_ENDON=21, KW_ENDWHILE=22, KW_ERROR=23, 
		KW_EVAL=24, KW_EVENT=25, KW_FAIL=26, KW_FALSE=27, KW_FOR=28, KW_FUNC=29, 
		KW_FUZZY=30, KW_IF=31, KW_IN=32, KW_INTO=33, KW_LAST=34, KW_LN=35, KW_LOG=36, 
		KW_MEANS=37, KW_MUST=38, KW_MUSTBE=39, KW_NAMED=40, KW_NEEDS=41, KW_NIL=42, 
		KW_NO=43, KW_NOT=44, KW_ON=45, KW_OPTIONAL=46, KW_OR=47, KW_RETURN=48, 
		KW_RETURNS=49, KW_SET=50, KW_SIN=51, KW_SOME=52, KW_TAN=53, KW_TIMEDATE=54, 
		KW_TOOL=55, KW_TRUE=56, KW_TYPEOF=57, KW_WHILE=58, STRING_LIT=59, TRIPLE_BACKTICK_STRING=60, 
		METADATA_LINE=61, NUMBER_LIT=62, IDENTIFIER=63, ASSIGN=64, PLUS=65, MINUS=66, 
		STAR=67, SLASH=68, PERCENT=69, STAR_STAR=70, AMPERSAND=71, PIPE=72, CARET=73, 
		TILDE=74, LPAREN=75, RPAREN=76, COMMA=77, LBRACK=78, RBRACK=79, LBRACE=80, 
		RBRACE=81, COLON=82, DOT=83, PLACEHOLDER_START=84, PLACEHOLDER_END=85, 
		EQ=86, NEQ=87, GT=88, LT=89, GTE=90, LTE=91, LINE_COMMENT=92, NEWLINE=93, 
		WS=94;
	public static final int
		RULE_program = 0, RULE_file_header = 1, RULE_code_block = 2, RULE_procedure_definition = 3, 
		RULE_signature_part = 4, RULE_needs_clause = 5, RULE_optional_clause = 6, 
		RULE_returns_clause = 7, RULE_param_list = 8, RULE_metadata_block = 9, 
		RULE_statement_list = 10, RULE_body_line = 11, RULE_statement = 12, RULE_simple_statement = 13, 
		RULE_block_statement = 14, RULE_on_stmt = 15, RULE_error_handler = 16, 
		RULE_event_handler = 17, RULE_clearEventStmt = 18, RULE_lvalue = 19, RULE_lvalue_list = 20, 
		RULE_set_statement = 21, RULE_call_statement = 22, RULE_return_statement = 23, 
		RULE_emit_statement = 24, RULE_must_statement = 25, RULE_fail_statement = 26, 
		RULE_clearErrorStmt = 27, RULE_ask_stmt = 28, RULE_break_statement = 29, 
		RULE_continue_statement = 30, RULE_if_statement = 31, RULE_while_statement = 32, 
		RULE_for_each_statement = 33, RULE_qualified_identifier = 34, RULE_call_target = 35, 
		RULE_expression = 36, RULE_logical_or_expr = 37, RULE_logical_and_expr = 38, 
		RULE_bitwise_or_expr = 39, RULE_bitwise_xor_expr = 40, RULE_bitwise_and_expr = 41, 
		RULE_equality_expr = 42, RULE_relational_expr = 43, RULE_additive_expr = 44, 
		RULE_multiplicative_expr = 45, RULE_unary_expr = 46, RULE_power_expr = 47, 
		RULE_accessor_expr = 48, RULE_primary = 49, RULE_callable_expr = 50, RULE_placeholder = 51, 
		RULE_literal = 52, RULE_nil_literal = 53, RULE_boolean_literal = 54, RULE_list_literal = 55, 
		RULE_map_literal = 56, RULE_expression_list_opt = 57, RULE_expression_list = 58, 
		RULE_map_entry_list_opt = 59, RULE_map_entry_list = 60, RULE_map_entry = 61;
	private static String[] makeRuleNames() {
		return new String[] {
			"program", "file_header", "code_block", "procedure_definition", "signature_part", 
			"needs_clause", "optional_clause", "returns_clause", "param_list", "metadata_block", 
			"statement_list", "body_line", "statement", "simple_statement", "block_statement", 
			"on_stmt", "error_handler", "event_handler", "clearEventStmt", "lvalue", 
			"lvalue_list", "set_statement", "call_statement", "return_statement", 
			"emit_statement", "must_statement", "fail_statement", "clearErrorStmt", 
			"ask_stmt", "break_statement", "continue_statement", "if_statement", 
			"while_statement", "for_each_statement", "qualified_identifier", "call_target", 
			"expression", "logical_or_expr", "logical_and_expr", "bitwise_or_expr", 
			"bitwise_xor_expr", "bitwise_and_expr", "equality_expr", "relational_expr", 
			"additive_expr", "multiplicative_expr", "unary_expr", "power_expr", "accessor_expr", 
			"primary", "callable_expr", "placeholder", "literal", "nil_literal", 
			"boolean_literal", "list_literal", "map_literal", "expression_list_opt", 
			"expression_list", "map_entry_list_opt", "map_entry_list", "map_entry"
		};
	}
	public static final String[] ruleNames = makeRuleNames();

	private static String[] makeLiteralNames() {
		return new String[] {
			null, null, "'acos'", "'and'", "'as'", "'asin'", "'ask'", "'atan'", "'break'", 
			"'call'", "'clear'", "'clear_error'", "'continue'", "'cos'", "'do'", 
			"'each'", "'else'", "'emit'", "'endfor'", "'endfunc'", "'endif'", "'endon'", 
			"'endwhile'", "'error'", "'eval'", "'event'", "'fail'", "'false'", "'for'", 
			"'func'", "'fuzzy'", "'if'", "'in'", "'into'", "'last'", "'ln'", "'log'", 
			"'means'", "'must'", "'mustbe'", "'named'", "'needs'", "'nil'", "'no'", 
			"'not'", "'on'", "'optional'", "'or'", "'return'", "'returns'", "'set'", 
			"'sin'", "'some'", "'tan'", "'timedate'", "'tool'", "'true'", "'typeof'", 
			"'while'", null, null, null, null, null, "'='", "'+'", "'-'", "'*'", 
			"'/'", "'%'", "'**'", "'&'", "'|'", "'^'", "'~'", "'('", "')'", "','", 
			"'['", "']'", "'{'", "'}'", "':'", "'.'", "'{{'", "'}}'", "'=='", "'!='", 
			"'>'", "'<'", "'>='", "'<='"
		};
	}
	private static final String[] _LITERAL_NAMES = makeLiteralNames();
	private static String[] makeSymbolicNames() {
		return new String[] {
			null, "LINE_ESCAPE_GLOBAL", "KW_ACOS", "KW_AND", "KW_AS", "KW_ASIN", 
			"KW_ASK", "KW_ATAN", "KW_BREAK", "KW_CALL", "KW_CLEAR", "KW_CLEAR_ERROR", 
			"KW_CONTINUE", "KW_COS", "KW_DO", "KW_EACH", "KW_ELSE", "KW_EMIT", "KW_ENDFOR", 
			"KW_ENDFUNC", "KW_ENDIF", "KW_ENDON", "KW_ENDWHILE", "KW_ERROR", "KW_EVAL", 
			"KW_EVENT", "KW_FAIL", "KW_FALSE", "KW_FOR", "KW_FUNC", "KW_FUZZY", "KW_IF", 
			"KW_IN", "KW_INTO", "KW_LAST", "KW_LN", "KW_LOG", "KW_MEANS", "KW_MUST", 
			"KW_MUSTBE", "KW_NAMED", "KW_NEEDS", "KW_NIL", "KW_NO", "KW_NOT", "KW_ON", 
			"KW_OPTIONAL", "KW_OR", "KW_RETURN", "KW_RETURNS", "KW_SET", "KW_SIN", 
			"KW_SOME", "KW_TAN", "KW_TIMEDATE", "KW_TOOL", "KW_TRUE", "KW_TYPEOF", 
			"KW_WHILE", "STRING_LIT", "TRIPLE_BACKTICK_STRING", "METADATA_LINE", 
			"NUMBER_LIT", "IDENTIFIER", "ASSIGN", "PLUS", "MINUS", "STAR", "SLASH", 
			"PERCENT", "STAR_STAR", "AMPERSAND", "PIPE", "CARET", "TILDE", "LPAREN", 
			"RPAREN", "COMMA", "LBRACK", "RBRACK", "LBRACE", "RBRACE", "COLON", "DOT", 
			"PLACEHOLDER_START", "PLACEHOLDER_END", "EQ", "NEQ", "GT", "LT", "GTE", 
			"LTE", "LINE_COMMENT", "NEWLINE", "WS"
		};
	}
	private static final String[] _SYMBOLIC_NAMES = makeSymbolicNames();
	public static final Vocabulary VOCABULARY = new VocabularyImpl(_LITERAL_NAMES, _SYMBOLIC_NAMES);

	/**
	 * @deprecated Use {@link #VOCABULARY} instead.
	 */
	@Deprecated
	public static final String[] tokenNames;
	static {
		tokenNames = new String[_SYMBOLIC_NAMES.length];
		for (int i = 0; i < tokenNames.length; i++) {
			tokenNames[i] = VOCABULARY.getLiteralName(i);
			if (tokenNames[i] == null) {
				tokenNames[i] = VOCABULARY.getSymbolicName(i);
			}

			if (tokenNames[i] == null) {
				tokenNames[i] = "<INVALID>";
			}
		}
	}

	@Override
	@Deprecated
	public String[] getTokenNames() {
		return tokenNames;
	}

	@Override

	public Vocabulary getVocabulary() {
		return VOCABULARY;
	}

	@Override
	public String getGrammarFileName() { return "NeuroScript.g4"; }

	@Override
	public String[] getRuleNames() { return ruleNames; }

	@Override
	public String getSerializedATN() { return _serializedATN; }

	@Override
	public ATN getATN() { return _ATN; }

	public NeuroScriptParser(TokenStream input) {
		super(input);
		_interp = new ParserATNSimulator(this,_ATN,_decisionToDFA,_sharedContextCache);
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ProgramContext extends ParserRuleContext {
		public File_headerContext file_header() {
			return getRuleContext(File_headerContext.class,0);
		}
		public TerminalNode EOF() { return getToken(NeuroScriptParser.EOF, 0); }
		public List<Code_blockContext> code_block() {
			return getRuleContexts(Code_blockContext.class);
		}
		public Code_blockContext code_block(int i) {
			return getRuleContext(Code_blockContext.class,i);
		}
		public ProgramContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_program; }
	}

	public final ProgramContext program() throws RecognitionException {
		ProgramContext _localctx = new ProgramContext(_ctx, getState());
		enterRule(_localctx, 0, RULE_program);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(124);
			file_header();
			setState(128);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==KW_FUNC || _la==KW_ON) {
				{
				{
				setState(125);
				code_block();
				}
				}
				setState(130);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(131);
			match(EOF);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class File_headerContext extends ParserRuleContext {
		public List<TerminalNode> METADATA_LINE() { return getTokens(NeuroScriptParser.METADATA_LINE); }
		public TerminalNode METADATA_LINE(int i) {
			return getToken(NeuroScriptParser.METADATA_LINE, i);
		}
		public List<TerminalNode> NEWLINE() { return getTokens(NeuroScriptParser.NEWLINE); }
		public TerminalNode NEWLINE(int i) {
			return getToken(NeuroScriptParser.NEWLINE, i);
		}
		public File_headerContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_file_header; }
	}

	public final File_headerContext file_header() throws RecognitionException {
		File_headerContext _localctx = new File_headerContext(_ctx, getState());
		enterRule(_localctx, 2, RULE_file_header);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(136);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==METADATA_LINE || _la==NEWLINE) {
				{
				{
				setState(133);
				_la = _input.LA(1);
				if ( !(_la==METADATA_LINE || _la==NEWLINE) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				}
				}
				setState(138);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Code_blockContext extends ParserRuleContext {
		public Procedure_definitionContext procedure_definition() {
			return getRuleContext(Procedure_definitionContext.class,0);
		}
		public On_stmtContext on_stmt() {
			return getRuleContext(On_stmtContext.class,0);
		}
		public List<TerminalNode> NEWLINE() { return getTokens(NeuroScriptParser.NEWLINE); }
		public TerminalNode NEWLINE(int i) {
			return getToken(NeuroScriptParser.NEWLINE, i);
		}
		public Code_blockContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_code_block; }
	}

	public final Code_blockContext code_block() throws RecognitionException {
		Code_blockContext _localctx = new Code_blockContext(_ctx, getState());
		enterRule(_localctx, 4, RULE_code_block);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(141);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_FUNC:
				{
				setState(139);
				procedure_definition();
				}
				break;
			case KW_ON:
				{
				setState(140);
				on_stmt();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			setState(146);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==NEWLINE) {
				{
				{
				setState(143);
				match(NEWLINE);
				}
				}
				setState(148);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Procedure_definitionContext extends ParserRuleContext {
		public TerminalNode KW_FUNC() { return getToken(NeuroScriptParser.KW_FUNC, 0); }
		public TerminalNode IDENTIFIER() { return getToken(NeuroScriptParser.IDENTIFIER, 0); }
		public Signature_partContext signature_part() {
			return getRuleContext(Signature_partContext.class,0);
		}
		public TerminalNode KW_MEANS() { return getToken(NeuroScriptParser.KW_MEANS, 0); }
		public TerminalNode NEWLINE() { return getToken(NeuroScriptParser.NEWLINE, 0); }
		public Metadata_blockContext metadata_block() {
			return getRuleContext(Metadata_blockContext.class,0);
		}
		public Statement_listContext statement_list() {
			return getRuleContext(Statement_listContext.class,0);
		}
		public TerminalNode KW_ENDFUNC() { return getToken(NeuroScriptParser.KW_ENDFUNC, 0); }
		public Procedure_definitionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_procedure_definition; }
	}

	public final Procedure_definitionContext procedure_definition() throws RecognitionException {
		Procedure_definitionContext _localctx = new Procedure_definitionContext(_ctx, getState());
		enterRule(_localctx, 6, RULE_procedure_definition);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(149);
			match(KW_FUNC);
			setState(150);
			match(IDENTIFIER);
			setState(151);
			signature_part();
			setState(152);
			match(KW_MEANS);
			setState(153);
			match(NEWLINE);
			setState(154);
			metadata_block();
			setState(155);
			statement_list();
			setState(156);
			match(KW_ENDFUNC);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Signature_partContext extends ParserRuleContext {
		public TerminalNode LPAREN() { return getToken(NeuroScriptParser.LPAREN, 0); }
		public TerminalNode RPAREN() { return getToken(NeuroScriptParser.RPAREN, 0); }
		public List<Needs_clauseContext> needs_clause() {
			return getRuleContexts(Needs_clauseContext.class);
		}
		public Needs_clauseContext needs_clause(int i) {
			return getRuleContext(Needs_clauseContext.class,i);
		}
		public List<Optional_clauseContext> optional_clause() {
			return getRuleContexts(Optional_clauseContext.class);
		}
		public Optional_clauseContext optional_clause(int i) {
			return getRuleContext(Optional_clauseContext.class,i);
		}
		public List<Returns_clauseContext> returns_clause() {
			return getRuleContexts(Returns_clauseContext.class);
		}
		public Returns_clauseContext returns_clause(int i) {
			return getRuleContext(Returns_clauseContext.class,i);
		}
		public Signature_partContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_signature_part; }
	}

	public final Signature_partContext signature_part() throws RecognitionException {
		Signature_partContext _localctx = new Signature_partContext(_ctx, getState());
		enterRule(_localctx, 8, RULE_signature_part);
		int _la;
		try {
			setState(176);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case LPAREN:
				enterOuterAlt(_localctx, 1);
				{
				setState(158);
				match(LPAREN);
				setState(164);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while ((((_la) & ~0x3f) == 0 && ((1L << _la) & 635517720854528L) != 0)) {
					{
					setState(162);
					_errHandler.sync(this);
					switch (_input.LA(1)) {
					case KW_NEEDS:
						{
						setState(159);
						needs_clause();
						}
						break;
					case KW_OPTIONAL:
						{
						setState(160);
						optional_clause();
						}
						break;
					case KW_RETURNS:
						{
						setState(161);
						returns_clause();
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					}
					setState(166);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(167);
				match(RPAREN);
				}
				break;
			case KW_NEEDS:
			case KW_OPTIONAL:
			case KW_RETURNS:
				enterOuterAlt(_localctx, 2);
				{
				setState(171); 
				_errHandler.sync(this);
				_la = _input.LA(1);
				do {
					{
					setState(171);
					_errHandler.sync(this);
					switch (_input.LA(1)) {
					case KW_NEEDS:
						{
						setState(168);
						needs_clause();
						}
						break;
					case KW_OPTIONAL:
						{
						setState(169);
						optional_clause();
						}
						break;
					case KW_RETURNS:
						{
						setState(170);
						returns_clause();
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					}
					setState(173); 
					_errHandler.sync(this);
					_la = _input.LA(1);
				} while ( (((_la) & ~0x3f) == 0 && ((1L << _la) & 635517720854528L) != 0) );
				}
				break;
			case KW_MEANS:
				enterOuterAlt(_localctx, 3);
				{
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Needs_clauseContext extends ParserRuleContext {
		public TerminalNode KW_NEEDS() { return getToken(NeuroScriptParser.KW_NEEDS, 0); }
		public Param_listContext param_list() {
			return getRuleContext(Param_listContext.class,0);
		}
		public Needs_clauseContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_needs_clause; }
	}

	public final Needs_clauseContext needs_clause() throws RecognitionException {
		Needs_clauseContext _localctx = new Needs_clauseContext(_ctx, getState());
		enterRule(_localctx, 10, RULE_needs_clause);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(178);
			match(KW_NEEDS);
			setState(179);
			param_list();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Optional_clauseContext extends ParserRuleContext {
		public TerminalNode KW_OPTIONAL() { return getToken(NeuroScriptParser.KW_OPTIONAL, 0); }
		public Param_listContext param_list() {
			return getRuleContext(Param_listContext.class,0);
		}
		public Optional_clauseContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_optional_clause; }
	}

	public final Optional_clauseContext optional_clause() throws RecognitionException {
		Optional_clauseContext _localctx = new Optional_clauseContext(_ctx, getState());
		enterRule(_localctx, 12, RULE_optional_clause);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(181);
			match(KW_OPTIONAL);
			setState(182);
			param_list();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Returns_clauseContext extends ParserRuleContext {
		public TerminalNode KW_RETURNS() { return getToken(NeuroScriptParser.KW_RETURNS, 0); }
		public Param_listContext param_list() {
			return getRuleContext(Param_listContext.class,0);
		}
		public Returns_clauseContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_returns_clause; }
	}

	public final Returns_clauseContext returns_clause() throws RecognitionException {
		Returns_clauseContext _localctx = new Returns_clauseContext(_ctx, getState());
		enterRule(_localctx, 14, RULE_returns_clause);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(184);
			match(KW_RETURNS);
			setState(185);
			param_list();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Param_listContext extends ParserRuleContext {
		public List<TerminalNode> IDENTIFIER() { return getTokens(NeuroScriptParser.IDENTIFIER); }
		public TerminalNode IDENTIFIER(int i) {
			return getToken(NeuroScriptParser.IDENTIFIER, i);
		}
		public List<TerminalNode> COMMA() { return getTokens(NeuroScriptParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(NeuroScriptParser.COMMA, i);
		}
		public Param_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_param_list; }
	}

	public final Param_listContext param_list() throws RecognitionException {
		Param_listContext _localctx = new Param_listContext(_ctx, getState());
		enterRule(_localctx, 16, RULE_param_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(187);
			match(IDENTIFIER);
			setState(192);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(188);
				match(COMMA);
				setState(189);
				match(IDENTIFIER);
				}
				}
				setState(194);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Metadata_blockContext extends ParserRuleContext {
		public List<TerminalNode> METADATA_LINE() { return getTokens(NeuroScriptParser.METADATA_LINE); }
		public TerminalNode METADATA_LINE(int i) {
			return getToken(NeuroScriptParser.METADATA_LINE, i);
		}
		public List<TerminalNode> NEWLINE() { return getTokens(NeuroScriptParser.NEWLINE); }
		public TerminalNode NEWLINE(int i) {
			return getToken(NeuroScriptParser.NEWLINE, i);
		}
		public Metadata_blockContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_metadata_block; }
	}

	public final Metadata_blockContext metadata_block() throws RecognitionException {
		Metadata_blockContext _localctx = new Metadata_blockContext(_ctx, getState());
		enterRule(_localctx, 18, RULE_metadata_block);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(199);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==METADATA_LINE) {
				{
				{
				setState(195);
				match(METADATA_LINE);
				setState(196);
				match(NEWLINE);
				}
				}
				setState(201);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Statement_listContext extends ParserRuleContext {
		public List<Body_lineContext> body_line() {
			return getRuleContexts(Body_lineContext.class);
		}
		public Body_lineContext body_line(int i) {
			return getRuleContext(Body_lineContext.class,i);
		}
		public Statement_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_statement_list; }
	}

	public final Statement_listContext statement_list() throws RecognitionException {
		Statement_listContext _localctx = new Statement_listContext(_ctx, getState());
		enterRule(_localctx, 20, RULE_statement_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(205);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while ((((_la) & ~0x3f) == 0 && ((1L << _la) & 289673762524241728L) != 0) || _la==NEWLINE) {
				{
				{
				setState(202);
				body_line();
				}
				}
				setState(207);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Body_lineContext extends ParserRuleContext {
		public StatementContext statement() {
			return getRuleContext(StatementContext.class,0);
		}
		public TerminalNode NEWLINE() { return getToken(NeuroScriptParser.NEWLINE, 0); }
		public Body_lineContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_body_line; }
	}

	public final Body_lineContext body_line() throws RecognitionException {
		Body_lineContext _localctx = new Body_lineContext(_ctx, getState());
		enterRule(_localctx, 22, RULE_body_line);
		try {
			setState(212);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_ASK:
			case KW_BREAK:
			case KW_CALL:
			case KW_CLEAR:
			case KW_CLEAR_ERROR:
			case KW_CONTINUE:
			case KW_EMIT:
			case KW_FAIL:
			case KW_FOR:
			case KW_IF:
			case KW_MUST:
			case KW_MUSTBE:
			case KW_ON:
			case KW_RETURN:
			case KW_SET:
			case KW_WHILE:
				enterOuterAlt(_localctx, 1);
				{
				setState(208);
				statement();
				setState(209);
				match(NEWLINE);
				}
				break;
			case NEWLINE:
				enterOuterAlt(_localctx, 2);
				{
				setState(211);
				match(NEWLINE);
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class StatementContext extends ParserRuleContext {
		public Simple_statementContext simple_statement() {
			return getRuleContext(Simple_statementContext.class,0);
		}
		public Block_statementContext block_statement() {
			return getRuleContext(Block_statementContext.class,0);
		}
		public On_stmtContext on_stmt() {
			return getRuleContext(On_stmtContext.class,0);
		}
		public StatementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_statement; }
	}

	public final StatementContext statement() throws RecognitionException {
		StatementContext _localctx = new StatementContext(_ctx, getState());
		enterRule(_localctx, 24, RULE_statement);
		try {
			setState(217);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_ASK:
			case KW_BREAK:
			case KW_CALL:
			case KW_CLEAR:
			case KW_CLEAR_ERROR:
			case KW_CONTINUE:
			case KW_EMIT:
			case KW_FAIL:
			case KW_MUST:
			case KW_MUSTBE:
			case KW_RETURN:
			case KW_SET:
				enterOuterAlt(_localctx, 1);
				{
				setState(214);
				simple_statement();
				}
				break;
			case KW_FOR:
			case KW_IF:
			case KW_WHILE:
				enterOuterAlt(_localctx, 2);
				{
				setState(215);
				block_statement();
				}
				break;
			case KW_ON:
				enterOuterAlt(_localctx, 3);
				{
				setState(216);
				on_stmt();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Simple_statementContext extends ParserRuleContext {
		public Set_statementContext set_statement() {
			return getRuleContext(Set_statementContext.class,0);
		}
		public Call_statementContext call_statement() {
			return getRuleContext(Call_statementContext.class,0);
		}
		public Return_statementContext return_statement() {
			return getRuleContext(Return_statementContext.class,0);
		}
		public Emit_statementContext emit_statement() {
			return getRuleContext(Emit_statementContext.class,0);
		}
		public Must_statementContext must_statement() {
			return getRuleContext(Must_statementContext.class,0);
		}
		public Fail_statementContext fail_statement() {
			return getRuleContext(Fail_statementContext.class,0);
		}
		public ClearErrorStmtContext clearErrorStmt() {
			return getRuleContext(ClearErrorStmtContext.class,0);
		}
		public ClearEventStmtContext clearEventStmt() {
			return getRuleContext(ClearEventStmtContext.class,0);
		}
		public Ask_stmtContext ask_stmt() {
			return getRuleContext(Ask_stmtContext.class,0);
		}
		public Break_statementContext break_statement() {
			return getRuleContext(Break_statementContext.class,0);
		}
		public Continue_statementContext continue_statement() {
			return getRuleContext(Continue_statementContext.class,0);
		}
		public Simple_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_simple_statement; }
	}

	public final Simple_statementContext simple_statement() throws RecognitionException {
		Simple_statementContext _localctx = new Simple_statementContext(_ctx, getState());
		enterRule(_localctx, 26, RULE_simple_statement);
		try {
			setState(230);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_SET:
				enterOuterAlt(_localctx, 1);
				{
				setState(219);
				set_statement();
				}
				break;
			case KW_CALL:
				enterOuterAlt(_localctx, 2);
				{
				setState(220);
				call_statement();
				}
				break;
			case KW_RETURN:
				enterOuterAlt(_localctx, 3);
				{
				setState(221);
				return_statement();
				}
				break;
			case KW_EMIT:
				enterOuterAlt(_localctx, 4);
				{
				setState(222);
				emit_statement();
				}
				break;
			case KW_MUST:
			case KW_MUSTBE:
				enterOuterAlt(_localctx, 5);
				{
				setState(223);
				must_statement();
				}
				break;
			case KW_FAIL:
				enterOuterAlt(_localctx, 6);
				{
				setState(224);
				fail_statement();
				}
				break;
			case KW_CLEAR_ERROR:
				enterOuterAlt(_localctx, 7);
				{
				setState(225);
				clearErrorStmt();
				}
				break;
			case KW_CLEAR:
				enterOuterAlt(_localctx, 8);
				{
				setState(226);
				clearEventStmt();
				}
				break;
			case KW_ASK:
				enterOuterAlt(_localctx, 9);
				{
				setState(227);
				ask_stmt();
				}
				break;
			case KW_BREAK:
				enterOuterAlt(_localctx, 10);
				{
				setState(228);
				break_statement();
				}
				break;
			case KW_CONTINUE:
				enterOuterAlt(_localctx, 11);
				{
				setState(229);
				continue_statement();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Block_statementContext extends ParserRuleContext {
		public If_statementContext if_statement() {
			return getRuleContext(If_statementContext.class,0);
		}
		public While_statementContext while_statement() {
			return getRuleContext(While_statementContext.class,0);
		}
		public For_each_statementContext for_each_statement() {
			return getRuleContext(For_each_statementContext.class,0);
		}
		public Block_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_block_statement; }
	}

	public final Block_statementContext block_statement() throws RecognitionException {
		Block_statementContext _localctx = new Block_statementContext(_ctx, getState());
		enterRule(_localctx, 28, RULE_block_statement);
		try {
			setState(235);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_IF:
				enterOuterAlt(_localctx, 1);
				{
				setState(232);
				if_statement();
				}
				break;
			case KW_WHILE:
				enterOuterAlt(_localctx, 2);
				{
				setState(233);
				while_statement();
				}
				break;
			case KW_FOR:
				enterOuterAlt(_localctx, 3);
				{
				setState(234);
				for_each_statement();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class On_stmtContext extends ParserRuleContext {
		public TerminalNode KW_ON() { return getToken(NeuroScriptParser.KW_ON, 0); }
		public Error_handlerContext error_handler() {
			return getRuleContext(Error_handlerContext.class,0);
		}
		public Event_handlerContext event_handler() {
			return getRuleContext(Event_handlerContext.class,0);
		}
		public On_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_on_stmt; }
	}

	public final On_stmtContext on_stmt() throws RecognitionException {
		On_stmtContext _localctx = new On_stmtContext(_ctx, getState());
		enterRule(_localctx, 30, RULE_on_stmt);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(237);
			match(KW_ON);
			setState(240);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_ERROR:
				{
				setState(238);
				error_handler();
				}
				break;
			case KW_EVENT:
				{
				setState(239);
				event_handler();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Error_handlerContext extends ParserRuleContext {
		public TerminalNode KW_ERROR() { return getToken(NeuroScriptParser.KW_ERROR, 0); }
		public TerminalNode KW_DO() { return getToken(NeuroScriptParser.KW_DO, 0); }
		public TerminalNode NEWLINE() { return getToken(NeuroScriptParser.NEWLINE, 0); }
		public Statement_listContext statement_list() {
			return getRuleContext(Statement_listContext.class,0);
		}
		public TerminalNode KW_ENDON() { return getToken(NeuroScriptParser.KW_ENDON, 0); }
		public Error_handlerContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_error_handler; }
	}

	public final Error_handlerContext error_handler() throws RecognitionException {
		Error_handlerContext _localctx = new Error_handlerContext(_ctx, getState());
		enterRule(_localctx, 32, RULE_error_handler);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(242);
			match(KW_ERROR);
			setState(243);
			match(KW_DO);
			setState(244);
			match(NEWLINE);
			setState(245);
			statement_list();
			setState(246);
			match(KW_ENDON);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Event_handlerContext extends ParserRuleContext {
		public TerminalNode KW_EVENT() { return getToken(NeuroScriptParser.KW_EVENT, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode KW_DO() { return getToken(NeuroScriptParser.KW_DO, 0); }
		public TerminalNode NEWLINE() { return getToken(NeuroScriptParser.NEWLINE, 0); }
		public Statement_listContext statement_list() {
			return getRuleContext(Statement_listContext.class,0);
		}
		public TerminalNode KW_ENDON() { return getToken(NeuroScriptParser.KW_ENDON, 0); }
		public TerminalNode KW_NAMED() { return getToken(NeuroScriptParser.KW_NAMED, 0); }
		public TerminalNode STRING_LIT() { return getToken(NeuroScriptParser.STRING_LIT, 0); }
		public TerminalNode KW_AS() { return getToken(NeuroScriptParser.KW_AS, 0); }
		public TerminalNode IDENTIFIER() { return getToken(NeuroScriptParser.IDENTIFIER, 0); }
		public Event_handlerContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_event_handler; }
	}

	public final Event_handlerContext event_handler() throws RecognitionException {
		Event_handlerContext _localctx = new Event_handlerContext(_ctx, getState());
		enterRule(_localctx, 34, RULE_event_handler);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(248);
			match(KW_EVENT);
			setState(249);
			expression();
			setState(252);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==KW_NAMED) {
				{
				setState(250);
				match(KW_NAMED);
				setState(251);
				match(STRING_LIT);
				}
			}

			setState(256);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==KW_AS) {
				{
				setState(254);
				match(KW_AS);
				setState(255);
				match(IDENTIFIER);
				}
			}

			setState(258);
			match(KW_DO);
			setState(259);
			match(NEWLINE);
			setState(260);
			statement_list();
			setState(261);
			match(KW_ENDON);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ClearEventStmtContext extends ParserRuleContext {
		public TerminalNode KW_CLEAR() { return getToken(NeuroScriptParser.KW_CLEAR, 0); }
		public TerminalNode KW_EVENT() { return getToken(NeuroScriptParser.KW_EVENT, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode KW_NAMED() { return getToken(NeuroScriptParser.KW_NAMED, 0); }
		public TerminalNode STRING_LIT() { return getToken(NeuroScriptParser.STRING_LIT, 0); }
		public ClearEventStmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_clearEventStmt; }
	}

	public final ClearEventStmtContext clearEventStmt() throws RecognitionException {
		ClearEventStmtContext _localctx = new ClearEventStmtContext(_ctx, getState());
		enterRule(_localctx, 36, RULE_clearEventStmt);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(263);
			match(KW_CLEAR);
			setState(264);
			match(KW_EVENT);
			setState(268);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_ACOS:
			case KW_ASIN:
			case KW_ATAN:
			case KW_COS:
			case KW_EVAL:
			case KW_FALSE:
			case KW_LAST:
			case KW_LN:
			case KW_LOG:
			case KW_MUST:
			case KW_NIL:
			case KW_NO:
			case KW_NOT:
			case KW_SIN:
			case KW_SOME:
			case KW_TAN:
			case KW_TOOL:
			case KW_TRUE:
			case KW_TYPEOF:
			case STRING_LIT:
			case TRIPLE_BACKTICK_STRING:
			case NUMBER_LIT:
			case IDENTIFIER:
			case MINUS:
			case TILDE:
			case LPAREN:
			case LBRACK:
			case LBRACE:
			case PLACEHOLDER_START:
				{
				setState(265);
				expression();
				}
				break;
			case KW_NAMED:
				{
				setState(266);
				match(KW_NAMED);
				setState(267);
				match(STRING_LIT);
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class LvalueContext extends ParserRuleContext {
		public List<TerminalNode> IDENTIFIER() { return getTokens(NeuroScriptParser.IDENTIFIER); }
		public TerminalNode IDENTIFIER(int i) {
			return getToken(NeuroScriptParser.IDENTIFIER, i);
		}
		public List<TerminalNode> LBRACK() { return getTokens(NeuroScriptParser.LBRACK); }
		public TerminalNode LBRACK(int i) {
			return getToken(NeuroScriptParser.LBRACK, i);
		}
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public List<TerminalNode> RBRACK() { return getTokens(NeuroScriptParser.RBRACK); }
		public TerminalNode RBRACK(int i) {
			return getToken(NeuroScriptParser.RBRACK, i);
		}
		public List<TerminalNode> DOT() { return getTokens(NeuroScriptParser.DOT); }
		public TerminalNode DOT(int i) {
			return getToken(NeuroScriptParser.DOT, i);
		}
		public LvalueContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_lvalue; }
	}

	public final LvalueContext lvalue() throws RecognitionException {
		LvalueContext _localctx = new LvalueContext(_ctx, getState());
		enterRule(_localctx, 38, RULE_lvalue);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(270);
			match(IDENTIFIER);
			setState(279);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==LBRACK || _la==DOT) {
				{
				setState(277);
				_errHandler.sync(this);
				switch (_input.LA(1)) {
				case LBRACK:
					{
					setState(271);
					match(LBRACK);
					setState(272);
					expression();
					setState(273);
					match(RBRACK);
					}
					break;
				case DOT:
					{
					setState(275);
					match(DOT);
					setState(276);
					match(IDENTIFIER);
					}
					break;
				default:
					throw new NoViableAltException(this);
				}
				}
				setState(281);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Lvalue_listContext extends ParserRuleContext {
		public List<LvalueContext> lvalue() {
			return getRuleContexts(LvalueContext.class);
		}
		public LvalueContext lvalue(int i) {
			return getRuleContext(LvalueContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(NeuroScriptParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(NeuroScriptParser.COMMA, i);
		}
		public Lvalue_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_lvalue_list; }
	}

	public final Lvalue_listContext lvalue_list() throws RecognitionException {
		Lvalue_listContext _localctx = new Lvalue_listContext(_ctx, getState());
		enterRule(_localctx, 40, RULE_lvalue_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(282);
			lvalue();
			setState(287);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(283);
				match(COMMA);
				setState(284);
				lvalue();
				}
				}
				setState(289);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Set_statementContext extends ParserRuleContext {
		public TerminalNode KW_SET() { return getToken(NeuroScriptParser.KW_SET, 0); }
		public Lvalue_listContext lvalue_list() {
			return getRuleContext(Lvalue_listContext.class,0);
		}
		public TerminalNode ASSIGN() { return getToken(NeuroScriptParser.ASSIGN, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public Set_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_set_statement; }
	}

	public final Set_statementContext set_statement() throws RecognitionException {
		Set_statementContext _localctx = new Set_statementContext(_ctx, getState());
		enterRule(_localctx, 42, RULE_set_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(290);
			match(KW_SET);
			setState(291);
			lvalue_list();
			setState(292);
			match(ASSIGN);
			setState(293);
			expression();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Call_statementContext extends ParserRuleContext {
		public TerminalNode KW_CALL() { return getToken(NeuroScriptParser.KW_CALL, 0); }
		public Callable_exprContext callable_expr() {
			return getRuleContext(Callable_exprContext.class,0);
		}
		public Call_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_call_statement; }
	}

	public final Call_statementContext call_statement() throws RecognitionException {
		Call_statementContext _localctx = new Call_statementContext(_ctx, getState());
		enterRule(_localctx, 44, RULE_call_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(295);
			match(KW_CALL);
			setState(296);
			callable_expr();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Return_statementContext extends ParserRuleContext {
		public TerminalNode KW_RETURN() { return getToken(NeuroScriptParser.KW_RETURN, 0); }
		public Expression_listContext expression_list() {
			return getRuleContext(Expression_listContext.class,0);
		}
		public Return_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_return_statement; }
	}

	public final Return_statementContext return_statement() throws RecognitionException {
		Return_statementContext _localctx = new Return_statementContext(_ctx, getState());
		enterRule(_localctx, 46, RULE_return_statement);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(298);
			match(KW_RETURN);
			setState(300);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if ((((_la) & ~0x3f) == 0 && ((1L << _la) & -2614308402075000668L) != 0) || ((((_la - 66)) & ~0x3f) == 0 && ((1L << (_la - 66)) & 283393L) != 0)) {
				{
				setState(299);
				expression_list();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Emit_statementContext extends ParserRuleContext {
		public TerminalNode KW_EMIT() { return getToken(NeuroScriptParser.KW_EMIT, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public Emit_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_emit_statement; }
	}

	public final Emit_statementContext emit_statement() throws RecognitionException {
		Emit_statementContext _localctx = new Emit_statementContext(_ctx, getState());
		enterRule(_localctx, 48, RULE_emit_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(302);
			match(KW_EMIT);
			setState(303);
			expression();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Must_statementContext extends ParserRuleContext {
		public TerminalNode KW_MUST() { return getToken(NeuroScriptParser.KW_MUST, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode KW_MUSTBE() { return getToken(NeuroScriptParser.KW_MUSTBE, 0); }
		public Callable_exprContext callable_expr() {
			return getRuleContext(Callable_exprContext.class,0);
		}
		public Must_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_must_statement; }
	}

	public final Must_statementContext must_statement() throws RecognitionException {
		Must_statementContext _localctx = new Must_statementContext(_ctx, getState());
		enterRule(_localctx, 50, RULE_must_statement);
		try {
			setState(309);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_MUST:
				enterOuterAlt(_localctx, 1);
				{
				setState(305);
				match(KW_MUST);
				setState(306);
				expression();
				}
				break;
			case KW_MUSTBE:
				enterOuterAlt(_localctx, 2);
				{
				setState(307);
				match(KW_MUSTBE);
				setState(308);
				callable_expr();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Fail_statementContext extends ParserRuleContext {
		public TerminalNode KW_FAIL() { return getToken(NeuroScriptParser.KW_FAIL, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public Fail_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_fail_statement; }
	}

	public final Fail_statementContext fail_statement() throws RecognitionException {
		Fail_statementContext _localctx = new Fail_statementContext(_ctx, getState());
		enterRule(_localctx, 52, RULE_fail_statement);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(311);
			match(KW_FAIL);
			setState(313);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if ((((_la) & ~0x3f) == 0 && ((1L << _la) & -2614308402075000668L) != 0) || ((((_la - 66)) & ~0x3f) == 0 && ((1L << (_la - 66)) & 283393L) != 0)) {
				{
				setState(312);
				expression();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ClearErrorStmtContext extends ParserRuleContext {
		public TerminalNode KW_CLEAR_ERROR() { return getToken(NeuroScriptParser.KW_CLEAR_ERROR, 0); }
		public ClearErrorStmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_clearErrorStmt; }
	}

	public final ClearErrorStmtContext clearErrorStmt() throws RecognitionException {
		ClearErrorStmtContext _localctx = new ClearErrorStmtContext(_ctx, getState());
		enterRule(_localctx, 54, RULE_clearErrorStmt);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(315);
			match(KW_CLEAR_ERROR);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Ask_stmtContext extends ParserRuleContext {
		public TerminalNode KW_ASK() { return getToken(NeuroScriptParser.KW_ASK, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode KW_INTO() { return getToken(NeuroScriptParser.KW_INTO, 0); }
		public TerminalNode IDENTIFIER() { return getToken(NeuroScriptParser.IDENTIFIER, 0); }
		public Ask_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_ask_stmt; }
	}

	public final Ask_stmtContext ask_stmt() throws RecognitionException {
		Ask_stmtContext _localctx = new Ask_stmtContext(_ctx, getState());
		enterRule(_localctx, 56, RULE_ask_stmt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(317);
			match(KW_ASK);
			setState(318);
			expression();
			setState(321);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==KW_INTO) {
				{
				setState(319);
				match(KW_INTO);
				setState(320);
				match(IDENTIFIER);
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Break_statementContext extends ParserRuleContext {
		public TerminalNode KW_BREAK() { return getToken(NeuroScriptParser.KW_BREAK, 0); }
		public Break_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_break_statement; }
	}

	public final Break_statementContext break_statement() throws RecognitionException {
		Break_statementContext _localctx = new Break_statementContext(_ctx, getState());
		enterRule(_localctx, 58, RULE_break_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(323);
			match(KW_BREAK);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Continue_statementContext extends ParserRuleContext {
		public TerminalNode KW_CONTINUE() { return getToken(NeuroScriptParser.KW_CONTINUE, 0); }
		public Continue_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_continue_statement; }
	}

	public final Continue_statementContext continue_statement() throws RecognitionException {
		Continue_statementContext _localctx = new Continue_statementContext(_ctx, getState());
		enterRule(_localctx, 60, RULE_continue_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(325);
			match(KW_CONTINUE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class If_statementContext extends ParserRuleContext {
		public TerminalNode KW_IF() { return getToken(NeuroScriptParser.KW_IF, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public List<TerminalNode> NEWLINE() { return getTokens(NeuroScriptParser.NEWLINE); }
		public TerminalNode NEWLINE(int i) {
			return getToken(NeuroScriptParser.NEWLINE, i);
		}
		public List<Statement_listContext> statement_list() {
			return getRuleContexts(Statement_listContext.class);
		}
		public Statement_listContext statement_list(int i) {
			return getRuleContext(Statement_listContext.class,i);
		}
		public TerminalNode KW_ENDIF() { return getToken(NeuroScriptParser.KW_ENDIF, 0); }
		public TerminalNode KW_ELSE() { return getToken(NeuroScriptParser.KW_ELSE, 0); }
		public If_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_if_statement; }
	}

	public final If_statementContext if_statement() throws RecognitionException {
		If_statementContext _localctx = new If_statementContext(_ctx, getState());
		enterRule(_localctx, 62, RULE_if_statement);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(327);
			match(KW_IF);
			setState(328);
			expression();
			setState(329);
			match(NEWLINE);
			setState(330);
			statement_list();
			setState(334);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==KW_ELSE) {
				{
				setState(331);
				match(KW_ELSE);
				setState(332);
				match(NEWLINE);
				setState(333);
				statement_list();
				}
			}

			setState(336);
			match(KW_ENDIF);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class While_statementContext extends ParserRuleContext {
		public TerminalNode KW_WHILE() { return getToken(NeuroScriptParser.KW_WHILE, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode NEWLINE() { return getToken(NeuroScriptParser.NEWLINE, 0); }
		public Statement_listContext statement_list() {
			return getRuleContext(Statement_listContext.class,0);
		}
		public TerminalNode KW_ENDWHILE() { return getToken(NeuroScriptParser.KW_ENDWHILE, 0); }
		public While_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_while_statement; }
	}

	public final While_statementContext while_statement() throws RecognitionException {
		While_statementContext _localctx = new While_statementContext(_ctx, getState());
		enterRule(_localctx, 64, RULE_while_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(338);
			match(KW_WHILE);
			setState(339);
			expression();
			setState(340);
			match(NEWLINE);
			setState(341);
			statement_list();
			setState(342);
			match(KW_ENDWHILE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class For_each_statementContext extends ParserRuleContext {
		public TerminalNode KW_FOR() { return getToken(NeuroScriptParser.KW_FOR, 0); }
		public TerminalNode KW_EACH() { return getToken(NeuroScriptParser.KW_EACH, 0); }
		public TerminalNode IDENTIFIER() { return getToken(NeuroScriptParser.IDENTIFIER, 0); }
		public TerminalNode KW_IN() { return getToken(NeuroScriptParser.KW_IN, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode NEWLINE() { return getToken(NeuroScriptParser.NEWLINE, 0); }
		public Statement_listContext statement_list() {
			return getRuleContext(Statement_listContext.class,0);
		}
		public TerminalNode KW_ENDFOR() { return getToken(NeuroScriptParser.KW_ENDFOR, 0); }
		public For_each_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_for_each_statement; }
	}

	public final For_each_statementContext for_each_statement() throws RecognitionException {
		For_each_statementContext _localctx = new For_each_statementContext(_ctx, getState());
		enterRule(_localctx, 66, RULE_for_each_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(344);
			match(KW_FOR);
			setState(345);
			match(KW_EACH);
			setState(346);
			match(IDENTIFIER);
			setState(347);
			match(KW_IN);
			setState(348);
			expression();
			setState(349);
			match(NEWLINE);
			setState(350);
			statement_list();
			setState(351);
			match(KW_ENDFOR);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Qualified_identifierContext extends ParserRuleContext {
		public List<TerminalNode> IDENTIFIER() { return getTokens(NeuroScriptParser.IDENTIFIER); }
		public TerminalNode IDENTIFIER(int i) {
			return getToken(NeuroScriptParser.IDENTIFIER, i);
		}
		public List<TerminalNode> DOT() { return getTokens(NeuroScriptParser.DOT); }
		public TerminalNode DOT(int i) {
			return getToken(NeuroScriptParser.DOT, i);
		}
		public Qualified_identifierContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_qualified_identifier; }
	}

	public final Qualified_identifierContext qualified_identifier() throws RecognitionException {
		Qualified_identifierContext _localctx = new Qualified_identifierContext(_ctx, getState());
		enterRule(_localctx, 68, RULE_qualified_identifier);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(353);
			match(IDENTIFIER);
			setState(358);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==DOT) {
				{
				{
				setState(354);
				match(DOT);
				setState(355);
				match(IDENTIFIER);
				}
				}
				setState(360);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Call_targetContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(NeuroScriptParser.IDENTIFIER, 0); }
		public TerminalNode KW_TOOL() { return getToken(NeuroScriptParser.KW_TOOL, 0); }
		public TerminalNode DOT() { return getToken(NeuroScriptParser.DOT, 0); }
		public Qualified_identifierContext qualified_identifier() {
			return getRuleContext(Qualified_identifierContext.class,0);
		}
		public Call_targetContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_call_target; }
	}

	public final Call_targetContext call_target() throws RecognitionException {
		Call_targetContext _localctx = new Call_targetContext(_ctx, getState());
		enterRule(_localctx, 70, RULE_call_target);
		try {
			setState(365);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case IDENTIFIER:
				enterOuterAlt(_localctx, 1);
				{
				setState(361);
				match(IDENTIFIER);
				}
				break;
			case KW_TOOL:
				enterOuterAlt(_localctx, 2);
				{
				setState(362);
				match(KW_TOOL);
				setState(363);
				match(DOT);
				setState(364);
				qualified_identifier();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ExpressionContext extends ParserRuleContext {
		public Logical_or_exprContext logical_or_expr() {
			return getRuleContext(Logical_or_exprContext.class,0);
		}
		public ExpressionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_expression; }
	}

	public final ExpressionContext expression() throws RecognitionException {
		ExpressionContext _localctx = new ExpressionContext(_ctx, getState());
		enterRule(_localctx, 72, RULE_expression);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(367);
			logical_or_expr();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Logical_or_exprContext extends ParserRuleContext {
		public List<Logical_and_exprContext> logical_and_expr() {
			return getRuleContexts(Logical_and_exprContext.class);
		}
		public Logical_and_exprContext logical_and_expr(int i) {
			return getRuleContext(Logical_and_exprContext.class,i);
		}
		public List<TerminalNode> KW_OR() { return getTokens(NeuroScriptParser.KW_OR); }
		public TerminalNode KW_OR(int i) {
			return getToken(NeuroScriptParser.KW_OR, i);
		}
		public Logical_or_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_logical_or_expr; }
	}

	public final Logical_or_exprContext logical_or_expr() throws RecognitionException {
		Logical_or_exprContext _localctx = new Logical_or_exprContext(_ctx, getState());
		enterRule(_localctx, 74, RULE_logical_or_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(369);
			logical_and_expr();
			setState(374);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==KW_OR) {
				{
				{
				setState(370);
				match(KW_OR);
				setState(371);
				logical_and_expr();
				}
				}
				setState(376);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Logical_and_exprContext extends ParserRuleContext {
		public List<Bitwise_or_exprContext> bitwise_or_expr() {
			return getRuleContexts(Bitwise_or_exprContext.class);
		}
		public Bitwise_or_exprContext bitwise_or_expr(int i) {
			return getRuleContext(Bitwise_or_exprContext.class,i);
		}
		public List<TerminalNode> KW_AND() { return getTokens(NeuroScriptParser.KW_AND); }
		public TerminalNode KW_AND(int i) {
			return getToken(NeuroScriptParser.KW_AND, i);
		}
		public Logical_and_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_logical_and_expr; }
	}

	public final Logical_and_exprContext logical_and_expr() throws RecognitionException {
		Logical_and_exprContext _localctx = new Logical_and_exprContext(_ctx, getState());
		enterRule(_localctx, 76, RULE_logical_and_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(377);
			bitwise_or_expr();
			setState(382);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==KW_AND) {
				{
				{
				setState(378);
				match(KW_AND);
				setState(379);
				bitwise_or_expr();
				}
				}
				setState(384);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Bitwise_or_exprContext extends ParserRuleContext {
		public List<Bitwise_xor_exprContext> bitwise_xor_expr() {
			return getRuleContexts(Bitwise_xor_exprContext.class);
		}
		public Bitwise_xor_exprContext bitwise_xor_expr(int i) {
			return getRuleContext(Bitwise_xor_exprContext.class,i);
		}
		public List<TerminalNode> PIPE() { return getTokens(NeuroScriptParser.PIPE); }
		public TerminalNode PIPE(int i) {
			return getToken(NeuroScriptParser.PIPE, i);
		}
		public Bitwise_or_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_bitwise_or_expr; }
	}

	public final Bitwise_or_exprContext bitwise_or_expr() throws RecognitionException {
		Bitwise_or_exprContext _localctx = new Bitwise_or_exprContext(_ctx, getState());
		enterRule(_localctx, 78, RULE_bitwise_or_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(385);
			bitwise_xor_expr();
			setState(390);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==PIPE) {
				{
				{
				setState(386);
				match(PIPE);
				setState(387);
				bitwise_xor_expr();
				}
				}
				setState(392);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Bitwise_xor_exprContext extends ParserRuleContext {
		public List<Bitwise_and_exprContext> bitwise_and_expr() {
			return getRuleContexts(Bitwise_and_exprContext.class);
		}
		public Bitwise_and_exprContext bitwise_and_expr(int i) {
			return getRuleContext(Bitwise_and_exprContext.class,i);
		}
		public List<TerminalNode> CARET() { return getTokens(NeuroScriptParser.CARET); }
		public TerminalNode CARET(int i) {
			return getToken(NeuroScriptParser.CARET, i);
		}
		public Bitwise_xor_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_bitwise_xor_expr; }
	}

	public final Bitwise_xor_exprContext bitwise_xor_expr() throws RecognitionException {
		Bitwise_xor_exprContext _localctx = new Bitwise_xor_exprContext(_ctx, getState());
		enterRule(_localctx, 80, RULE_bitwise_xor_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(393);
			bitwise_and_expr();
			setState(398);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==CARET) {
				{
				{
				setState(394);
				match(CARET);
				setState(395);
				bitwise_and_expr();
				}
				}
				setState(400);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Bitwise_and_exprContext extends ParserRuleContext {
		public List<Equality_exprContext> equality_expr() {
			return getRuleContexts(Equality_exprContext.class);
		}
		public Equality_exprContext equality_expr(int i) {
			return getRuleContext(Equality_exprContext.class,i);
		}
		public List<TerminalNode> AMPERSAND() { return getTokens(NeuroScriptParser.AMPERSAND); }
		public TerminalNode AMPERSAND(int i) {
			return getToken(NeuroScriptParser.AMPERSAND, i);
		}
		public Bitwise_and_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_bitwise_and_expr; }
	}

	public final Bitwise_and_exprContext bitwise_and_expr() throws RecognitionException {
		Bitwise_and_exprContext _localctx = new Bitwise_and_exprContext(_ctx, getState());
		enterRule(_localctx, 82, RULE_bitwise_and_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(401);
			equality_expr();
			setState(406);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==AMPERSAND) {
				{
				{
				setState(402);
				match(AMPERSAND);
				setState(403);
				equality_expr();
				}
				}
				setState(408);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Equality_exprContext extends ParserRuleContext {
		public List<Relational_exprContext> relational_expr() {
			return getRuleContexts(Relational_exprContext.class);
		}
		public Relational_exprContext relational_expr(int i) {
			return getRuleContext(Relational_exprContext.class,i);
		}
		public List<TerminalNode> EQ() { return getTokens(NeuroScriptParser.EQ); }
		public TerminalNode EQ(int i) {
			return getToken(NeuroScriptParser.EQ, i);
		}
		public List<TerminalNode> NEQ() { return getTokens(NeuroScriptParser.NEQ); }
		public TerminalNode NEQ(int i) {
			return getToken(NeuroScriptParser.NEQ, i);
		}
		public Equality_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_equality_expr; }
	}

	public final Equality_exprContext equality_expr() throws RecognitionException {
		Equality_exprContext _localctx = new Equality_exprContext(_ctx, getState());
		enterRule(_localctx, 84, RULE_equality_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(409);
			relational_expr();
			setState(414);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==EQ || _la==NEQ) {
				{
				{
				setState(410);
				_la = _input.LA(1);
				if ( !(_la==EQ || _la==NEQ) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(411);
				relational_expr();
				}
				}
				setState(416);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Relational_exprContext extends ParserRuleContext {
		public List<Additive_exprContext> additive_expr() {
			return getRuleContexts(Additive_exprContext.class);
		}
		public Additive_exprContext additive_expr(int i) {
			return getRuleContext(Additive_exprContext.class,i);
		}
		public List<TerminalNode> GT() { return getTokens(NeuroScriptParser.GT); }
		public TerminalNode GT(int i) {
			return getToken(NeuroScriptParser.GT, i);
		}
		public List<TerminalNode> LT() { return getTokens(NeuroScriptParser.LT); }
		public TerminalNode LT(int i) {
			return getToken(NeuroScriptParser.LT, i);
		}
		public List<TerminalNode> GTE() { return getTokens(NeuroScriptParser.GTE); }
		public TerminalNode GTE(int i) {
			return getToken(NeuroScriptParser.GTE, i);
		}
		public List<TerminalNode> LTE() { return getTokens(NeuroScriptParser.LTE); }
		public TerminalNode LTE(int i) {
			return getToken(NeuroScriptParser.LTE, i);
		}
		public Relational_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_relational_expr; }
	}

	public final Relational_exprContext relational_expr() throws RecognitionException {
		Relational_exprContext _localctx = new Relational_exprContext(_ctx, getState());
		enterRule(_localctx, 86, RULE_relational_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(417);
			additive_expr();
			setState(422);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (((((_la - 88)) & ~0x3f) == 0 && ((1L << (_la - 88)) & 15L) != 0)) {
				{
				{
				setState(418);
				_la = _input.LA(1);
				if ( !(((((_la - 88)) & ~0x3f) == 0 && ((1L << (_la - 88)) & 15L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(419);
				additive_expr();
				}
				}
				setState(424);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Additive_exprContext extends ParserRuleContext {
		public List<Multiplicative_exprContext> multiplicative_expr() {
			return getRuleContexts(Multiplicative_exprContext.class);
		}
		public Multiplicative_exprContext multiplicative_expr(int i) {
			return getRuleContext(Multiplicative_exprContext.class,i);
		}
		public List<TerminalNode> PLUS() { return getTokens(NeuroScriptParser.PLUS); }
		public TerminalNode PLUS(int i) {
			return getToken(NeuroScriptParser.PLUS, i);
		}
		public List<TerminalNode> MINUS() { return getTokens(NeuroScriptParser.MINUS); }
		public TerminalNode MINUS(int i) {
			return getToken(NeuroScriptParser.MINUS, i);
		}
		public Additive_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_additive_expr; }
	}

	public final Additive_exprContext additive_expr() throws RecognitionException {
		Additive_exprContext _localctx = new Additive_exprContext(_ctx, getState());
		enterRule(_localctx, 88, RULE_additive_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(425);
			multiplicative_expr();
			setState(430);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==PLUS || _la==MINUS) {
				{
				{
				setState(426);
				_la = _input.LA(1);
				if ( !(_la==PLUS || _la==MINUS) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(427);
				multiplicative_expr();
				}
				}
				setState(432);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Multiplicative_exprContext extends ParserRuleContext {
		public List<Unary_exprContext> unary_expr() {
			return getRuleContexts(Unary_exprContext.class);
		}
		public Unary_exprContext unary_expr(int i) {
			return getRuleContext(Unary_exprContext.class,i);
		}
		public List<TerminalNode> STAR() { return getTokens(NeuroScriptParser.STAR); }
		public TerminalNode STAR(int i) {
			return getToken(NeuroScriptParser.STAR, i);
		}
		public List<TerminalNode> SLASH() { return getTokens(NeuroScriptParser.SLASH); }
		public TerminalNode SLASH(int i) {
			return getToken(NeuroScriptParser.SLASH, i);
		}
		public List<TerminalNode> PERCENT() { return getTokens(NeuroScriptParser.PERCENT); }
		public TerminalNode PERCENT(int i) {
			return getToken(NeuroScriptParser.PERCENT, i);
		}
		public Multiplicative_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_multiplicative_expr; }
	}

	public final Multiplicative_exprContext multiplicative_expr() throws RecognitionException {
		Multiplicative_exprContext _localctx = new Multiplicative_exprContext(_ctx, getState());
		enterRule(_localctx, 90, RULE_multiplicative_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(433);
			unary_expr();
			setState(438);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (((((_la - 67)) & ~0x3f) == 0 && ((1L << (_la - 67)) & 7L) != 0)) {
				{
				{
				setState(434);
				_la = _input.LA(1);
				if ( !(((((_la - 67)) & ~0x3f) == 0 && ((1L << (_la - 67)) & 7L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(435);
				unary_expr();
				}
				}
				setState(440);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Unary_exprContext extends ParserRuleContext {
		public Unary_exprContext unary_expr() {
			return getRuleContext(Unary_exprContext.class,0);
		}
		public TerminalNode MINUS() { return getToken(NeuroScriptParser.MINUS, 0); }
		public TerminalNode KW_NOT() { return getToken(NeuroScriptParser.KW_NOT, 0); }
		public TerminalNode KW_NO() { return getToken(NeuroScriptParser.KW_NO, 0); }
		public TerminalNode KW_SOME() { return getToken(NeuroScriptParser.KW_SOME, 0); }
		public TerminalNode TILDE() { return getToken(NeuroScriptParser.TILDE, 0); }
		public TerminalNode KW_MUST() { return getToken(NeuroScriptParser.KW_MUST, 0); }
		public TerminalNode KW_TYPEOF() { return getToken(NeuroScriptParser.KW_TYPEOF, 0); }
		public Power_exprContext power_expr() {
			return getRuleContext(Power_exprContext.class,0);
		}
		public Unary_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_unary_expr; }
	}

	public final Unary_exprContext unary_expr() throws RecognitionException {
		Unary_exprContext _localctx = new Unary_exprContext(_ctx, getState());
		enterRule(_localctx, 92, RULE_unary_expr);
		int _la;
		try {
			setState(446);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_MUST:
			case KW_NO:
			case KW_NOT:
			case KW_SOME:
			case MINUS:
			case TILDE:
				enterOuterAlt(_localctx, 1);
				{
				setState(441);
				_la = _input.LA(1);
				if ( !(((((_la - 38)) & ~0x3f) == 0 && ((1L << (_la - 38)) & 68987928673L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(442);
				unary_expr();
				}
				break;
			case KW_TYPEOF:
				enterOuterAlt(_localctx, 2);
				{
				setState(443);
				match(KW_TYPEOF);
				setState(444);
				unary_expr();
				}
				break;
			case KW_ACOS:
			case KW_ASIN:
			case KW_ATAN:
			case KW_COS:
			case KW_EVAL:
			case KW_FALSE:
			case KW_LAST:
			case KW_LN:
			case KW_LOG:
			case KW_NIL:
			case KW_SIN:
			case KW_TAN:
			case KW_TOOL:
			case KW_TRUE:
			case STRING_LIT:
			case TRIPLE_BACKTICK_STRING:
			case NUMBER_LIT:
			case IDENTIFIER:
			case LPAREN:
			case LBRACK:
			case LBRACE:
			case PLACEHOLDER_START:
				enterOuterAlt(_localctx, 3);
				{
				setState(445);
				power_expr();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Power_exprContext extends ParserRuleContext {
		public Accessor_exprContext accessor_expr() {
			return getRuleContext(Accessor_exprContext.class,0);
		}
		public TerminalNode STAR_STAR() { return getToken(NeuroScriptParser.STAR_STAR, 0); }
		public Power_exprContext power_expr() {
			return getRuleContext(Power_exprContext.class,0);
		}
		public Power_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_power_expr; }
	}

	public final Power_exprContext power_expr() throws RecognitionException {
		Power_exprContext _localctx = new Power_exprContext(_ctx, getState());
		enterRule(_localctx, 94, RULE_power_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(448);
			accessor_expr();
			setState(451);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==STAR_STAR) {
				{
				setState(449);
				match(STAR_STAR);
				setState(450);
				power_expr();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Accessor_exprContext extends ParserRuleContext {
		public PrimaryContext primary() {
			return getRuleContext(PrimaryContext.class,0);
		}
		public List<TerminalNode> LBRACK() { return getTokens(NeuroScriptParser.LBRACK); }
		public TerminalNode LBRACK(int i) {
			return getToken(NeuroScriptParser.LBRACK, i);
		}
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public List<TerminalNode> RBRACK() { return getTokens(NeuroScriptParser.RBRACK); }
		public TerminalNode RBRACK(int i) {
			return getToken(NeuroScriptParser.RBRACK, i);
		}
		public Accessor_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_accessor_expr; }
	}

	public final Accessor_exprContext accessor_expr() throws RecognitionException {
		Accessor_exprContext _localctx = new Accessor_exprContext(_ctx, getState());
		enterRule(_localctx, 96, RULE_accessor_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(453);
			primary();
			setState(460);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==LBRACK) {
				{
				{
				setState(454);
				match(LBRACK);
				setState(455);
				expression();
				setState(456);
				match(RBRACK);
				}
				}
				setState(462);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class PrimaryContext extends ParserRuleContext {
		public LiteralContext literal() {
			return getRuleContext(LiteralContext.class,0);
		}
		public PlaceholderContext placeholder() {
			return getRuleContext(PlaceholderContext.class,0);
		}
		public TerminalNode IDENTIFIER() { return getToken(NeuroScriptParser.IDENTIFIER, 0); }
		public TerminalNode KW_LAST() { return getToken(NeuroScriptParser.KW_LAST, 0); }
		public Callable_exprContext callable_expr() {
			return getRuleContext(Callable_exprContext.class,0);
		}
		public TerminalNode KW_EVAL() { return getToken(NeuroScriptParser.KW_EVAL, 0); }
		public TerminalNode LPAREN() { return getToken(NeuroScriptParser.LPAREN, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode RPAREN() { return getToken(NeuroScriptParser.RPAREN, 0); }
		public PrimaryContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_primary; }
	}

	public final PrimaryContext primary() throws RecognitionException {
		PrimaryContext _localctx = new PrimaryContext(_ctx, getState());
		enterRule(_localctx, 98, RULE_primary);
		try {
			setState(477);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,42,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(463);
				literal();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(464);
				placeholder();
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(465);
				match(IDENTIFIER);
				}
				break;
			case 4:
				enterOuterAlt(_localctx, 4);
				{
				setState(466);
				match(KW_LAST);
				}
				break;
			case 5:
				enterOuterAlt(_localctx, 5);
				{
				setState(467);
				callable_expr();
				}
				break;
			case 6:
				enterOuterAlt(_localctx, 6);
				{
				setState(468);
				match(KW_EVAL);
				setState(469);
				match(LPAREN);
				setState(470);
				expression();
				setState(471);
				match(RPAREN);
				}
				break;
			case 7:
				enterOuterAlt(_localctx, 7);
				{
				setState(473);
				match(LPAREN);
				setState(474);
				expression();
				setState(475);
				match(RPAREN);
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Callable_exprContext extends ParserRuleContext {
		public TerminalNode LPAREN() { return getToken(NeuroScriptParser.LPAREN, 0); }
		public Expression_list_optContext expression_list_opt() {
			return getRuleContext(Expression_list_optContext.class,0);
		}
		public TerminalNode RPAREN() { return getToken(NeuroScriptParser.RPAREN, 0); }
		public Call_targetContext call_target() {
			return getRuleContext(Call_targetContext.class,0);
		}
		public TerminalNode KW_LN() { return getToken(NeuroScriptParser.KW_LN, 0); }
		public TerminalNode KW_LOG() { return getToken(NeuroScriptParser.KW_LOG, 0); }
		public TerminalNode KW_SIN() { return getToken(NeuroScriptParser.KW_SIN, 0); }
		public TerminalNode KW_COS() { return getToken(NeuroScriptParser.KW_COS, 0); }
		public TerminalNode KW_TAN() { return getToken(NeuroScriptParser.KW_TAN, 0); }
		public TerminalNode KW_ASIN() { return getToken(NeuroScriptParser.KW_ASIN, 0); }
		public TerminalNode KW_ACOS() { return getToken(NeuroScriptParser.KW_ACOS, 0); }
		public TerminalNode KW_ATAN() { return getToken(NeuroScriptParser.KW_ATAN, 0); }
		public Callable_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_callable_expr; }
	}

	public final Callable_exprContext callable_expr() throws RecognitionException {
		Callable_exprContext _localctx = new Callable_exprContext(_ctx, getState());
		enterRule(_localctx, 100, RULE_callable_expr);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(488);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_TOOL:
			case IDENTIFIER:
				{
				setState(479);
				call_target();
				}
				break;
			case KW_LN:
				{
				setState(480);
				match(KW_LN);
				}
				break;
			case KW_LOG:
				{
				setState(481);
				match(KW_LOG);
				}
				break;
			case KW_SIN:
				{
				setState(482);
				match(KW_SIN);
				}
				break;
			case KW_COS:
				{
				setState(483);
				match(KW_COS);
				}
				break;
			case KW_TAN:
				{
				setState(484);
				match(KW_TAN);
				}
				break;
			case KW_ASIN:
				{
				setState(485);
				match(KW_ASIN);
				}
				break;
			case KW_ACOS:
				{
				setState(486);
				match(KW_ACOS);
				}
				break;
			case KW_ATAN:
				{
				setState(487);
				match(KW_ATAN);
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			setState(490);
			match(LPAREN);
			setState(491);
			expression_list_opt();
			setState(492);
			match(RPAREN);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class PlaceholderContext extends ParserRuleContext {
		public TerminalNode PLACEHOLDER_START() { return getToken(NeuroScriptParser.PLACEHOLDER_START, 0); }
		public TerminalNode PLACEHOLDER_END() { return getToken(NeuroScriptParser.PLACEHOLDER_END, 0); }
		public TerminalNode IDENTIFIER() { return getToken(NeuroScriptParser.IDENTIFIER, 0); }
		public TerminalNode KW_LAST() { return getToken(NeuroScriptParser.KW_LAST, 0); }
		public PlaceholderContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_placeholder; }
	}

	public final PlaceholderContext placeholder() throws RecognitionException {
		PlaceholderContext _localctx = new PlaceholderContext(_ctx, getState());
		enterRule(_localctx, 102, RULE_placeholder);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(494);
			match(PLACEHOLDER_START);
			setState(495);
			_la = _input.LA(1);
			if ( !(_la==KW_LAST || _la==IDENTIFIER) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			setState(496);
			match(PLACEHOLDER_END);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class LiteralContext extends ParserRuleContext {
		public TerminalNode STRING_LIT() { return getToken(NeuroScriptParser.STRING_LIT, 0); }
		public TerminalNode TRIPLE_BACKTICK_STRING() { return getToken(NeuroScriptParser.TRIPLE_BACKTICK_STRING, 0); }
		public TerminalNode NUMBER_LIT() { return getToken(NeuroScriptParser.NUMBER_LIT, 0); }
		public List_literalContext list_literal() {
			return getRuleContext(List_literalContext.class,0);
		}
		public Map_literalContext map_literal() {
			return getRuleContext(Map_literalContext.class,0);
		}
		public Boolean_literalContext boolean_literal() {
			return getRuleContext(Boolean_literalContext.class,0);
		}
		public Nil_literalContext nil_literal() {
			return getRuleContext(Nil_literalContext.class,0);
		}
		public LiteralContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_literal; }
	}

	public final LiteralContext literal() throws RecognitionException {
		LiteralContext _localctx = new LiteralContext(_ctx, getState());
		enterRule(_localctx, 104, RULE_literal);
		try {
			setState(505);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case STRING_LIT:
				enterOuterAlt(_localctx, 1);
				{
				setState(498);
				match(STRING_LIT);
				}
				break;
			case TRIPLE_BACKTICK_STRING:
				enterOuterAlt(_localctx, 2);
				{
				setState(499);
				match(TRIPLE_BACKTICK_STRING);
				}
				break;
			case NUMBER_LIT:
				enterOuterAlt(_localctx, 3);
				{
				setState(500);
				match(NUMBER_LIT);
				}
				break;
			case LBRACK:
				enterOuterAlt(_localctx, 4);
				{
				setState(501);
				list_literal();
				}
				break;
			case LBRACE:
				enterOuterAlt(_localctx, 5);
				{
				setState(502);
				map_literal();
				}
				break;
			case KW_FALSE:
			case KW_TRUE:
				enterOuterAlt(_localctx, 6);
				{
				setState(503);
				boolean_literal();
				}
				break;
			case KW_NIL:
				enterOuterAlt(_localctx, 7);
				{
				setState(504);
				nil_literal();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Nil_literalContext extends ParserRuleContext {
		public TerminalNode KW_NIL() { return getToken(NeuroScriptParser.KW_NIL, 0); }
		public Nil_literalContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_nil_literal; }
	}

	public final Nil_literalContext nil_literal() throws RecognitionException {
		Nil_literalContext _localctx = new Nil_literalContext(_ctx, getState());
		enterRule(_localctx, 106, RULE_nil_literal);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(507);
			match(KW_NIL);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Boolean_literalContext extends ParserRuleContext {
		public TerminalNode KW_TRUE() { return getToken(NeuroScriptParser.KW_TRUE, 0); }
		public TerminalNode KW_FALSE() { return getToken(NeuroScriptParser.KW_FALSE, 0); }
		public Boolean_literalContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_boolean_literal; }
	}

	public final Boolean_literalContext boolean_literal() throws RecognitionException {
		Boolean_literalContext _localctx = new Boolean_literalContext(_ctx, getState());
		enterRule(_localctx, 108, RULE_boolean_literal);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(509);
			_la = _input.LA(1);
			if ( !(_la==KW_FALSE || _la==KW_TRUE) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class List_literalContext extends ParserRuleContext {
		public TerminalNode LBRACK() { return getToken(NeuroScriptParser.LBRACK, 0); }
		public Expression_list_optContext expression_list_opt() {
			return getRuleContext(Expression_list_optContext.class,0);
		}
		public TerminalNode RBRACK() { return getToken(NeuroScriptParser.RBRACK, 0); }
		public List_literalContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_list_literal; }
	}

	public final List_literalContext list_literal() throws RecognitionException {
		List_literalContext _localctx = new List_literalContext(_ctx, getState());
		enterRule(_localctx, 110, RULE_list_literal);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(511);
			match(LBRACK);
			setState(512);
			expression_list_opt();
			setState(513);
			match(RBRACK);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Map_literalContext extends ParserRuleContext {
		public TerminalNode LBRACE() { return getToken(NeuroScriptParser.LBRACE, 0); }
		public Map_entry_list_optContext map_entry_list_opt() {
			return getRuleContext(Map_entry_list_optContext.class,0);
		}
		public TerminalNode RBRACE() { return getToken(NeuroScriptParser.RBRACE, 0); }
		public Map_literalContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_map_literal; }
	}

	public final Map_literalContext map_literal() throws RecognitionException {
		Map_literalContext _localctx = new Map_literalContext(_ctx, getState());
		enterRule(_localctx, 112, RULE_map_literal);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(515);
			match(LBRACE);
			setState(516);
			map_entry_list_opt();
			setState(517);
			match(RBRACE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Expression_list_optContext extends ParserRuleContext {
		public Expression_listContext expression_list() {
			return getRuleContext(Expression_listContext.class,0);
		}
		public Expression_list_optContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_expression_list_opt; }
	}

	public final Expression_list_optContext expression_list_opt() throws RecognitionException {
		Expression_list_optContext _localctx = new Expression_list_optContext(_ctx, getState());
		enterRule(_localctx, 114, RULE_expression_list_opt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(520);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if ((((_la) & ~0x3f) == 0 && ((1L << _la) & -2614308402075000668L) != 0) || ((((_la - 66)) & ~0x3f) == 0 && ((1L << (_la - 66)) & 283393L) != 0)) {
				{
				setState(519);
				expression_list();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Expression_listContext extends ParserRuleContext {
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(NeuroScriptParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(NeuroScriptParser.COMMA, i);
		}
		public Expression_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_expression_list; }
	}

	public final Expression_listContext expression_list() throws RecognitionException {
		Expression_listContext _localctx = new Expression_listContext(_ctx, getState());
		enterRule(_localctx, 116, RULE_expression_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(522);
			expression();
			setState(527);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(523);
				match(COMMA);
				setState(524);
				expression();
				}
				}
				setState(529);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Map_entry_list_optContext extends ParserRuleContext {
		public Map_entry_listContext map_entry_list() {
			return getRuleContext(Map_entry_listContext.class,0);
		}
		public Map_entry_list_optContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_map_entry_list_opt; }
	}

	public final Map_entry_list_optContext map_entry_list_opt() throws RecognitionException {
		Map_entry_list_optContext _localctx = new Map_entry_list_optContext(_ctx, getState());
		enterRule(_localctx, 118, RULE_map_entry_list_opt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(531);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==STRING_LIT) {
				{
				setState(530);
				map_entry_list();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Map_entry_listContext extends ParserRuleContext {
		public List<Map_entryContext> map_entry() {
			return getRuleContexts(Map_entryContext.class);
		}
		public Map_entryContext map_entry(int i) {
			return getRuleContext(Map_entryContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(NeuroScriptParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(NeuroScriptParser.COMMA, i);
		}
		public Map_entry_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_map_entry_list; }
	}

	public final Map_entry_listContext map_entry_list() throws RecognitionException {
		Map_entry_listContext _localctx = new Map_entry_listContext(_ctx, getState());
		enterRule(_localctx, 120, RULE_map_entry_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(533);
			map_entry();
			setState(538);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(534);
				match(COMMA);
				setState(535);
				map_entry();
				}
				}
				setState(540);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Map_entryContext extends ParserRuleContext {
		public TerminalNode STRING_LIT() { return getToken(NeuroScriptParser.STRING_LIT, 0); }
		public TerminalNode COLON() { return getToken(NeuroScriptParser.COLON, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public Map_entryContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_map_entry; }
	}

	public final Map_entryContext map_entry() throws RecognitionException {
		Map_entryContext _localctx = new Map_entryContext(_ctx, getState());
		enterRule(_localctx, 122, RULE_map_entry);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(541);
			match(STRING_LIT);
			setState(542);
			match(COLON);
			setState(543);
			expression();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	public static final String _serializedATN =
		"\u0004\u0001^\u0222\u0002\u0000\u0007\u0000\u0002\u0001\u0007\u0001\u0002"+
		"\u0002\u0007\u0002\u0002\u0003\u0007\u0003\u0002\u0004\u0007\u0004\u0002"+
		"\u0005\u0007\u0005\u0002\u0006\u0007\u0006\u0002\u0007\u0007\u0007\u0002"+
		"\b\u0007\b\u0002\t\u0007\t\u0002\n\u0007\n\u0002\u000b\u0007\u000b\u0002"+
		"\f\u0007\f\u0002\r\u0007\r\u0002\u000e\u0007\u000e\u0002\u000f\u0007\u000f"+
		"\u0002\u0010\u0007\u0010\u0002\u0011\u0007\u0011\u0002\u0012\u0007\u0012"+
		"\u0002\u0013\u0007\u0013\u0002\u0014\u0007\u0014\u0002\u0015\u0007\u0015"+
		"\u0002\u0016\u0007\u0016\u0002\u0017\u0007\u0017\u0002\u0018\u0007\u0018"+
		"\u0002\u0019\u0007\u0019\u0002\u001a\u0007\u001a\u0002\u001b\u0007\u001b"+
		"\u0002\u001c\u0007\u001c\u0002\u001d\u0007\u001d\u0002\u001e\u0007\u001e"+
		"\u0002\u001f\u0007\u001f\u0002 \u0007 \u0002!\u0007!\u0002\"\u0007\"\u0002"+
		"#\u0007#\u0002$\u0007$\u0002%\u0007%\u0002&\u0007&\u0002\'\u0007\'\u0002"+
		"(\u0007(\u0002)\u0007)\u0002*\u0007*\u0002+\u0007+\u0002,\u0007,\u0002"+
		"-\u0007-\u0002.\u0007.\u0002/\u0007/\u00020\u00070\u00021\u00071\u0002"+
		"2\u00072\u00023\u00073\u00024\u00074\u00025\u00075\u00026\u00076\u0002"+
		"7\u00077\u00028\u00078\u00029\u00079\u0002:\u0007:\u0002;\u0007;\u0002"+
		"<\u0007<\u0002=\u0007=\u0001\u0000\u0001\u0000\u0005\u0000\u007f\b\u0000"+
		"\n\u0000\f\u0000\u0082\t\u0000\u0001\u0000\u0001\u0000\u0001\u0001\u0005"+
		"\u0001\u0087\b\u0001\n\u0001\f\u0001\u008a\t\u0001\u0001\u0002\u0001\u0002"+
		"\u0003\u0002\u008e\b\u0002\u0001\u0002\u0005\u0002\u0091\b\u0002\n\u0002"+
		"\f\u0002\u0094\t\u0002\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003"+
		"\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0003\u0001\u0004"+
		"\u0001\u0004\u0001\u0004\u0001\u0004\u0005\u0004\u00a3\b\u0004\n\u0004"+
		"\f\u0004\u00a6\t\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004"+
		"\u0004\u0004\u00ac\b\u0004\u000b\u0004\f\u0004\u00ad\u0001\u0004\u0003"+
		"\u0004\u00b1\b\u0004\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0006\u0001"+
		"\u0006\u0001\u0006\u0001\u0007\u0001\u0007\u0001\u0007\u0001\b\u0001\b"+
		"\u0001\b\u0005\b\u00bf\b\b\n\b\f\b\u00c2\t\b\u0001\t\u0001\t\u0005\t\u00c6"+
		"\b\t\n\t\f\t\u00c9\t\t\u0001\n\u0005\n\u00cc\b\n\n\n\f\n\u00cf\t\n\u0001"+
		"\u000b\u0001\u000b\u0001\u000b\u0001\u000b\u0003\u000b\u00d5\b\u000b\u0001"+
		"\f\u0001\f\u0001\f\u0003\f\u00da\b\f\u0001\r\u0001\r\u0001\r\u0001\r\u0001"+
		"\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0003\r\u00e7\b\r\u0001"+
		"\u000e\u0001\u000e\u0001\u000e\u0003\u000e\u00ec\b\u000e\u0001\u000f\u0001"+
		"\u000f\u0001\u000f\u0003\u000f\u00f1\b\u000f\u0001\u0010\u0001\u0010\u0001"+
		"\u0010\u0001\u0010\u0001\u0010\u0001\u0010\u0001\u0011\u0001\u0011\u0001"+
		"\u0011\u0001\u0011\u0003\u0011\u00fd\b\u0011\u0001\u0011\u0001\u0011\u0003"+
		"\u0011\u0101\b\u0011\u0001\u0011\u0001\u0011\u0001\u0011\u0001\u0011\u0001"+
		"\u0011\u0001\u0012\u0001\u0012\u0001\u0012\u0001\u0012\u0001\u0012\u0003"+
		"\u0012\u010d\b\u0012\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001"+
		"\u0013\u0001\u0013\u0001\u0013\u0005\u0013\u0116\b\u0013\n\u0013\f\u0013"+
		"\u0119\t\u0013\u0001\u0014\u0001\u0014\u0001\u0014\u0005\u0014\u011e\b"+
		"\u0014\n\u0014\f\u0014\u0121\t\u0014\u0001\u0015\u0001\u0015\u0001\u0015"+
		"\u0001\u0015\u0001\u0015\u0001\u0016\u0001\u0016\u0001\u0016\u0001\u0017"+
		"\u0001\u0017\u0003\u0017\u012d\b\u0017\u0001\u0018\u0001\u0018\u0001\u0018"+
		"\u0001\u0019\u0001\u0019\u0001\u0019\u0001\u0019\u0003\u0019\u0136\b\u0019"+
		"\u0001\u001a\u0001\u001a\u0003\u001a\u013a\b\u001a\u0001\u001b\u0001\u001b"+
		"\u0001\u001c\u0001\u001c\u0001\u001c\u0001\u001c\u0003\u001c\u0142\b\u001c"+
		"\u0001\u001d\u0001\u001d\u0001\u001e\u0001\u001e\u0001\u001f\u0001\u001f"+
		"\u0001\u001f\u0001\u001f\u0001\u001f\u0001\u001f\u0001\u001f\u0003\u001f"+
		"\u014f\b\u001f\u0001\u001f\u0001\u001f\u0001 \u0001 \u0001 \u0001 \u0001"+
		" \u0001 \u0001!\u0001!\u0001!\u0001!\u0001!\u0001!\u0001!\u0001!\u0001"+
		"!\u0001\"\u0001\"\u0001\"\u0005\"\u0165\b\"\n\"\f\"\u0168\t\"\u0001#\u0001"+
		"#\u0001#\u0001#\u0003#\u016e\b#\u0001$\u0001$\u0001%\u0001%\u0001%\u0005"+
		"%\u0175\b%\n%\f%\u0178\t%\u0001&\u0001&\u0001&\u0005&\u017d\b&\n&\f&\u0180"+
		"\t&\u0001\'\u0001\'\u0001\'\u0005\'\u0185\b\'\n\'\f\'\u0188\t\'\u0001"+
		"(\u0001(\u0001(\u0005(\u018d\b(\n(\f(\u0190\t(\u0001)\u0001)\u0001)\u0005"+
		")\u0195\b)\n)\f)\u0198\t)\u0001*\u0001*\u0001*\u0005*\u019d\b*\n*\f*\u01a0"+
		"\t*\u0001+\u0001+\u0001+\u0005+\u01a5\b+\n+\f+\u01a8\t+\u0001,\u0001,"+
		"\u0001,\u0005,\u01ad\b,\n,\f,\u01b0\t,\u0001-\u0001-\u0001-\u0005-\u01b5"+
		"\b-\n-\f-\u01b8\t-\u0001.\u0001.\u0001.\u0001.\u0001.\u0003.\u01bf\b."+
		"\u0001/\u0001/\u0001/\u0003/\u01c4\b/\u00010\u00010\u00010\u00010\u0001"+
		"0\u00050\u01cb\b0\n0\f0\u01ce\t0\u00011\u00011\u00011\u00011\u00011\u0001"+
		"1\u00011\u00011\u00011\u00011\u00011\u00011\u00011\u00011\u00031\u01de"+
		"\b1\u00012\u00012\u00012\u00012\u00012\u00012\u00012\u00012\u00012\u0003"+
		"2\u01e9\b2\u00012\u00012\u00012\u00012\u00013\u00013\u00013\u00013\u0001"+
		"4\u00014\u00014\u00014\u00014\u00014\u00014\u00034\u01fa\b4\u00015\u0001"+
		"5\u00016\u00016\u00017\u00017\u00017\u00017\u00018\u00018\u00018\u0001"+
		"8\u00019\u00039\u0209\b9\u0001:\u0001:\u0001:\u0005:\u020e\b:\n:\f:\u0211"+
		"\t:\u0001;\u0003;\u0214\b;\u0001<\u0001<\u0001<\u0005<\u0219\b<\n<\f<"+
		"\u021c\t<\u0001=\u0001=\u0001=\u0001=\u0001=\u0000\u0000>\u0000\u0002"+
		"\u0004\u0006\b\n\f\u000e\u0010\u0012\u0014\u0016\u0018\u001a\u001c\u001e"+
		" \"$&(*,.02468:<>@BDFHJLNPRTVXZ\\^`bdfhjlnprtvxz\u0000\b\u0002\u0000="+
		"=]]\u0001\u0000VW\u0001\u0000X[\u0001\u0000AB\u0001\u0000CE\u0005\u0000"+
		"&&+,44BBJJ\u0002\u0000\"\"??\u0002\u0000\u001b\u001b88\u0234\u0000|\u0001"+
		"\u0000\u0000\u0000\u0002\u0088\u0001\u0000\u0000\u0000\u0004\u008d\u0001"+
		"\u0000\u0000\u0000\u0006\u0095\u0001\u0000\u0000\u0000\b\u00b0\u0001\u0000"+
		"\u0000\u0000\n\u00b2\u0001\u0000\u0000\u0000\f\u00b5\u0001\u0000\u0000"+
		"\u0000\u000e\u00b8\u0001\u0000\u0000\u0000\u0010\u00bb\u0001\u0000\u0000"+
		"\u0000\u0012\u00c7\u0001\u0000\u0000\u0000\u0014\u00cd\u0001\u0000\u0000"+
		"\u0000\u0016\u00d4\u0001\u0000\u0000\u0000\u0018\u00d9\u0001\u0000\u0000"+
		"\u0000\u001a\u00e6\u0001\u0000\u0000\u0000\u001c\u00eb\u0001\u0000\u0000"+
		"\u0000\u001e\u00ed\u0001\u0000\u0000\u0000 \u00f2\u0001\u0000\u0000\u0000"+
		"\"\u00f8\u0001\u0000\u0000\u0000$\u0107\u0001\u0000\u0000\u0000&\u010e"+
		"\u0001\u0000\u0000\u0000(\u011a\u0001\u0000\u0000\u0000*\u0122\u0001\u0000"+
		"\u0000\u0000,\u0127\u0001\u0000\u0000\u0000.\u012a\u0001\u0000\u0000\u0000"+
		"0\u012e\u0001\u0000\u0000\u00002\u0135\u0001\u0000\u0000\u00004\u0137"+
		"\u0001\u0000\u0000\u00006\u013b\u0001\u0000\u0000\u00008\u013d\u0001\u0000"+
		"\u0000\u0000:\u0143\u0001\u0000\u0000\u0000<\u0145\u0001\u0000\u0000\u0000"+
		">\u0147\u0001\u0000\u0000\u0000@\u0152\u0001\u0000\u0000\u0000B\u0158"+
		"\u0001\u0000\u0000\u0000D\u0161\u0001\u0000\u0000\u0000F\u016d\u0001\u0000"+
		"\u0000\u0000H\u016f\u0001\u0000\u0000\u0000J\u0171\u0001\u0000\u0000\u0000"+
		"L\u0179\u0001\u0000\u0000\u0000N\u0181\u0001\u0000\u0000\u0000P\u0189"+
		"\u0001\u0000\u0000\u0000R\u0191\u0001\u0000\u0000\u0000T\u0199\u0001\u0000"+
		"\u0000\u0000V\u01a1\u0001\u0000\u0000\u0000X\u01a9\u0001\u0000\u0000\u0000"+
		"Z\u01b1\u0001\u0000\u0000\u0000\\\u01be\u0001\u0000\u0000\u0000^\u01c0"+
		"\u0001\u0000\u0000\u0000`\u01c5\u0001\u0000\u0000\u0000b\u01dd\u0001\u0000"+
		"\u0000\u0000d\u01e8\u0001\u0000\u0000\u0000f\u01ee\u0001\u0000\u0000\u0000"+
		"h\u01f9\u0001\u0000\u0000\u0000j\u01fb\u0001\u0000\u0000\u0000l\u01fd"+
		"\u0001\u0000\u0000\u0000n\u01ff\u0001\u0000\u0000\u0000p\u0203\u0001\u0000"+
		"\u0000\u0000r\u0208\u0001\u0000\u0000\u0000t\u020a\u0001\u0000\u0000\u0000"+
		"v\u0213\u0001\u0000\u0000\u0000x\u0215\u0001\u0000\u0000\u0000z\u021d"+
		"\u0001\u0000\u0000\u0000|\u0080\u0003\u0002\u0001\u0000}\u007f\u0003\u0004"+
		"\u0002\u0000~}\u0001\u0000\u0000\u0000\u007f\u0082\u0001\u0000\u0000\u0000"+
		"\u0080~\u0001\u0000\u0000\u0000\u0080\u0081\u0001\u0000\u0000\u0000\u0081"+
		"\u0083\u0001\u0000\u0000\u0000\u0082\u0080\u0001\u0000\u0000\u0000\u0083"+
		"\u0084\u0005\u0000\u0000\u0001\u0084\u0001\u0001\u0000\u0000\u0000\u0085"+
		"\u0087\u0007\u0000\u0000\u0000\u0086\u0085\u0001\u0000\u0000\u0000\u0087"+
		"\u008a\u0001\u0000\u0000\u0000\u0088\u0086\u0001\u0000\u0000\u0000\u0088"+
		"\u0089\u0001\u0000\u0000\u0000\u0089\u0003\u0001\u0000\u0000\u0000\u008a"+
		"\u0088\u0001\u0000\u0000\u0000\u008b\u008e\u0003\u0006\u0003\u0000\u008c"+
		"\u008e\u0003\u001e\u000f\u0000\u008d\u008b\u0001\u0000\u0000\u0000\u008d"+
		"\u008c\u0001\u0000\u0000\u0000\u008e\u0092\u0001\u0000\u0000\u0000\u008f"+
		"\u0091\u0005]\u0000\u0000\u0090\u008f\u0001\u0000\u0000\u0000\u0091\u0094"+
		"\u0001\u0000\u0000\u0000\u0092\u0090\u0001\u0000\u0000\u0000\u0092\u0093"+
		"\u0001\u0000\u0000\u0000\u0093\u0005\u0001\u0000\u0000\u0000\u0094\u0092"+
		"\u0001\u0000\u0000\u0000\u0095\u0096\u0005\u001d\u0000\u0000\u0096\u0097"+
		"\u0005?\u0000\u0000\u0097\u0098\u0003\b\u0004\u0000\u0098\u0099\u0005"+
		"%\u0000\u0000\u0099\u009a\u0005]\u0000\u0000\u009a\u009b\u0003\u0012\t"+
		"\u0000\u009b\u009c\u0003\u0014\n\u0000\u009c\u009d\u0005\u0013\u0000\u0000"+
		"\u009d\u0007\u0001\u0000\u0000\u0000\u009e\u00a4\u0005K\u0000\u0000\u009f"+
		"\u00a3\u0003\n\u0005\u0000\u00a0\u00a3\u0003\f\u0006\u0000\u00a1\u00a3"+
		"\u0003\u000e\u0007\u0000\u00a2\u009f\u0001\u0000\u0000\u0000\u00a2\u00a0"+
		"\u0001\u0000\u0000\u0000\u00a2\u00a1\u0001\u0000\u0000\u0000\u00a3\u00a6"+
		"\u0001\u0000\u0000\u0000\u00a4\u00a2\u0001\u0000\u0000\u0000\u00a4\u00a5"+
		"\u0001\u0000\u0000\u0000\u00a5\u00a7\u0001\u0000\u0000\u0000\u00a6\u00a4"+
		"\u0001\u0000\u0000\u0000\u00a7\u00b1\u0005L\u0000\u0000\u00a8\u00ac\u0003"+
		"\n\u0005\u0000\u00a9\u00ac\u0003\f\u0006\u0000\u00aa\u00ac\u0003\u000e"+
		"\u0007\u0000\u00ab\u00a8\u0001\u0000\u0000\u0000\u00ab\u00a9\u0001\u0000"+
		"\u0000\u0000\u00ab\u00aa\u0001\u0000\u0000\u0000\u00ac\u00ad\u0001\u0000"+
		"\u0000\u0000\u00ad\u00ab\u0001\u0000\u0000\u0000\u00ad\u00ae\u0001\u0000"+
		"\u0000\u0000\u00ae\u00b1\u0001\u0000\u0000\u0000\u00af\u00b1\u0001\u0000"+
		"\u0000\u0000\u00b0\u009e\u0001\u0000\u0000\u0000\u00b0\u00ab\u0001\u0000"+
		"\u0000\u0000\u00b0\u00af\u0001\u0000\u0000\u0000\u00b1\t\u0001\u0000\u0000"+
		"\u0000\u00b2\u00b3\u0005)\u0000\u0000\u00b3\u00b4\u0003\u0010\b\u0000"+
		"\u00b4\u000b\u0001\u0000\u0000\u0000\u00b5\u00b6\u0005.\u0000\u0000\u00b6"+
		"\u00b7\u0003\u0010\b\u0000\u00b7\r\u0001\u0000\u0000\u0000\u00b8\u00b9"+
		"\u00051\u0000\u0000\u00b9\u00ba\u0003\u0010\b\u0000\u00ba\u000f\u0001"+
		"\u0000\u0000\u0000\u00bb\u00c0\u0005?\u0000\u0000\u00bc\u00bd\u0005M\u0000"+
		"\u0000\u00bd\u00bf\u0005?\u0000\u0000\u00be\u00bc\u0001\u0000\u0000\u0000"+
		"\u00bf\u00c2\u0001\u0000\u0000\u0000\u00c0\u00be\u0001\u0000\u0000\u0000"+
		"\u00c0\u00c1\u0001\u0000\u0000\u0000\u00c1\u0011\u0001\u0000\u0000\u0000"+
		"\u00c2\u00c0\u0001\u0000\u0000\u0000\u00c3\u00c4\u0005=\u0000\u0000\u00c4"+
		"\u00c6\u0005]\u0000\u0000\u00c5\u00c3\u0001\u0000\u0000\u0000\u00c6\u00c9"+
		"\u0001\u0000\u0000\u0000\u00c7\u00c5\u0001\u0000\u0000\u0000\u00c7\u00c8"+
		"\u0001\u0000\u0000\u0000\u00c8\u0013\u0001\u0000\u0000\u0000\u00c9\u00c7"+
		"\u0001\u0000\u0000\u0000\u00ca\u00cc\u0003\u0016\u000b\u0000\u00cb\u00ca"+
		"\u0001\u0000\u0000\u0000\u00cc\u00cf\u0001\u0000\u0000\u0000\u00cd\u00cb"+
		"\u0001\u0000\u0000\u0000\u00cd\u00ce\u0001\u0000\u0000\u0000\u00ce\u0015"+
		"\u0001\u0000\u0000\u0000\u00cf\u00cd\u0001\u0000\u0000\u0000\u00d0\u00d1"+
		"\u0003\u0018\f\u0000\u00d1\u00d2\u0005]\u0000\u0000\u00d2\u00d5\u0001"+
		"\u0000\u0000\u0000\u00d3\u00d5\u0005]\u0000\u0000\u00d4\u00d0\u0001\u0000"+
		"\u0000\u0000\u00d4\u00d3\u0001\u0000\u0000\u0000\u00d5\u0017\u0001\u0000"+
		"\u0000\u0000\u00d6\u00da\u0003\u001a\r\u0000\u00d7\u00da\u0003\u001c\u000e"+
		"\u0000\u00d8\u00da\u0003\u001e\u000f\u0000\u00d9\u00d6\u0001\u0000\u0000"+
		"\u0000\u00d9\u00d7\u0001\u0000\u0000\u0000\u00d9\u00d8\u0001\u0000\u0000"+
		"\u0000\u00da\u0019\u0001\u0000\u0000\u0000\u00db\u00e7\u0003*\u0015\u0000"+
		"\u00dc\u00e7\u0003,\u0016\u0000\u00dd\u00e7\u0003.\u0017\u0000\u00de\u00e7"+
		"\u00030\u0018\u0000\u00df\u00e7\u00032\u0019\u0000\u00e0\u00e7\u00034"+
		"\u001a\u0000\u00e1\u00e7\u00036\u001b\u0000\u00e2\u00e7\u0003$\u0012\u0000"+
		"\u00e3\u00e7\u00038\u001c\u0000\u00e4\u00e7\u0003:\u001d\u0000\u00e5\u00e7"+
		"\u0003<\u001e\u0000\u00e6\u00db\u0001\u0000\u0000\u0000\u00e6\u00dc\u0001"+
		"\u0000\u0000\u0000\u00e6\u00dd\u0001\u0000\u0000\u0000\u00e6\u00de\u0001"+
		"\u0000\u0000\u0000\u00e6\u00df\u0001\u0000\u0000\u0000\u00e6\u00e0\u0001"+
		"\u0000\u0000\u0000\u00e6\u00e1\u0001\u0000\u0000\u0000\u00e6\u00e2\u0001"+
		"\u0000\u0000\u0000\u00e6\u00e3\u0001\u0000\u0000\u0000\u00e6\u00e4\u0001"+
		"\u0000\u0000\u0000\u00e6\u00e5\u0001\u0000\u0000\u0000\u00e7\u001b\u0001"+
		"\u0000\u0000\u0000\u00e8\u00ec\u0003>\u001f\u0000\u00e9\u00ec\u0003@ "+
		"\u0000\u00ea\u00ec\u0003B!\u0000\u00eb\u00e8\u0001\u0000\u0000\u0000\u00eb"+
		"\u00e9\u0001\u0000\u0000\u0000\u00eb\u00ea\u0001\u0000\u0000\u0000\u00ec"+
		"\u001d\u0001\u0000\u0000\u0000\u00ed\u00f0\u0005-\u0000\u0000\u00ee\u00f1"+
		"\u0003 \u0010\u0000\u00ef\u00f1\u0003\"\u0011\u0000\u00f0\u00ee\u0001"+
		"\u0000\u0000\u0000\u00f0\u00ef\u0001\u0000\u0000\u0000\u00f1\u001f\u0001"+
		"\u0000\u0000\u0000\u00f2\u00f3\u0005\u0017\u0000\u0000\u00f3\u00f4\u0005"+
		"\u000e\u0000\u0000\u00f4\u00f5\u0005]\u0000\u0000\u00f5\u00f6\u0003\u0014"+
		"\n\u0000\u00f6\u00f7\u0005\u0015\u0000\u0000\u00f7!\u0001\u0000\u0000"+
		"\u0000\u00f8\u00f9\u0005\u0019\u0000\u0000\u00f9\u00fc\u0003H$\u0000\u00fa"+
		"\u00fb\u0005(\u0000\u0000\u00fb\u00fd\u0005;\u0000\u0000\u00fc\u00fa\u0001"+
		"\u0000\u0000\u0000\u00fc\u00fd\u0001\u0000\u0000\u0000\u00fd\u0100\u0001"+
		"\u0000\u0000\u0000\u00fe\u00ff\u0005\u0004\u0000\u0000\u00ff\u0101\u0005"+
		"?\u0000\u0000\u0100\u00fe\u0001\u0000\u0000\u0000\u0100\u0101\u0001\u0000"+
		"\u0000\u0000\u0101\u0102\u0001\u0000\u0000\u0000\u0102\u0103\u0005\u000e"+
		"\u0000\u0000\u0103\u0104\u0005]\u0000\u0000\u0104\u0105\u0003\u0014\n"+
		"\u0000\u0105\u0106\u0005\u0015\u0000\u0000\u0106#\u0001\u0000\u0000\u0000"+
		"\u0107\u0108\u0005\n\u0000\u0000\u0108\u010c\u0005\u0019\u0000\u0000\u0109"+
		"\u010d\u0003H$\u0000\u010a\u010b\u0005(\u0000\u0000\u010b\u010d\u0005"+
		";\u0000\u0000\u010c\u0109\u0001\u0000\u0000\u0000\u010c\u010a\u0001\u0000"+
		"\u0000\u0000\u010d%\u0001\u0000\u0000\u0000\u010e\u0117\u0005?\u0000\u0000"+
		"\u010f\u0110\u0005N\u0000\u0000\u0110\u0111\u0003H$\u0000\u0111\u0112"+
		"\u0005O\u0000\u0000\u0112\u0116\u0001\u0000\u0000\u0000\u0113\u0114\u0005"+
		"S\u0000\u0000\u0114\u0116\u0005?\u0000\u0000\u0115\u010f\u0001\u0000\u0000"+
		"\u0000\u0115\u0113\u0001\u0000\u0000\u0000\u0116\u0119\u0001\u0000\u0000"+
		"\u0000\u0117\u0115\u0001\u0000\u0000\u0000\u0117\u0118\u0001\u0000\u0000"+
		"\u0000\u0118\'\u0001\u0000\u0000\u0000\u0119\u0117\u0001\u0000\u0000\u0000"+
		"\u011a\u011f\u0003&\u0013\u0000\u011b\u011c\u0005M\u0000\u0000\u011c\u011e"+
		"\u0003&\u0013\u0000\u011d\u011b\u0001\u0000\u0000\u0000\u011e\u0121\u0001"+
		"\u0000\u0000\u0000\u011f\u011d\u0001\u0000\u0000\u0000\u011f\u0120\u0001"+
		"\u0000\u0000\u0000\u0120)\u0001\u0000\u0000\u0000\u0121\u011f\u0001\u0000"+
		"\u0000\u0000\u0122\u0123\u00052\u0000\u0000\u0123\u0124\u0003(\u0014\u0000"+
		"\u0124\u0125\u0005@\u0000\u0000\u0125\u0126\u0003H$\u0000\u0126+\u0001"+
		"\u0000\u0000\u0000\u0127\u0128\u0005\t\u0000\u0000\u0128\u0129\u0003d"+
		"2\u0000\u0129-\u0001\u0000\u0000\u0000\u012a\u012c\u00050\u0000\u0000"+
		"\u012b\u012d\u0003t:\u0000\u012c\u012b\u0001\u0000\u0000\u0000\u012c\u012d"+
		"\u0001\u0000\u0000\u0000\u012d/\u0001\u0000\u0000\u0000\u012e\u012f\u0005"+
		"\u0011\u0000\u0000\u012f\u0130\u0003H$\u0000\u01301\u0001\u0000\u0000"+
		"\u0000\u0131\u0132\u0005&\u0000\u0000\u0132\u0136\u0003H$\u0000\u0133"+
		"\u0134\u0005\'\u0000\u0000\u0134\u0136\u0003d2\u0000\u0135\u0131\u0001"+
		"\u0000\u0000\u0000\u0135\u0133\u0001\u0000\u0000\u0000\u01363\u0001\u0000"+
		"\u0000\u0000\u0137\u0139\u0005\u001a\u0000\u0000\u0138\u013a\u0003H$\u0000"+
		"\u0139\u0138\u0001\u0000\u0000\u0000\u0139\u013a\u0001\u0000\u0000\u0000"+
		"\u013a5\u0001\u0000\u0000\u0000\u013b\u013c\u0005\u000b\u0000\u0000\u013c"+
		"7\u0001\u0000\u0000\u0000\u013d\u013e\u0005\u0006\u0000\u0000\u013e\u0141"+
		"\u0003H$\u0000\u013f\u0140\u0005!\u0000\u0000\u0140\u0142\u0005?\u0000"+
		"\u0000\u0141\u013f\u0001\u0000\u0000\u0000\u0141\u0142\u0001\u0000\u0000"+
		"\u0000\u01429\u0001\u0000\u0000\u0000\u0143\u0144\u0005\b\u0000\u0000"+
		"\u0144;\u0001\u0000\u0000\u0000\u0145\u0146\u0005\f\u0000\u0000\u0146"+
		"=\u0001\u0000\u0000\u0000\u0147\u0148\u0005\u001f\u0000\u0000\u0148\u0149"+
		"\u0003H$\u0000\u0149\u014a\u0005]\u0000\u0000\u014a\u014e\u0003\u0014"+
		"\n\u0000\u014b\u014c\u0005\u0010\u0000\u0000\u014c\u014d\u0005]\u0000"+
		"\u0000\u014d\u014f\u0003\u0014\n\u0000\u014e\u014b\u0001\u0000\u0000\u0000"+
		"\u014e\u014f\u0001\u0000\u0000\u0000\u014f\u0150\u0001\u0000\u0000\u0000"+
		"\u0150\u0151\u0005\u0014\u0000\u0000\u0151?\u0001\u0000\u0000\u0000\u0152"+
		"\u0153\u0005:\u0000\u0000\u0153\u0154\u0003H$\u0000\u0154\u0155\u0005"+
		"]\u0000\u0000\u0155\u0156\u0003\u0014\n\u0000\u0156\u0157\u0005\u0016"+
		"\u0000\u0000\u0157A\u0001\u0000\u0000\u0000\u0158\u0159\u0005\u001c\u0000"+
		"\u0000\u0159\u015a\u0005\u000f\u0000\u0000\u015a\u015b\u0005?\u0000\u0000"+
		"\u015b\u015c\u0005 \u0000\u0000\u015c\u015d\u0003H$\u0000\u015d\u015e"+
		"\u0005]\u0000\u0000\u015e\u015f\u0003\u0014\n\u0000\u015f\u0160\u0005"+
		"\u0012\u0000\u0000\u0160C\u0001\u0000\u0000\u0000\u0161\u0166\u0005?\u0000"+
		"\u0000\u0162\u0163\u0005S\u0000\u0000\u0163\u0165\u0005?\u0000\u0000\u0164"+
		"\u0162\u0001\u0000\u0000\u0000\u0165\u0168\u0001\u0000\u0000\u0000\u0166"+
		"\u0164\u0001\u0000\u0000\u0000\u0166\u0167\u0001\u0000\u0000\u0000\u0167"+
		"E\u0001\u0000\u0000\u0000\u0168\u0166\u0001\u0000\u0000\u0000\u0169\u016e"+
		"\u0005?\u0000\u0000\u016a\u016b\u00057\u0000\u0000\u016b\u016c\u0005S"+
		"\u0000\u0000\u016c\u016e\u0003D\"\u0000\u016d\u0169\u0001\u0000\u0000"+
		"\u0000\u016d\u016a\u0001\u0000\u0000\u0000\u016eG\u0001\u0000\u0000\u0000"+
		"\u016f\u0170\u0003J%\u0000\u0170I\u0001\u0000\u0000\u0000\u0171\u0176"+
		"\u0003L&\u0000\u0172\u0173\u0005/\u0000\u0000\u0173\u0175\u0003L&\u0000"+
		"\u0174\u0172\u0001\u0000\u0000\u0000\u0175\u0178\u0001\u0000\u0000\u0000"+
		"\u0176\u0174\u0001\u0000\u0000\u0000\u0176\u0177\u0001\u0000\u0000\u0000"+
		"\u0177K\u0001\u0000\u0000\u0000\u0178\u0176\u0001\u0000\u0000\u0000\u0179"+
		"\u017e\u0003N\'\u0000\u017a\u017b\u0005\u0003\u0000\u0000\u017b\u017d"+
		"\u0003N\'\u0000\u017c\u017a\u0001\u0000\u0000\u0000\u017d\u0180\u0001"+
		"\u0000\u0000\u0000\u017e\u017c\u0001\u0000\u0000\u0000\u017e\u017f\u0001"+
		"\u0000\u0000\u0000\u017fM\u0001\u0000\u0000\u0000\u0180\u017e\u0001\u0000"+
		"\u0000\u0000\u0181\u0186\u0003P(\u0000\u0182\u0183\u0005H\u0000\u0000"+
		"\u0183\u0185\u0003P(\u0000\u0184\u0182\u0001\u0000\u0000\u0000\u0185\u0188"+
		"\u0001\u0000\u0000\u0000\u0186\u0184\u0001\u0000\u0000\u0000\u0186\u0187"+
		"\u0001\u0000\u0000\u0000\u0187O\u0001\u0000\u0000\u0000\u0188\u0186\u0001"+
		"\u0000\u0000\u0000\u0189\u018e\u0003R)\u0000\u018a\u018b\u0005I\u0000"+
		"\u0000\u018b\u018d\u0003R)\u0000\u018c\u018a\u0001\u0000\u0000\u0000\u018d"+
		"\u0190\u0001\u0000\u0000\u0000\u018e\u018c\u0001\u0000\u0000\u0000\u018e"+
		"\u018f\u0001\u0000\u0000\u0000\u018fQ\u0001\u0000\u0000\u0000\u0190\u018e"+
		"\u0001\u0000\u0000\u0000\u0191\u0196\u0003T*\u0000\u0192\u0193\u0005G"+
		"\u0000\u0000\u0193\u0195\u0003T*\u0000\u0194\u0192\u0001\u0000\u0000\u0000"+
		"\u0195\u0198\u0001\u0000\u0000\u0000\u0196\u0194\u0001\u0000\u0000\u0000"+
		"\u0196\u0197\u0001\u0000\u0000\u0000\u0197S\u0001\u0000\u0000\u0000\u0198"+
		"\u0196\u0001\u0000\u0000\u0000\u0199\u019e\u0003V+\u0000\u019a\u019b\u0007"+
		"\u0001\u0000\u0000\u019b\u019d\u0003V+\u0000\u019c\u019a\u0001\u0000\u0000"+
		"\u0000\u019d\u01a0\u0001\u0000\u0000\u0000\u019e\u019c\u0001\u0000\u0000"+
		"\u0000\u019e\u019f\u0001\u0000\u0000\u0000\u019fU\u0001\u0000\u0000\u0000"+
		"\u01a0\u019e\u0001\u0000\u0000\u0000\u01a1\u01a6\u0003X,\u0000\u01a2\u01a3"+
		"\u0007\u0002\u0000\u0000\u01a3\u01a5\u0003X,\u0000\u01a4\u01a2\u0001\u0000"+
		"\u0000\u0000\u01a5\u01a8\u0001\u0000\u0000\u0000\u01a6\u01a4\u0001\u0000"+
		"\u0000\u0000\u01a6\u01a7\u0001\u0000\u0000\u0000\u01a7W\u0001\u0000\u0000"+
		"\u0000\u01a8\u01a6\u0001\u0000\u0000\u0000\u01a9\u01ae\u0003Z-\u0000\u01aa"+
		"\u01ab\u0007\u0003\u0000\u0000\u01ab\u01ad\u0003Z-\u0000\u01ac\u01aa\u0001"+
		"\u0000\u0000\u0000\u01ad\u01b0\u0001\u0000\u0000\u0000\u01ae\u01ac\u0001"+
		"\u0000\u0000\u0000\u01ae\u01af\u0001\u0000\u0000\u0000\u01afY\u0001\u0000"+
		"\u0000\u0000\u01b0\u01ae\u0001\u0000\u0000\u0000\u01b1\u01b6\u0003\\."+
		"\u0000\u01b2\u01b3\u0007\u0004\u0000\u0000\u01b3\u01b5\u0003\\.\u0000"+
		"\u01b4\u01b2\u0001\u0000\u0000\u0000\u01b5\u01b8\u0001\u0000\u0000\u0000"+
		"\u01b6\u01b4\u0001\u0000\u0000\u0000\u01b6\u01b7\u0001\u0000\u0000\u0000"+
		"\u01b7[\u0001\u0000\u0000\u0000\u01b8\u01b6\u0001\u0000\u0000\u0000\u01b9"+
		"\u01ba\u0007\u0005\u0000\u0000\u01ba\u01bf\u0003\\.\u0000\u01bb\u01bc"+
		"\u00059\u0000\u0000\u01bc\u01bf\u0003\\.\u0000\u01bd\u01bf\u0003^/\u0000"+
		"\u01be\u01b9\u0001\u0000\u0000\u0000\u01be\u01bb\u0001\u0000\u0000\u0000"+
		"\u01be\u01bd\u0001\u0000\u0000\u0000\u01bf]\u0001\u0000\u0000\u0000\u01c0"+
		"\u01c3\u0003`0\u0000\u01c1\u01c2\u0005F\u0000\u0000\u01c2\u01c4\u0003"+
		"^/\u0000\u01c3\u01c1\u0001\u0000\u0000\u0000\u01c3\u01c4\u0001\u0000\u0000"+
		"\u0000\u01c4_\u0001\u0000\u0000\u0000\u01c5\u01cc\u0003b1\u0000\u01c6"+
		"\u01c7\u0005N\u0000\u0000\u01c7\u01c8\u0003H$\u0000\u01c8\u01c9\u0005"+
		"O\u0000\u0000\u01c9\u01cb\u0001\u0000\u0000\u0000\u01ca\u01c6\u0001\u0000"+
		"\u0000\u0000\u01cb\u01ce\u0001\u0000\u0000\u0000\u01cc\u01ca\u0001\u0000"+
		"\u0000\u0000\u01cc\u01cd\u0001\u0000\u0000\u0000\u01cda\u0001\u0000\u0000"+
		"\u0000\u01ce\u01cc\u0001\u0000\u0000\u0000\u01cf\u01de\u0003h4\u0000\u01d0"+
		"\u01de\u0003f3\u0000\u01d1\u01de\u0005?\u0000\u0000\u01d2\u01de\u0005"+
		"\"\u0000\u0000\u01d3\u01de\u0003d2\u0000\u01d4\u01d5\u0005\u0018\u0000"+
		"\u0000\u01d5\u01d6\u0005K\u0000\u0000\u01d6\u01d7\u0003H$\u0000\u01d7"+
		"\u01d8\u0005L\u0000\u0000\u01d8\u01de\u0001\u0000\u0000\u0000\u01d9\u01da"+
		"\u0005K\u0000\u0000\u01da\u01db\u0003H$\u0000\u01db\u01dc\u0005L\u0000"+
		"\u0000\u01dc\u01de\u0001\u0000\u0000\u0000\u01dd\u01cf\u0001\u0000\u0000"+
		"\u0000\u01dd\u01d0\u0001\u0000\u0000\u0000\u01dd\u01d1\u0001\u0000\u0000"+
		"\u0000\u01dd\u01d2\u0001\u0000\u0000\u0000\u01dd\u01d3\u0001\u0000\u0000"+
		"\u0000\u01dd\u01d4\u0001\u0000\u0000\u0000\u01dd\u01d9\u0001\u0000\u0000"+
		"\u0000\u01dec\u0001\u0000\u0000\u0000\u01df\u01e9\u0003F#\u0000\u01e0"+
		"\u01e9\u0005#\u0000\u0000\u01e1\u01e9\u0005$\u0000\u0000\u01e2\u01e9\u0005"+
		"3\u0000\u0000\u01e3\u01e9\u0005\r\u0000\u0000\u01e4\u01e9\u00055\u0000"+
		"\u0000\u01e5\u01e9\u0005\u0005\u0000\u0000\u01e6\u01e9\u0005\u0002\u0000"+
		"\u0000\u01e7\u01e9\u0005\u0007\u0000\u0000\u01e8\u01df\u0001\u0000\u0000"+
		"\u0000\u01e8\u01e0\u0001\u0000\u0000\u0000\u01e8\u01e1\u0001\u0000\u0000"+
		"\u0000\u01e8\u01e2\u0001\u0000\u0000\u0000\u01e8\u01e3\u0001\u0000\u0000"+
		"\u0000\u01e8\u01e4\u0001\u0000\u0000\u0000\u01e8\u01e5\u0001\u0000\u0000"+
		"\u0000\u01e8\u01e6\u0001\u0000\u0000\u0000\u01e8\u01e7\u0001\u0000\u0000"+
		"\u0000\u01e9\u01ea\u0001\u0000\u0000\u0000\u01ea\u01eb\u0005K\u0000\u0000"+
		"\u01eb\u01ec\u0003r9\u0000\u01ec\u01ed\u0005L\u0000\u0000\u01ede\u0001"+
		"\u0000\u0000\u0000\u01ee\u01ef\u0005T\u0000\u0000\u01ef\u01f0\u0007\u0006"+
		"\u0000\u0000\u01f0\u01f1\u0005U\u0000\u0000\u01f1g\u0001\u0000\u0000\u0000"+
		"\u01f2\u01fa\u0005;\u0000\u0000\u01f3\u01fa\u0005<\u0000\u0000\u01f4\u01fa"+
		"\u0005>\u0000\u0000\u01f5\u01fa\u0003n7\u0000\u01f6\u01fa\u0003p8\u0000"+
		"\u01f7\u01fa\u0003l6\u0000\u01f8\u01fa\u0003j5\u0000\u01f9\u01f2\u0001"+
		"\u0000\u0000\u0000\u01f9\u01f3\u0001\u0000\u0000\u0000\u01f9\u01f4\u0001"+
		"\u0000\u0000\u0000\u01f9\u01f5\u0001\u0000\u0000\u0000\u01f9\u01f6\u0001"+
		"\u0000\u0000\u0000\u01f9\u01f7\u0001\u0000\u0000\u0000\u01f9\u01f8\u0001"+
		"\u0000\u0000\u0000\u01fai\u0001\u0000\u0000\u0000\u01fb\u01fc\u0005*\u0000"+
		"\u0000\u01fck\u0001\u0000\u0000\u0000\u01fd\u01fe\u0007\u0007\u0000\u0000"+
		"\u01fem\u0001\u0000\u0000\u0000\u01ff\u0200\u0005N\u0000\u0000\u0200\u0201"+
		"\u0003r9\u0000\u0201\u0202\u0005O\u0000\u0000\u0202o\u0001\u0000\u0000"+
		"\u0000\u0203\u0204\u0005P\u0000\u0000\u0204\u0205\u0003v;\u0000\u0205"+
		"\u0206\u0005Q\u0000\u0000\u0206q\u0001\u0000\u0000\u0000\u0207\u0209\u0003"+
		"t:\u0000\u0208\u0207\u0001\u0000\u0000\u0000\u0208\u0209\u0001\u0000\u0000"+
		"\u0000\u0209s\u0001\u0000\u0000\u0000\u020a\u020f\u0003H$\u0000\u020b"+
		"\u020c\u0005M\u0000\u0000\u020c\u020e\u0003H$\u0000\u020d\u020b\u0001"+
		"\u0000\u0000\u0000\u020e\u0211\u0001\u0000\u0000\u0000\u020f\u020d\u0001"+
		"\u0000\u0000\u0000\u020f\u0210\u0001\u0000\u0000\u0000\u0210u\u0001\u0000"+
		"\u0000\u0000\u0211\u020f\u0001\u0000\u0000\u0000\u0212\u0214\u0003x<\u0000"+
		"\u0213\u0212\u0001\u0000\u0000\u0000\u0213\u0214\u0001\u0000\u0000\u0000"+
		"\u0214w\u0001\u0000\u0000\u0000\u0215\u021a\u0003z=\u0000\u0216\u0217"+
		"\u0005M\u0000\u0000\u0217\u0219\u0003z=\u0000\u0218\u0216\u0001\u0000"+
		"\u0000\u0000\u0219\u021c\u0001\u0000\u0000\u0000\u021a\u0218\u0001\u0000"+
		"\u0000\u0000\u021a\u021b\u0001\u0000\u0000\u0000\u021by\u0001\u0000\u0000"+
		"\u0000\u021c\u021a\u0001\u0000\u0000\u0000\u021d\u021e\u0005;\u0000\u0000"+
		"\u021e\u021f\u0005R\u0000\u0000\u021f\u0220\u0003H$\u0000\u0220{\u0001"+
		"\u0000\u0000\u00001\u0080\u0088\u008d\u0092\u00a2\u00a4\u00ab\u00ad\u00b0"+
		"\u00c0\u00c7\u00cd\u00d4\u00d9\u00e6\u00eb\u00f0\u00fc\u0100\u010c\u0115"+
		"\u0117\u011f\u012c\u0135\u0139\u0141\u014e\u0166\u016d\u0176\u017e\u0186"+
		"\u018e\u0196\u019e\u01a6\u01ae\u01b6\u01be\u01c3\u01cc\u01dd\u01e8\u01f9"+
		"\u0208\u020f\u0213\u021a";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}