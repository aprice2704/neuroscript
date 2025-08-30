// Generated from /home/aprice/dev/neuroscript/pkg/antlr/NeuroScript.g4 by ANTLR 4.13.1
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
		KW_ATAN=7, KW_BREAK=8, KW_CALL=9, KW_CLEAR=10, KW_CLEAR_ERROR=11, KW_COMMAND=12, 
		KW_CONTINUE=13, KW_COS=14, KW_DO=15, KW_EACH=16, KW_ELSE=17, KW_EMIT=18, 
		KW_ENDCOMMAND=19, KW_ENDFOR=20, KW_ENDFUNC=21, KW_ENDIF=22, KW_ENDON=23, 
		KW_ENDWHILE=24, KW_ERROR=25, KW_EVAL=26, KW_EVENT=27, KW_FAIL=28, KW_FALSE=29, 
		KW_FOR=30, KW_FUNC=31, KW_FUZZY=32, KW_IF=33, KW_IN=34, KW_INTO=35, KW_LAST=36, 
		KW_LEN=37, KW_LN=38, KW_LOG=39, KW_MEANS=40, KW_MUST=41, KW_MUSTBE=42, 
		KW_NAMED=43, KW_NEEDS=44, KW_NIL=45, KW_NO=46, KW_NOT=47, KW_ON=48, KW_OPTIONAL=49, 
		KW_OR=50, KW_PROMPTUSER=51, KW_RETURN=52, KW_RETURNS=53, KW_SET=54, KW_SIN=55, 
		KW_SOME=56, KW_TAN=57, KW_TIMEDATE=58, KW_TOOL=59, KW_TRUE=60, KW_TYPEOF=61, 
		KW_WHILE=62, KW_WHISPER=63, KW_WITH=64, STRING_LIT=65, TRIPLE_BACKTICK_STRING=66, 
		METADATA_LINE=67, NUMBER_LIT=68, IDENTIFIER=69, ASSIGN=70, PLUS=71, MINUS=72, 
		STAR=73, SLASH=74, PERCENT=75, STAR_STAR=76, AMPERSAND=77, PIPE=78, CARET=79, 
		TILDE=80, LPAREN=81, RPAREN=82, COMMA=83, LBRACK=84, RBRACK=85, LBRACE=86, 
		RBRACE=87, COLON=88, DOT=89, PLACEHOLDER_START=90, EQ=91, NEQ=92, GT=93, 
		LT=94, GTE=95, LTE=96, LINE_COMMENT=97, NEWLINE=98, WS=99;
	public static final int
		RULE_program = 0, RULE_file_header = 1, RULE_library_script = 2, RULE_command_script = 3, 
		RULE_library_block = 4, RULE_command_block = 5, RULE_command_statement_list = 6, 
		RULE_command_body_line = 7, RULE_command_statement = 8, RULE_on_error_only_stmt = 9, 
		RULE_simple_command_statement = 10, RULE_procedure_definition = 11, RULE_signature_part = 12, 
		RULE_needs_clause = 13, RULE_optional_clause = 14, RULE_returns_clause = 15, 
		RULE_param_list = 16, RULE_metadata_block = 17, RULE_non_empty_statement_list = 18, 
		RULE_statement_list = 19, RULE_body_line = 20, RULE_statement = 21, RULE_simple_statement = 22, 
		RULE_block_statement = 23, RULE_on_stmt = 24, RULE_error_handler = 25, 
		RULE_event_handler = 26, RULE_clearEventStmt = 27, RULE_lvalue = 28, RULE_lvalue_list = 29, 
		RULE_set_statement = 30, RULE_call_statement = 31, RULE_return_statement = 32, 
		RULE_emit_statement = 33, RULE_whisper_stmt = 34, RULE_must_statement = 35, 
		RULE_fail_statement = 36, RULE_clearErrorStmt = 37, RULE_ask_stmt = 38, 
		RULE_promptuser_stmt = 39, RULE_break_statement = 40, RULE_continue_statement = 41, 
		RULE_if_statement = 42, RULE_while_statement = 43, RULE_for_each_statement = 44, 
		RULE_qualified_identifier = 45, RULE_call_target = 46, RULE_expression = 47, 
		RULE_logical_or_expr = 48, RULE_logical_and_expr = 49, RULE_bitwise_or_expr = 50, 
		RULE_bitwise_xor_expr = 51, RULE_bitwise_and_expr = 52, RULE_equality_expr = 53, 
		RULE_relational_expr = 54, RULE_additive_expr = 55, RULE_multiplicative_expr = 56, 
		RULE_unary_expr = 57, RULE_power_expr = 58, RULE_accessor_expr = 59, RULE_primary = 60, 
		RULE_callable_expr = 61, RULE_placeholder = 62, RULE_literal = 63, RULE_nil_literal = 64, 
		RULE_boolean_literal = 65, RULE_list_literal = 66, RULE_map_literal = 67, 
		RULE_expression_list_opt = 68, RULE_expression_list = 69, RULE_map_entry_list_opt = 70, 
		RULE_map_entry_list = 71, RULE_map_entry = 72;
	private static String[] makeRuleNames() {
		return new String[] {
			"program", "file_header", "library_script", "command_script", "library_block", 
			"command_block", "command_statement_list", "command_body_line", "command_statement", 
			"on_error_only_stmt", "simple_command_statement", "procedure_definition", 
			"signature_part", "needs_clause", "optional_clause", "returns_clause", 
			"param_list", "metadata_block", "non_empty_statement_list", "statement_list", 
			"body_line", "statement", "simple_statement", "block_statement", "on_stmt", 
			"error_handler", "event_handler", "clearEventStmt", "lvalue", "lvalue_list", 
			"set_statement", "call_statement", "return_statement", "emit_statement", 
			"whisper_stmt", "must_statement", "fail_statement", "clearErrorStmt", 
			"ask_stmt", "promptuser_stmt", "break_statement", "continue_statement", 
			"if_statement", "while_statement", "for_each_statement", "qualified_identifier", 
			"call_target", "expression", "logical_or_expr", "logical_and_expr", "bitwise_or_expr", 
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
			"'call'", "'clear'", "'clear_error'", "'command'", "'continue'", "'cos'", 
			"'do'", "'each'", "'else'", "'emit'", "'endcommand'", "'endfor'", "'endfunc'", 
			"'endif'", "'endon'", "'endwhile'", "'error'", "'eval'", "'event'", "'fail'", 
			"'false'", "'for'", "'func'", "'fuzzy'", "'if'", "'in'", "'into'", "'last'", 
			"'len'", "'ln'", "'log'", "'means'", "'must'", "'mustbe'", "'named'", 
			"'needs'", "'nil'", "'no'", "'not'", "'on'", "'optional'", "'or'", "'promptuser'", 
			"'return'", "'returns'", "'set'", "'sin'", "'some'", "'tan'", "'timedate'", 
			"'tool'", "'true'", "'typeof'", "'while'", "'whisper'", "'with'", null, 
			null, null, null, null, "'='", "'+'", "'-'", "'*'", "'/'", "'%'", "'**'", 
			"'&'", "'|'", "'^'", "'~'", "'('", "')'", "','", "'['", "']'", "'{'", 
			"'}'", "':'", "'.'", "'{{'", "'=='", "'!='", "'>'", "'<'", "'>='", "'<='"
		};
	}
	private static final String[] _LITERAL_NAMES = makeLiteralNames();
	private static String[] makeSymbolicNames() {
		return new String[] {
			null, "LINE_ESCAPE_GLOBAL", "KW_ACOS", "KW_AND", "KW_AS", "KW_ASIN", 
			"KW_ASK", "KW_ATAN", "KW_BREAK", "KW_CALL", "KW_CLEAR", "KW_CLEAR_ERROR", 
			"KW_COMMAND", "KW_CONTINUE", "KW_COS", "KW_DO", "KW_EACH", "KW_ELSE", 
			"KW_EMIT", "KW_ENDCOMMAND", "KW_ENDFOR", "KW_ENDFUNC", "KW_ENDIF", "KW_ENDON", 
			"KW_ENDWHILE", "KW_ERROR", "KW_EVAL", "KW_EVENT", "KW_FAIL", "KW_FALSE", 
			"KW_FOR", "KW_FUNC", "KW_FUZZY", "KW_IF", "KW_IN", "KW_INTO", "KW_LAST", 
			"KW_LEN", "KW_LN", "KW_LOG", "KW_MEANS", "KW_MUST", "KW_MUSTBE", "KW_NAMED", 
			"KW_NEEDS", "KW_NIL", "KW_NO", "KW_NOT", "KW_ON", "KW_OPTIONAL", "KW_OR", 
			"KW_PROMPTUSER", "KW_RETURN", "KW_RETURNS", "KW_SET", "KW_SIN", "KW_SOME", 
			"KW_TAN", "KW_TIMEDATE", "KW_TOOL", "KW_TRUE", "KW_TYPEOF", "KW_WHILE", 
			"KW_WHISPER", "KW_WITH", "STRING_LIT", "TRIPLE_BACKTICK_STRING", "METADATA_LINE", 
			"NUMBER_LIT", "IDENTIFIER", "ASSIGN", "PLUS", "MINUS", "STAR", "SLASH", 
			"PERCENT", "STAR_STAR", "AMPERSAND", "PIPE", "CARET", "TILDE", "LPAREN", 
			"RPAREN", "COMMA", "LBRACK", "RBRACK", "LBRACE", "RBRACE", "COLON", "DOT", 
			"PLACEHOLDER_START", "EQ", "NEQ", "GT", "LT", "GTE", "LTE", "LINE_COMMENT", 
			"NEWLINE", "WS"
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
		public Library_scriptContext library_script() {
			return getRuleContext(Library_scriptContext.class,0);
		}
		public Command_scriptContext command_script() {
			return getRuleContext(Command_scriptContext.class,0);
		}
		public ProgramContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_program; }
	}

	public final ProgramContext program() throws RecognitionException {
		ProgramContext _localctx = new ProgramContext(_ctx, getState());
		enterRule(_localctx, 0, RULE_program);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(146);
			file_header();
			setState(149);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_FUNC:
			case KW_ON:
				{
				setState(147);
				library_script();
				}
				break;
			case KW_COMMAND:
				{
				setState(148);
				command_script();
				}
				break;
			case EOF:
				break;
			default:
				break;
			}
			setState(151);
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
			setState(156);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==METADATA_LINE || _la==NEWLINE) {
				{
				{
				setState(153);
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
				setState(158);
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
	public static class Library_scriptContext extends ParserRuleContext {
		public List<Library_blockContext> library_block() {
			return getRuleContexts(Library_blockContext.class);
		}
		public Library_blockContext library_block(int i) {
			return getRuleContext(Library_blockContext.class,i);
		}
		public Library_scriptContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_library_script; }
	}

	public final Library_scriptContext library_script() throws RecognitionException {
		Library_scriptContext _localctx = new Library_scriptContext(_ctx, getState());
		enterRule(_localctx, 4, RULE_library_script);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(160); 
			_errHandler.sync(this);
			_la = _input.LA(1);
			do {
				{
				{
				setState(159);
				library_block();
				}
				}
				setState(162); 
				_errHandler.sync(this);
				_la = _input.LA(1);
			} while ( _la==KW_FUNC || _la==KW_ON );
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
	public static class Command_scriptContext extends ParserRuleContext {
		public List<Command_blockContext> command_block() {
			return getRuleContexts(Command_blockContext.class);
		}
		public Command_blockContext command_block(int i) {
			return getRuleContext(Command_blockContext.class,i);
		}
		public Command_scriptContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_command_script; }
	}

	public final Command_scriptContext command_script() throws RecognitionException {
		Command_scriptContext _localctx = new Command_scriptContext(_ctx, getState());
		enterRule(_localctx, 6, RULE_command_script);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(165); 
			_errHandler.sync(this);
			_la = _input.LA(1);
			do {
				{
				{
				setState(164);
				command_block();
				}
				}
				setState(167); 
				_errHandler.sync(this);
				_la = _input.LA(1);
			} while ( _la==KW_COMMAND );
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
	public static class Library_blockContext extends ParserRuleContext {
		public Procedure_definitionContext procedure_definition() {
			return getRuleContext(Procedure_definitionContext.class,0);
		}
		public TerminalNode KW_ON() { return getToken(NeuroScriptParser.KW_ON, 0); }
		public Event_handlerContext event_handler() {
			return getRuleContext(Event_handlerContext.class,0);
		}
		public List<TerminalNode> NEWLINE() { return getTokens(NeuroScriptParser.NEWLINE); }
		public TerminalNode NEWLINE(int i) {
			return getToken(NeuroScriptParser.NEWLINE, i);
		}
		public Library_blockContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_library_block; }
	}

	public final Library_blockContext library_block() throws RecognitionException {
		Library_blockContext _localctx = new Library_blockContext(_ctx, getState());
		enterRule(_localctx, 8, RULE_library_block);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(172);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_FUNC:
				{
				setState(169);
				procedure_definition();
				}
				break;
			case KW_ON:
				{
				setState(170);
				match(KW_ON);
				setState(171);
				event_handler();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			setState(177);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==NEWLINE) {
				{
				{
				setState(174);
				match(NEWLINE);
				}
				}
				setState(179);
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
	public static class Command_blockContext extends ParserRuleContext {
		public TerminalNode KW_COMMAND() { return getToken(NeuroScriptParser.KW_COMMAND, 0); }
		public List<TerminalNode> NEWLINE() { return getTokens(NeuroScriptParser.NEWLINE); }
		public TerminalNode NEWLINE(int i) {
			return getToken(NeuroScriptParser.NEWLINE, i);
		}
		public Metadata_blockContext metadata_block() {
			return getRuleContext(Metadata_blockContext.class,0);
		}
		public Command_statement_listContext command_statement_list() {
			return getRuleContext(Command_statement_listContext.class,0);
		}
		public TerminalNode KW_ENDCOMMAND() { return getToken(NeuroScriptParser.KW_ENDCOMMAND, 0); }
		public Command_blockContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_command_block; }
	}

	public final Command_blockContext command_block() throws RecognitionException {
		Command_blockContext _localctx = new Command_blockContext(_ctx, getState());
		enterRule(_localctx, 10, RULE_command_block);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(180);
			match(KW_COMMAND);
			setState(181);
			match(NEWLINE);
			setState(182);
			metadata_block();
			setState(183);
			command_statement_list();
			setState(184);
			match(KW_ENDCOMMAND);
			setState(188);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==NEWLINE) {
				{
				{
				setState(185);
				match(NEWLINE);
				}
				}
				setState(190);
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
	public static class Command_statement_listContext extends ParserRuleContext {
		public Command_statementContext command_statement() {
			return getRuleContext(Command_statementContext.class,0);
		}
		public List<TerminalNode> NEWLINE() { return getTokens(NeuroScriptParser.NEWLINE); }
		public TerminalNode NEWLINE(int i) {
			return getToken(NeuroScriptParser.NEWLINE, i);
		}
		public List<Command_body_lineContext> command_body_line() {
			return getRuleContexts(Command_body_lineContext.class);
		}
		public Command_body_lineContext command_body_line(int i) {
			return getRuleContext(Command_body_lineContext.class,i);
		}
		public Command_statement_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_command_statement_list; }
	}

	public final Command_statement_listContext command_statement_list() throws RecognitionException {
		Command_statement_listContext _localctx = new Command_statement_listContext(_ctx, getState());
		enterRule(_localctx, 12, RULE_command_statement_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(194);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==NEWLINE) {
				{
				{
				setState(191);
				match(NEWLINE);
				}
				}
				setState(196);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(197);
			command_statement();
			setState(198);
			match(NEWLINE);
			setState(202);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while ((((_la) & ~0x3f) == 0 && ((1L << _la) & -4591136136171870400L) != 0) || _la==NEWLINE) {
				{
				{
				setState(199);
				command_body_line();
				}
				}
				setState(204);
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
	public static class Command_body_lineContext extends ParserRuleContext {
		public Command_statementContext command_statement() {
			return getRuleContext(Command_statementContext.class,0);
		}
		public TerminalNode NEWLINE() { return getToken(NeuroScriptParser.NEWLINE, 0); }
		public Command_body_lineContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_command_body_line; }
	}

	public final Command_body_lineContext command_body_line() throws RecognitionException {
		Command_body_lineContext _localctx = new Command_body_lineContext(_ctx, getState());
		enterRule(_localctx, 14, RULE_command_body_line);
		try {
			setState(209);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_ASK:
			case KW_BREAK:
			case KW_CALL:
			case KW_CLEAR:
			case KW_CONTINUE:
			case KW_EMIT:
			case KW_FAIL:
			case KW_FOR:
			case KW_IF:
			case KW_MUST:
			case KW_ON:
			case KW_PROMPTUSER:
			case KW_SET:
			case KW_WHILE:
			case KW_WHISPER:
				enterOuterAlt(_localctx, 1);
				{
				setState(205);
				command_statement();
				setState(206);
				match(NEWLINE);
				}
				break;
			case NEWLINE:
				enterOuterAlt(_localctx, 2);
				{
				setState(208);
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
	public static class Command_statementContext extends ParserRuleContext {
		public Simple_command_statementContext simple_command_statement() {
			return getRuleContext(Simple_command_statementContext.class,0);
		}
		public Block_statementContext block_statement() {
			return getRuleContext(Block_statementContext.class,0);
		}
		public On_error_only_stmtContext on_error_only_stmt() {
			return getRuleContext(On_error_only_stmtContext.class,0);
		}
		public Command_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_command_statement; }
	}

	public final Command_statementContext command_statement() throws RecognitionException {
		Command_statementContext _localctx = new Command_statementContext(_ctx, getState());
		enterRule(_localctx, 16, RULE_command_statement);
		try {
			setState(214);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_ASK:
			case KW_BREAK:
			case KW_CALL:
			case KW_CLEAR:
			case KW_CONTINUE:
			case KW_EMIT:
			case KW_FAIL:
			case KW_MUST:
			case KW_PROMPTUSER:
			case KW_SET:
			case KW_WHISPER:
				enterOuterAlt(_localctx, 1);
				{
				setState(211);
				simple_command_statement();
				}
				break;
			case KW_FOR:
			case KW_IF:
			case KW_WHILE:
				enterOuterAlt(_localctx, 2);
				{
				setState(212);
				block_statement();
				}
				break;
			case KW_ON:
				enterOuterAlt(_localctx, 3);
				{
				setState(213);
				on_error_only_stmt();
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
	public static class On_error_only_stmtContext extends ParserRuleContext {
		public TerminalNode KW_ON() { return getToken(NeuroScriptParser.KW_ON, 0); }
		public Error_handlerContext error_handler() {
			return getRuleContext(Error_handlerContext.class,0);
		}
		public On_error_only_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_on_error_only_stmt; }
	}

	public final On_error_only_stmtContext on_error_only_stmt() throws RecognitionException {
		On_error_only_stmtContext _localctx = new On_error_only_stmtContext(_ctx, getState());
		enterRule(_localctx, 18, RULE_on_error_only_stmt);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(216);
			match(KW_ON);
			setState(217);
			error_handler();
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
	public static class Simple_command_statementContext extends ParserRuleContext {
		public Set_statementContext set_statement() {
			return getRuleContext(Set_statementContext.class,0);
		}
		public Call_statementContext call_statement() {
			return getRuleContext(Call_statementContext.class,0);
		}
		public Emit_statementContext emit_statement() {
			return getRuleContext(Emit_statementContext.class,0);
		}
		public Whisper_stmtContext whisper_stmt() {
			return getRuleContext(Whisper_stmtContext.class,0);
		}
		public Must_statementContext must_statement() {
			return getRuleContext(Must_statementContext.class,0);
		}
		public Fail_statementContext fail_statement() {
			return getRuleContext(Fail_statementContext.class,0);
		}
		public ClearEventStmtContext clearEventStmt() {
			return getRuleContext(ClearEventStmtContext.class,0);
		}
		public Ask_stmtContext ask_stmt() {
			return getRuleContext(Ask_stmtContext.class,0);
		}
		public Promptuser_stmtContext promptuser_stmt() {
			return getRuleContext(Promptuser_stmtContext.class,0);
		}
		public Break_statementContext break_statement() {
			return getRuleContext(Break_statementContext.class,0);
		}
		public Continue_statementContext continue_statement() {
			return getRuleContext(Continue_statementContext.class,0);
		}
		public Simple_command_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_simple_command_statement; }
	}

	public final Simple_command_statementContext simple_command_statement() throws RecognitionException {
		Simple_command_statementContext _localctx = new Simple_command_statementContext(_ctx, getState());
		enterRule(_localctx, 20, RULE_simple_command_statement);
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
			case KW_EMIT:
				enterOuterAlt(_localctx, 3);
				{
				setState(221);
				emit_statement();
				}
				break;
			case KW_WHISPER:
				enterOuterAlt(_localctx, 4);
				{
				setState(222);
				whisper_stmt();
				}
				break;
			case KW_MUST:
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
			case KW_CLEAR:
				enterOuterAlt(_localctx, 7);
				{
				setState(225);
				clearEventStmt();
				}
				break;
			case KW_ASK:
				enterOuterAlt(_localctx, 8);
				{
				setState(226);
				ask_stmt();
				}
				break;
			case KW_PROMPTUSER:
				enterOuterAlt(_localctx, 9);
				{
				setState(227);
				promptuser_stmt();
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
		public Non_empty_statement_listContext non_empty_statement_list() {
			return getRuleContext(Non_empty_statement_listContext.class,0);
		}
		public TerminalNode KW_ENDFUNC() { return getToken(NeuroScriptParser.KW_ENDFUNC, 0); }
		public Procedure_definitionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_procedure_definition; }
	}

	public final Procedure_definitionContext procedure_definition() throws RecognitionException {
		Procedure_definitionContext _localctx = new Procedure_definitionContext(_ctx, getState());
		enterRule(_localctx, 22, RULE_procedure_definition);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(232);
			match(KW_FUNC);
			setState(233);
			match(IDENTIFIER);
			setState(234);
			signature_part();
			setState(235);
			match(KW_MEANS);
			setState(236);
			match(NEWLINE);
			setState(237);
			metadata_block();
			setState(238);
			non_empty_statement_list();
			setState(239);
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
		enterRule(_localctx, 24, RULE_signature_part);
		int _la;
		try {
			setState(259);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case LPAREN:
				enterOuterAlt(_localctx, 1);
				{
				setState(241);
				match(LPAREN);
				setState(247);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while ((((_la) & ~0x3f) == 0 && ((1L << _la) & 9587741394206720L) != 0)) {
					{
					setState(245);
					_errHandler.sync(this);
					switch (_input.LA(1)) {
					case KW_NEEDS:
						{
						setState(242);
						needs_clause();
						}
						break;
					case KW_OPTIONAL:
						{
						setState(243);
						optional_clause();
						}
						break;
					case KW_RETURNS:
						{
						setState(244);
						returns_clause();
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					}
					setState(249);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(250);
				match(RPAREN);
				}
				break;
			case KW_NEEDS:
			case KW_OPTIONAL:
			case KW_RETURNS:
				enterOuterAlt(_localctx, 2);
				{
				setState(254); 
				_errHandler.sync(this);
				_la = _input.LA(1);
				do {
					{
					setState(254);
					_errHandler.sync(this);
					switch (_input.LA(1)) {
					case KW_NEEDS:
						{
						setState(251);
						needs_clause();
						}
						break;
					case KW_OPTIONAL:
						{
						setState(252);
						optional_clause();
						}
						break;
					case KW_RETURNS:
						{
						setState(253);
						returns_clause();
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					}
					setState(256); 
					_errHandler.sync(this);
					_la = _input.LA(1);
				} while ( (((_la) & ~0x3f) == 0 && ((1L << _la) & 9587741394206720L) != 0) );
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
		enterRule(_localctx, 26, RULE_needs_clause);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(261);
			match(KW_NEEDS);
			setState(262);
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
		enterRule(_localctx, 28, RULE_optional_clause);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(264);
			match(KW_OPTIONAL);
			setState(265);
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
		enterRule(_localctx, 30, RULE_returns_clause);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(267);
			match(KW_RETURNS);
			setState(268);
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
		enterRule(_localctx, 32, RULE_param_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(270);
			match(IDENTIFIER);
			setState(275);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(271);
				match(COMMA);
				setState(272);
				match(IDENTIFIER);
				}
				}
				setState(277);
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
		enterRule(_localctx, 34, RULE_metadata_block);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(282);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==METADATA_LINE) {
				{
				{
				setState(278);
				match(METADATA_LINE);
				setState(279);
				match(NEWLINE);
				}
				}
				setState(284);
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
	public static class Non_empty_statement_listContext extends ParserRuleContext {
		public StatementContext statement() {
			return getRuleContext(StatementContext.class,0);
		}
		public List<TerminalNode> NEWLINE() { return getTokens(NeuroScriptParser.NEWLINE); }
		public TerminalNode NEWLINE(int i) {
			return getToken(NeuroScriptParser.NEWLINE, i);
		}
		public List<Body_lineContext> body_line() {
			return getRuleContexts(Body_lineContext.class);
		}
		public Body_lineContext body_line(int i) {
			return getRuleContext(Body_lineContext.class,i);
		}
		public Non_empty_statement_listContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_non_empty_statement_list; }
	}

	public final Non_empty_statement_listContext non_empty_statement_list() throws RecognitionException {
		Non_empty_statement_listContext _localctx = new Non_empty_statement_listContext(_ctx, getState());
		enterRule(_localctx, 36, RULE_non_empty_statement_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(288);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==NEWLINE) {
				{
				{
				setState(285);
				match(NEWLINE);
				}
				}
				setState(290);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(291);
			statement();
			setState(292);
			match(NEWLINE);
			setState(296);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while ((((_la) & ~0x3f) == 0 && ((1L << _la) & -4586632536544497856L) != 0) || _la==NEWLINE) {
				{
				{
				setState(293);
				body_line();
				}
				}
				setState(298);
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
		enterRule(_localctx, 38, RULE_statement_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(302);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while ((((_la) & ~0x3f) == 0 && ((1L << _la) & -4586632536544497856L) != 0) || _la==NEWLINE) {
				{
				{
				setState(299);
				body_line();
				}
				}
				setState(304);
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
		enterRule(_localctx, 40, RULE_body_line);
		try {
			setState(309);
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
			case KW_ON:
			case KW_PROMPTUSER:
			case KW_RETURN:
			case KW_SET:
			case KW_WHILE:
			case KW_WHISPER:
				enterOuterAlt(_localctx, 1);
				{
				setState(305);
				statement();
				setState(306);
				match(NEWLINE);
				}
				break;
			case NEWLINE:
				enterOuterAlt(_localctx, 2);
				{
				setState(308);
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
		enterRule(_localctx, 42, RULE_statement);
		try {
			setState(314);
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
			case KW_PROMPTUSER:
			case KW_RETURN:
			case KW_SET:
			case KW_WHISPER:
				enterOuterAlt(_localctx, 1);
				{
				setState(311);
				simple_statement();
				}
				break;
			case KW_FOR:
			case KW_IF:
			case KW_WHILE:
				enterOuterAlt(_localctx, 2);
				{
				setState(312);
				block_statement();
				}
				break;
			case KW_ON:
				enterOuterAlt(_localctx, 3);
				{
				setState(313);
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
		public Whisper_stmtContext whisper_stmt() {
			return getRuleContext(Whisper_stmtContext.class,0);
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
		public Promptuser_stmtContext promptuser_stmt() {
			return getRuleContext(Promptuser_stmtContext.class,0);
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
		enterRule(_localctx, 44, RULE_simple_statement);
		try {
			setState(329);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_SET:
				enterOuterAlt(_localctx, 1);
				{
				setState(316);
				set_statement();
				}
				break;
			case KW_CALL:
				enterOuterAlt(_localctx, 2);
				{
				setState(317);
				call_statement();
				}
				break;
			case KW_RETURN:
				enterOuterAlt(_localctx, 3);
				{
				setState(318);
				return_statement();
				}
				break;
			case KW_EMIT:
				enterOuterAlt(_localctx, 4);
				{
				setState(319);
				emit_statement();
				}
				break;
			case KW_WHISPER:
				enterOuterAlt(_localctx, 5);
				{
				setState(320);
				whisper_stmt();
				}
				break;
			case KW_MUST:
				enterOuterAlt(_localctx, 6);
				{
				setState(321);
				must_statement();
				}
				break;
			case KW_FAIL:
				enterOuterAlt(_localctx, 7);
				{
				setState(322);
				fail_statement();
				}
				break;
			case KW_CLEAR_ERROR:
				enterOuterAlt(_localctx, 8);
				{
				setState(323);
				clearErrorStmt();
				}
				break;
			case KW_CLEAR:
				enterOuterAlt(_localctx, 9);
				{
				setState(324);
				clearEventStmt();
				}
				break;
			case KW_ASK:
				enterOuterAlt(_localctx, 10);
				{
				setState(325);
				ask_stmt();
				}
				break;
			case KW_PROMPTUSER:
				enterOuterAlt(_localctx, 11);
				{
				setState(326);
				promptuser_stmt();
				}
				break;
			case KW_BREAK:
				enterOuterAlt(_localctx, 12);
				{
				setState(327);
				break_statement();
				}
				break;
			case KW_CONTINUE:
				enterOuterAlt(_localctx, 13);
				{
				setState(328);
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
		enterRule(_localctx, 46, RULE_block_statement);
		try {
			setState(334);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_IF:
				enterOuterAlt(_localctx, 1);
				{
				setState(331);
				if_statement();
				}
				break;
			case KW_WHILE:
				enterOuterAlt(_localctx, 2);
				{
				setState(332);
				while_statement();
				}
				break;
			case KW_FOR:
				enterOuterAlt(_localctx, 3);
				{
				setState(333);
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
		enterRule(_localctx, 48, RULE_on_stmt);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(336);
			match(KW_ON);
			setState(339);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_ERROR:
				{
				setState(337);
				error_handler();
				}
				break;
			case KW_EVENT:
				{
				setState(338);
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
		public Non_empty_statement_listContext non_empty_statement_list() {
			return getRuleContext(Non_empty_statement_listContext.class,0);
		}
		public TerminalNode KW_ENDON() { return getToken(NeuroScriptParser.KW_ENDON, 0); }
		public Error_handlerContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_error_handler; }
	}

	public final Error_handlerContext error_handler() throws RecognitionException {
		Error_handlerContext _localctx = new Error_handlerContext(_ctx, getState());
		enterRule(_localctx, 50, RULE_error_handler);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(341);
			match(KW_ERROR);
			setState(342);
			match(KW_DO);
			setState(343);
			match(NEWLINE);
			setState(344);
			non_empty_statement_list();
			setState(345);
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
		public Non_empty_statement_listContext non_empty_statement_list() {
			return getRuleContext(Non_empty_statement_listContext.class,0);
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
		enterRule(_localctx, 52, RULE_event_handler);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(347);
			match(KW_EVENT);
			setState(348);
			expression();
			setState(351);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==KW_NAMED) {
				{
				setState(349);
				match(KW_NAMED);
				setState(350);
				match(STRING_LIT);
				}
			}

			setState(355);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==KW_AS) {
				{
				setState(353);
				match(KW_AS);
				setState(354);
				match(IDENTIFIER);
				}
			}

			setState(357);
			match(KW_DO);
			setState(358);
			match(NEWLINE);
			setState(359);
			non_empty_statement_list();
			setState(360);
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
		enterRule(_localctx, 54, RULE_clearEventStmt);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(362);
			match(KW_CLEAR);
			setState(363);
			match(KW_EVENT);
			setState(367);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_ACOS:
			case KW_ASIN:
			case KW_ATAN:
			case KW_COS:
			case KW_EVAL:
			case KW_FALSE:
			case KW_LAST:
			case KW_LEN:
			case KW_LN:
			case KW_LOG:
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
				setState(364);
				expression();
				}
				break;
			case KW_NAMED:
				{
				setState(365);
				match(KW_NAMED);
				setState(366);
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
		enterRule(_localctx, 56, RULE_lvalue);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(369);
			match(IDENTIFIER);
			setState(378);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==LBRACK || _la==DOT) {
				{
				setState(376);
				_errHandler.sync(this);
				switch (_input.LA(1)) {
				case LBRACK:
					{
					setState(370);
					match(LBRACK);
					setState(371);
					expression();
					setState(372);
					match(RBRACK);
					}
					break;
				case DOT:
					{
					setState(374);
					match(DOT);
					setState(375);
					match(IDENTIFIER);
					}
					break;
				default:
					throw new NoViableAltException(this);
				}
				}
				setState(380);
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
		enterRule(_localctx, 58, RULE_lvalue_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(381);
			lvalue();
			setState(386);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(382);
				match(COMMA);
				setState(383);
				lvalue();
				}
				}
				setState(388);
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
		enterRule(_localctx, 60, RULE_set_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(389);
			match(KW_SET);
			setState(390);
			lvalue_list();
			setState(391);
			match(ASSIGN);
			setState(392);
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
		enterRule(_localctx, 62, RULE_call_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(394);
			match(KW_CALL);
			setState(395);
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
		enterRule(_localctx, 64, RULE_return_statement);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(397);
			match(KW_RETURN);
			setState(399);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if ((((_la) & ~0x3f) == 0 && ((1L << _la) & 4287674167257481380L) != 0) || ((((_la - 65)) & ~0x3f) == 0 && ((1L << (_la - 65)) & 36274331L) != 0)) {
				{
				setState(398);
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
		enterRule(_localctx, 66, RULE_emit_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(401);
			match(KW_EMIT);
			setState(402);
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
	public static class Whisper_stmtContext extends ParserRuleContext {
		public TerminalNode KW_WHISPER() { return getToken(NeuroScriptParser.KW_WHISPER, 0); }
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public TerminalNode COMMA() { return getToken(NeuroScriptParser.COMMA, 0); }
		public Whisper_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_whisper_stmt; }
	}

	public final Whisper_stmtContext whisper_stmt() throws RecognitionException {
		Whisper_stmtContext _localctx = new Whisper_stmtContext(_ctx, getState());
		enterRule(_localctx, 68, RULE_whisper_stmt);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(404);
			match(KW_WHISPER);
			setState(405);
			expression();
			setState(406);
			match(COMMA);
			setState(407);
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
		public Must_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_must_statement; }
	}

	public final Must_statementContext must_statement() throws RecognitionException {
		Must_statementContext _localctx = new Must_statementContext(_ctx, getState());
		enterRule(_localctx, 70, RULE_must_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(409);
			match(KW_MUST);
			setState(410);
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
		enterRule(_localctx, 72, RULE_fail_statement);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(412);
			match(KW_FAIL);
			setState(414);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if ((((_la) & ~0x3f) == 0 && ((1L << _la) & 4287674167257481380L) != 0) || ((((_la - 65)) & ~0x3f) == 0 && ((1L << (_la - 65)) & 36274331L) != 0)) {
				{
				setState(413);
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
		enterRule(_localctx, 74, RULE_clearErrorStmt);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(416);
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
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public TerminalNode COMMA() { return getToken(NeuroScriptParser.COMMA, 0); }
		public TerminalNode KW_WITH() { return getToken(NeuroScriptParser.KW_WITH, 0); }
		public TerminalNode KW_INTO() { return getToken(NeuroScriptParser.KW_INTO, 0); }
		public LvalueContext lvalue() {
			return getRuleContext(LvalueContext.class,0);
		}
		public Ask_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_ask_stmt; }
	}

	public final Ask_stmtContext ask_stmt() throws RecognitionException {
		Ask_stmtContext _localctx = new Ask_stmtContext(_ctx, getState());
		enterRule(_localctx, 76, RULE_ask_stmt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(418);
			match(KW_ASK);
			setState(419);
			expression();
			setState(420);
			match(COMMA);
			setState(421);
			expression();
			setState(424);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==KW_WITH) {
				{
				setState(422);
				match(KW_WITH);
				setState(423);
				expression();
				}
			}

			setState(428);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==KW_INTO) {
				{
				setState(426);
				match(KW_INTO);
				setState(427);
				lvalue();
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
	public static class Promptuser_stmtContext extends ParserRuleContext {
		public TerminalNode KW_PROMPTUSER() { return getToken(NeuroScriptParser.KW_PROMPTUSER, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode KW_INTO() { return getToken(NeuroScriptParser.KW_INTO, 0); }
		public LvalueContext lvalue() {
			return getRuleContext(LvalueContext.class,0);
		}
		public Promptuser_stmtContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_promptuser_stmt; }
	}

	public final Promptuser_stmtContext promptuser_stmt() throws RecognitionException {
		Promptuser_stmtContext _localctx = new Promptuser_stmtContext(_ctx, getState());
		enterRule(_localctx, 78, RULE_promptuser_stmt);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(430);
			match(KW_PROMPTUSER);
			setState(431);
			expression();
			setState(432);
			match(KW_INTO);
			setState(433);
			lvalue();
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
		enterRule(_localctx, 80, RULE_break_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(435);
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
		enterRule(_localctx, 82, RULE_continue_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(437);
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
		public List<Non_empty_statement_listContext> non_empty_statement_list() {
			return getRuleContexts(Non_empty_statement_listContext.class);
		}
		public Non_empty_statement_listContext non_empty_statement_list(int i) {
			return getRuleContext(Non_empty_statement_listContext.class,i);
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
		enterRule(_localctx, 84, RULE_if_statement);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(439);
			match(KW_IF);
			setState(440);
			expression();
			setState(441);
			match(NEWLINE);
			setState(442);
			non_empty_statement_list();
			setState(446);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==KW_ELSE) {
				{
				setState(443);
				match(KW_ELSE);
				setState(444);
				match(NEWLINE);
				setState(445);
				non_empty_statement_list();
				}
			}

			setState(448);
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
		public Non_empty_statement_listContext non_empty_statement_list() {
			return getRuleContext(Non_empty_statement_listContext.class,0);
		}
		public TerminalNode KW_ENDWHILE() { return getToken(NeuroScriptParser.KW_ENDWHILE, 0); }
		public While_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_while_statement; }
	}

	public final While_statementContext while_statement() throws RecognitionException {
		While_statementContext _localctx = new While_statementContext(_ctx, getState());
		enterRule(_localctx, 86, RULE_while_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(450);
			match(KW_WHILE);
			setState(451);
			expression();
			setState(452);
			match(NEWLINE);
			setState(453);
			non_empty_statement_list();
			setState(454);
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
		public Non_empty_statement_listContext non_empty_statement_list() {
			return getRuleContext(Non_empty_statement_listContext.class,0);
		}
		public TerminalNode KW_ENDFOR() { return getToken(NeuroScriptParser.KW_ENDFOR, 0); }
		public For_each_statementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_for_each_statement; }
	}

	public final For_each_statementContext for_each_statement() throws RecognitionException {
		For_each_statementContext _localctx = new For_each_statementContext(_ctx, getState());
		enterRule(_localctx, 88, RULE_for_each_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(456);
			match(KW_FOR);
			setState(457);
			match(KW_EACH);
			setState(458);
			match(IDENTIFIER);
			setState(459);
			match(KW_IN);
			setState(460);
			expression();
			setState(461);
			match(NEWLINE);
			setState(462);
			non_empty_statement_list();
			setState(463);
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
		enterRule(_localctx, 90, RULE_qualified_identifier);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(465);
			match(IDENTIFIER);
			setState(470);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==DOT) {
				{
				{
				setState(466);
				match(DOT);
				setState(467);
				match(IDENTIFIER);
				}
				}
				setState(472);
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
		enterRule(_localctx, 92, RULE_call_target);
		try {
			setState(477);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case IDENTIFIER:
				enterOuterAlt(_localctx, 1);
				{
				setState(473);
				match(IDENTIFIER);
				}
				break;
			case KW_TOOL:
				enterOuterAlt(_localctx, 2);
				{
				setState(474);
				match(KW_TOOL);
				setState(475);
				match(DOT);
				setState(476);
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
		enterRule(_localctx, 94, RULE_expression);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(479);
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
		enterRule(_localctx, 96, RULE_logical_or_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(481);
			logical_and_expr();
			setState(486);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==KW_OR) {
				{
				{
				setState(482);
				match(KW_OR);
				setState(483);
				logical_and_expr();
				}
				}
				setState(488);
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
		enterRule(_localctx, 98, RULE_logical_and_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(489);
			bitwise_or_expr();
			setState(494);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==KW_AND) {
				{
				{
				setState(490);
				match(KW_AND);
				setState(491);
				bitwise_or_expr();
				}
				}
				setState(496);
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
		enterRule(_localctx, 100, RULE_bitwise_or_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(497);
			bitwise_xor_expr();
			setState(502);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==PIPE) {
				{
				{
				setState(498);
				match(PIPE);
				setState(499);
				bitwise_xor_expr();
				}
				}
				setState(504);
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
		enterRule(_localctx, 102, RULE_bitwise_xor_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(505);
			bitwise_and_expr();
			setState(510);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==CARET) {
				{
				{
				setState(506);
				match(CARET);
				setState(507);
				bitwise_and_expr();
				}
				}
				setState(512);
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
		enterRule(_localctx, 104, RULE_bitwise_and_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(513);
			equality_expr();
			setState(518);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==AMPERSAND) {
				{
				{
				setState(514);
				match(AMPERSAND);
				setState(515);
				equality_expr();
				}
				}
				setState(520);
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
		enterRule(_localctx, 106, RULE_equality_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(521);
			relational_expr();
			setState(526);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==EQ || _la==NEQ) {
				{
				{
				setState(522);
				_la = _input.LA(1);
				if ( !(_la==EQ || _la==NEQ) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(523);
				relational_expr();
				}
				}
				setState(528);
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
		enterRule(_localctx, 108, RULE_relational_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(529);
			additive_expr();
			setState(534);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (((((_la - 93)) & ~0x3f) == 0 && ((1L << (_la - 93)) & 15L) != 0)) {
				{
				{
				setState(530);
				_la = _input.LA(1);
				if ( !(((((_la - 93)) & ~0x3f) == 0 && ((1L << (_la - 93)) & 15L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(531);
				additive_expr();
				}
				}
				setState(536);
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
		enterRule(_localctx, 110, RULE_additive_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(537);
			multiplicative_expr();
			setState(542);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==PLUS || _la==MINUS) {
				{
				{
				setState(538);
				_la = _input.LA(1);
				if ( !(_la==PLUS || _la==MINUS) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(539);
				multiplicative_expr();
				}
				}
				setState(544);
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
		enterRule(_localctx, 112, RULE_multiplicative_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(545);
			unary_expr();
			setState(550);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (((((_la - 73)) & ~0x3f) == 0 && ((1L << (_la - 73)) & 7L) != 0)) {
				{
				{
				setState(546);
				_la = _input.LA(1);
				if ( !(((((_la - 73)) & ~0x3f) == 0 && ((1L << (_la - 73)) & 7L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(547);
				unary_expr();
				}
				}
				setState(552);
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
		enterRule(_localctx, 114, RULE_unary_expr);
		int _la;
		try {
			setState(558);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_NO:
			case KW_NOT:
			case KW_SOME:
			case MINUS:
			case TILDE:
				enterOuterAlt(_localctx, 1);
				{
				setState(553);
				_la = _input.LA(1);
				if ( !(((((_la - 46)) & ~0x3f) == 0 && ((1L << (_la - 46)) & 17246979075L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(554);
				unary_expr();
				}
				break;
			case KW_TYPEOF:
				enterOuterAlt(_localctx, 2);
				{
				setState(555);
				match(KW_TYPEOF);
				setState(556);
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
			case KW_LEN:
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
				setState(557);
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
		enterRule(_localctx, 116, RULE_power_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(560);
			accessor_expr();
			setState(563);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==STAR_STAR) {
				{
				setState(561);
				match(STAR_STAR);
				setState(562);
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
		enterRule(_localctx, 118, RULE_accessor_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(565);
			primary();
			setState(572);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==LBRACK) {
				{
				{
				setState(566);
				match(LBRACK);
				setState(567);
				expression();
				setState(568);
				match(RBRACK);
				}
				}
				setState(574);
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
		enterRule(_localctx, 120, RULE_primary);
		try {
			setState(589);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,52,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(575);
				literal();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(576);
				placeholder();
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(577);
				match(IDENTIFIER);
				}
				break;
			case 4:
				enterOuterAlt(_localctx, 4);
				{
				setState(578);
				match(KW_LAST);
				}
				break;
			case 5:
				enterOuterAlt(_localctx, 5);
				{
				setState(579);
				callable_expr();
				}
				break;
			case 6:
				enterOuterAlt(_localctx, 6);
				{
				setState(580);
				match(KW_EVAL);
				setState(581);
				match(LPAREN);
				setState(582);
				expression();
				setState(583);
				match(RPAREN);
				}
				break;
			case 7:
				enterOuterAlt(_localctx, 7);
				{
				setState(585);
				match(LPAREN);
				setState(586);
				expression();
				setState(587);
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
		public TerminalNode KW_LEN() { return getToken(NeuroScriptParser.KW_LEN, 0); }
		public Callable_exprContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_callable_expr; }
	}

	public final Callable_exprContext callable_expr() throws RecognitionException {
		Callable_exprContext _localctx = new Callable_exprContext(_ctx, getState());
		enterRule(_localctx, 122, RULE_callable_expr);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(601);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_TOOL:
			case IDENTIFIER:
				{
				setState(591);
				call_target();
				}
				break;
			case KW_LN:
				{
				setState(592);
				match(KW_LN);
				}
				break;
			case KW_LOG:
				{
				setState(593);
				match(KW_LOG);
				}
				break;
			case KW_SIN:
				{
				setState(594);
				match(KW_SIN);
				}
				break;
			case KW_COS:
				{
				setState(595);
				match(KW_COS);
				}
				break;
			case KW_TAN:
				{
				setState(596);
				match(KW_TAN);
				}
				break;
			case KW_ASIN:
				{
				setState(597);
				match(KW_ASIN);
				}
				break;
			case KW_ACOS:
				{
				setState(598);
				match(KW_ACOS);
				}
				break;
			case KW_ATAN:
				{
				setState(599);
				match(KW_ATAN);
				}
				break;
			case KW_LEN:
				{
				setState(600);
				match(KW_LEN);
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			setState(603);
			match(LPAREN);
			setState(604);
			expression_list_opt();
			setState(605);
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
		public List<TerminalNode> RBRACE() { return getTokens(NeuroScriptParser.RBRACE); }
		public TerminalNode RBRACE(int i) {
			return getToken(NeuroScriptParser.RBRACE, i);
		}
		public TerminalNode IDENTIFIER() { return getToken(NeuroScriptParser.IDENTIFIER, 0); }
		public TerminalNode KW_LAST() { return getToken(NeuroScriptParser.KW_LAST, 0); }
		public PlaceholderContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_placeholder; }
	}

	public final PlaceholderContext placeholder() throws RecognitionException {
		PlaceholderContext _localctx = new PlaceholderContext(_ctx, getState());
		enterRule(_localctx, 124, RULE_placeholder);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(607);
			match(PLACEHOLDER_START);
			setState(608);
			_la = _input.LA(1);
			if ( !(_la==KW_LAST || _la==IDENTIFIER) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			setState(609);
			match(RBRACE);
			setState(610);
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
		enterRule(_localctx, 126, RULE_literal);
		try {
			setState(619);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case STRING_LIT:
				enterOuterAlt(_localctx, 1);
				{
				setState(612);
				match(STRING_LIT);
				}
				break;
			case TRIPLE_BACKTICK_STRING:
				enterOuterAlt(_localctx, 2);
				{
				setState(613);
				match(TRIPLE_BACKTICK_STRING);
				}
				break;
			case NUMBER_LIT:
				enterOuterAlt(_localctx, 3);
				{
				setState(614);
				match(NUMBER_LIT);
				}
				break;
			case LBRACK:
				enterOuterAlt(_localctx, 4);
				{
				setState(615);
				list_literal();
				}
				break;
			case LBRACE:
				enterOuterAlt(_localctx, 5);
				{
				setState(616);
				map_literal();
				}
				break;
			case KW_FALSE:
			case KW_TRUE:
				enterOuterAlt(_localctx, 6);
				{
				setState(617);
				boolean_literal();
				}
				break;
			case KW_NIL:
				enterOuterAlt(_localctx, 7);
				{
				setState(618);
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
		enterRule(_localctx, 128, RULE_nil_literal);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(621);
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
		enterRule(_localctx, 130, RULE_boolean_literal);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(623);
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
		enterRule(_localctx, 132, RULE_list_literal);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(625);
			match(LBRACK);
			setState(626);
			expression_list_opt();
			setState(627);
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
		enterRule(_localctx, 134, RULE_map_literal);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(629);
			match(LBRACE);
			setState(630);
			map_entry_list_opt();
			setState(631);
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
		enterRule(_localctx, 136, RULE_expression_list_opt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(634);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if ((((_la) & ~0x3f) == 0 && ((1L << _la) & 4287674167257481380L) != 0) || ((((_la - 65)) & ~0x3f) == 0 && ((1L << (_la - 65)) & 36274331L) != 0)) {
				{
				setState(633);
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
		enterRule(_localctx, 138, RULE_expression_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(636);
			expression();
			setState(641);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(637);
				match(COMMA);
				setState(638);
				expression();
				}
				}
				setState(643);
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
		enterRule(_localctx, 140, RULE_map_entry_list_opt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(645);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==STRING_LIT) {
				{
				setState(644);
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
		enterRule(_localctx, 142, RULE_map_entry_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(647);
			map_entry();
			setState(652);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(648);
				match(COMMA);
				setState(649);
				map_entry();
				}
				}
				setState(654);
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
		enterRule(_localctx, 144, RULE_map_entry);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(655);
			match(STRING_LIT);
			setState(656);
			match(COLON);
			setState(657);
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
		"\u0004\u0001c\u0294\u0002\u0000\u0007\u0000\u0002\u0001\u0007\u0001\u0002"+
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
		"<\u0007<\u0002=\u0007=\u0002>\u0007>\u0002?\u0007?\u0002@\u0007@\u0002"+
		"A\u0007A\u0002B\u0007B\u0002C\u0007C\u0002D\u0007D\u0002E\u0007E\u0002"+
		"F\u0007F\u0002G\u0007G\u0002H\u0007H\u0001\u0000\u0001\u0000\u0001\u0000"+
		"\u0003\u0000\u0096\b\u0000\u0001\u0000\u0001\u0000\u0001\u0001\u0005\u0001"+
		"\u009b\b\u0001\n\u0001\f\u0001\u009e\t\u0001\u0001\u0002\u0004\u0002\u00a1"+
		"\b\u0002\u000b\u0002\f\u0002\u00a2\u0001\u0003\u0004\u0003\u00a6\b\u0003"+
		"\u000b\u0003\f\u0003\u00a7\u0001\u0004\u0001\u0004\u0001\u0004\u0003\u0004"+
		"\u00ad\b\u0004\u0001\u0004\u0005\u0004\u00b0\b\u0004\n\u0004\f\u0004\u00b3"+
		"\t\u0004\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001"+
		"\u0005\u0005\u0005\u00bb\b\u0005\n\u0005\f\u0005\u00be\t\u0005\u0001\u0006"+
		"\u0005\u0006\u00c1\b\u0006\n\u0006\f\u0006\u00c4\t\u0006\u0001\u0006\u0001"+
		"\u0006\u0001\u0006\u0005\u0006\u00c9\b\u0006\n\u0006\f\u0006\u00cc\t\u0006"+
		"\u0001\u0007\u0001\u0007\u0001\u0007\u0001\u0007\u0003\u0007\u00d2\b\u0007"+
		"\u0001\b\u0001\b\u0001\b\u0003\b\u00d7\b\b\u0001\t\u0001\t\u0001\t\u0001"+
		"\n\u0001\n\u0001\n\u0001\n\u0001\n\u0001\n\u0001\n\u0001\n\u0001\n\u0001"+
		"\n\u0001\n\u0003\n\u00e7\b\n\u0001\u000b\u0001\u000b\u0001\u000b\u0001"+
		"\u000b\u0001\u000b\u0001\u000b\u0001\u000b\u0001\u000b\u0001\u000b\u0001"+
		"\f\u0001\f\u0001\f\u0001\f\u0005\f\u00f6\b\f\n\f\f\f\u00f9\t\f\u0001\f"+
		"\u0001\f\u0001\f\u0001\f\u0004\f\u00ff\b\f\u000b\f\f\f\u0100\u0001\f\u0003"+
		"\f\u0104\b\f\u0001\r\u0001\r\u0001\r\u0001\u000e\u0001\u000e\u0001\u000e"+
		"\u0001\u000f\u0001\u000f\u0001\u000f\u0001\u0010\u0001\u0010\u0001\u0010"+
		"\u0005\u0010\u0112\b\u0010\n\u0010\f\u0010\u0115\t\u0010\u0001\u0011\u0001"+
		"\u0011\u0005\u0011\u0119\b\u0011\n\u0011\f\u0011\u011c\t\u0011\u0001\u0012"+
		"\u0005\u0012\u011f\b\u0012\n\u0012\f\u0012\u0122\t\u0012\u0001\u0012\u0001"+
		"\u0012\u0001\u0012\u0005\u0012\u0127\b\u0012\n\u0012\f\u0012\u012a\t\u0012"+
		"\u0001\u0013\u0005\u0013\u012d\b\u0013\n\u0013\f\u0013\u0130\t\u0013\u0001"+
		"\u0014\u0001\u0014\u0001\u0014\u0001\u0014\u0003\u0014\u0136\b\u0014\u0001"+
		"\u0015\u0001\u0015\u0001\u0015\u0003\u0015\u013b\b\u0015\u0001\u0016\u0001"+
		"\u0016\u0001\u0016\u0001\u0016\u0001\u0016\u0001\u0016\u0001\u0016\u0001"+
		"\u0016\u0001\u0016\u0001\u0016\u0001\u0016\u0001\u0016\u0001\u0016\u0003"+
		"\u0016\u014a\b\u0016\u0001\u0017\u0001\u0017\u0001\u0017\u0003\u0017\u014f"+
		"\b\u0017\u0001\u0018\u0001\u0018\u0001\u0018\u0003\u0018\u0154\b\u0018"+
		"\u0001\u0019\u0001\u0019\u0001\u0019\u0001\u0019\u0001\u0019\u0001\u0019"+
		"\u0001\u001a\u0001\u001a\u0001\u001a\u0001\u001a\u0003\u001a\u0160\b\u001a"+
		"\u0001\u001a\u0001\u001a\u0003\u001a\u0164\b\u001a\u0001\u001a\u0001\u001a"+
		"\u0001\u001a\u0001\u001a\u0001\u001a\u0001\u001b\u0001\u001b\u0001\u001b"+
		"\u0001\u001b\u0001\u001b\u0003\u001b\u0170\b\u001b\u0001\u001c\u0001\u001c"+
		"\u0001\u001c\u0001\u001c\u0001\u001c\u0001\u001c\u0001\u001c\u0005\u001c"+
		"\u0179\b\u001c\n\u001c\f\u001c\u017c\t\u001c\u0001\u001d\u0001\u001d\u0001"+
		"\u001d\u0005\u001d\u0181\b\u001d\n\u001d\f\u001d\u0184\t\u001d\u0001\u001e"+
		"\u0001\u001e\u0001\u001e\u0001\u001e\u0001\u001e\u0001\u001f\u0001\u001f"+
		"\u0001\u001f\u0001 \u0001 \u0003 \u0190\b \u0001!\u0001!\u0001!\u0001"+
		"\"\u0001\"\u0001\"\u0001\"\u0001\"\u0001#\u0001#\u0001#\u0001$\u0001$"+
		"\u0003$\u019f\b$\u0001%\u0001%\u0001&\u0001&\u0001&\u0001&\u0001&\u0001"+
		"&\u0003&\u01a9\b&\u0001&\u0001&\u0003&\u01ad\b&\u0001\'\u0001\'\u0001"+
		"\'\u0001\'\u0001\'\u0001(\u0001(\u0001)\u0001)\u0001*\u0001*\u0001*\u0001"+
		"*\u0001*\u0001*\u0001*\u0003*\u01bf\b*\u0001*\u0001*\u0001+\u0001+\u0001"+
		"+\u0001+\u0001+\u0001+\u0001,\u0001,\u0001,\u0001,\u0001,\u0001,\u0001"+
		",\u0001,\u0001,\u0001-\u0001-\u0001-\u0005-\u01d5\b-\n-\f-\u01d8\t-\u0001"+
		".\u0001.\u0001.\u0001.\u0003.\u01de\b.\u0001/\u0001/\u00010\u00010\u0001"+
		"0\u00050\u01e5\b0\n0\f0\u01e8\t0\u00011\u00011\u00011\u00051\u01ed\b1"+
		"\n1\f1\u01f0\t1\u00012\u00012\u00012\u00052\u01f5\b2\n2\f2\u01f8\t2\u0001"+
		"3\u00013\u00013\u00053\u01fd\b3\n3\f3\u0200\t3\u00014\u00014\u00014\u0005"+
		"4\u0205\b4\n4\f4\u0208\t4\u00015\u00015\u00015\u00055\u020d\b5\n5\f5\u0210"+
		"\t5\u00016\u00016\u00016\u00056\u0215\b6\n6\f6\u0218\t6\u00017\u00017"+
		"\u00017\u00057\u021d\b7\n7\f7\u0220\t7\u00018\u00018\u00018\u00058\u0225"+
		"\b8\n8\f8\u0228\t8\u00019\u00019\u00019\u00019\u00019\u00039\u022f\b9"+
		"\u0001:\u0001:\u0001:\u0003:\u0234\b:\u0001;\u0001;\u0001;\u0001;\u0001"+
		";\u0005;\u023b\b;\n;\f;\u023e\t;\u0001<\u0001<\u0001<\u0001<\u0001<\u0001"+
		"<\u0001<\u0001<\u0001<\u0001<\u0001<\u0001<\u0001<\u0001<\u0003<\u024e"+
		"\b<\u0001=\u0001=\u0001=\u0001=\u0001=\u0001=\u0001=\u0001=\u0001=\u0001"+
		"=\u0003=\u025a\b=\u0001=\u0001=\u0001=\u0001=\u0001>\u0001>\u0001>\u0001"+
		">\u0001>\u0001?\u0001?\u0001?\u0001?\u0001?\u0001?\u0001?\u0003?\u026c"+
		"\b?\u0001@\u0001@\u0001A\u0001A\u0001B\u0001B\u0001B\u0001B\u0001C\u0001"+
		"C\u0001C\u0001C\u0001D\u0003D\u027b\bD\u0001E\u0001E\u0001E\u0005E\u0280"+
		"\bE\nE\fE\u0283\tE\u0001F\u0003F\u0286\bF\u0001G\u0001G\u0001G\u0005G"+
		"\u028b\bG\nG\fG\u028e\tG\u0001H\u0001H\u0001H\u0001H\u0001H\u0000\u0000"+
		"I\u0000\u0002\u0004\u0006\b\n\f\u000e\u0010\u0012\u0014\u0016\u0018\u001a"+
		"\u001c\u001e \"$&(*,.02468:<>@BDFHJLNPRTVXZ\\^`bdfhjlnprtvxz|~\u0080\u0082"+
		"\u0084\u0086\u0088\u008a\u008c\u008e\u0090\u0000\b\u0002\u0000CCbb\u0001"+
		"\u0000[\\\u0001\u0000]`\u0001\u0000GH\u0001\u0000IK\u0004\u0000./88HH"+
		"PP\u0002\u0000$$EE\u0002\u0000\u001d\u001d<<\u02b3\u0000\u0092\u0001\u0000"+
		"\u0000\u0000\u0002\u009c\u0001\u0000\u0000\u0000\u0004\u00a0\u0001\u0000"+
		"\u0000\u0000\u0006\u00a5\u0001\u0000\u0000\u0000\b\u00ac\u0001\u0000\u0000"+
		"\u0000\n\u00b4\u0001\u0000\u0000\u0000\f\u00c2\u0001\u0000\u0000\u0000"+
		"\u000e\u00d1\u0001\u0000\u0000\u0000\u0010\u00d6\u0001\u0000\u0000\u0000"+
		"\u0012\u00d8\u0001\u0000\u0000\u0000\u0014\u00e6\u0001\u0000\u0000\u0000"+
		"\u0016\u00e8\u0001\u0000\u0000\u0000\u0018\u0103\u0001\u0000\u0000\u0000"+
		"\u001a\u0105\u0001\u0000\u0000\u0000\u001c\u0108\u0001\u0000\u0000\u0000"+
		"\u001e\u010b\u0001\u0000\u0000\u0000 \u010e\u0001\u0000\u0000\u0000\""+
		"\u011a\u0001\u0000\u0000\u0000$\u0120\u0001\u0000\u0000\u0000&\u012e\u0001"+
		"\u0000\u0000\u0000(\u0135\u0001\u0000\u0000\u0000*\u013a\u0001\u0000\u0000"+
		"\u0000,\u0149\u0001\u0000\u0000\u0000.\u014e\u0001\u0000\u0000\u00000"+
		"\u0150\u0001\u0000\u0000\u00002\u0155\u0001\u0000\u0000\u00004\u015b\u0001"+
		"\u0000\u0000\u00006\u016a\u0001\u0000\u0000\u00008\u0171\u0001\u0000\u0000"+
		"\u0000:\u017d\u0001\u0000\u0000\u0000<\u0185\u0001\u0000\u0000\u0000>"+
		"\u018a\u0001\u0000\u0000\u0000@\u018d\u0001\u0000\u0000\u0000B\u0191\u0001"+
		"\u0000\u0000\u0000D\u0194\u0001\u0000\u0000\u0000F\u0199\u0001\u0000\u0000"+
		"\u0000H\u019c\u0001\u0000\u0000\u0000J\u01a0\u0001\u0000\u0000\u0000L"+
		"\u01a2\u0001\u0000\u0000\u0000N\u01ae\u0001\u0000\u0000\u0000P\u01b3\u0001"+
		"\u0000\u0000\u0000R\u01b5\u0001\u0000\u0000\u0000T\u01b7\u0001\u0000\u0000"+
		"\u0000V\u01c2\u0001\u0000\u0000\u0000X\u01c8\u0001\u0000\u0000\u0000Z"+
		"\u01d1\u0001\u0000\u0000\u0000\\\u01dd\u0001\u0000\u0000\u0000^\u01df"+
		"\u0001\u0000\u0000\u0000`\u01e1\u0001\u0000\u0000\u0000b\u01e9\u0001\u0000"+
		"\u0000\u0000d\u01f1\u0001\u0000\u0000\u0000f\u01f9\u0001\u0000\u0000\u0000"+
		"h\u0201\u0001\u0000\u0000\u0000j\u0209\u0001\u0000\u0000\u0000l\u0211"+
		"\u0001\u0000\u0000\u0000n\u0219\u0001\u0000\u0000\u0000p\u0221\u0001\u0000"+
		"\u0000\u0000r\u022e\u0001\u0000\u0000\u0000t\u0230\u0001\u0000\u0000\u0000"+
		"v\u0235\u0001\u0000\u0000\u0000x\u024d\u0001\u0000\u0000\u0000z\u0259"+
		"\u0001\u0000\u0000\u0000|\u025f\u0001\u0000\u0000\u0000~\u026b\u0001\u0000"+
		"\u0000\u0000\u0080\u026d\u0001\u0000\u0000\u0000\u0082\u026f\u0001\u0000"+
		"\u0000\u0000\u0084\u0271\u0001\u0000\u0000\u0000\u0086\u0275\u0001\u0000"+
		"\u0000\u0000\u0088\u027a\u0001\u0000\u0000\u0000\u008a\u027c\u0001\u0000"+
		"\u0000\u0000\u008c\u0285\u0001\u0000\u0000\u0000\u008e\u0287\u0001\u0000"+
		"\u0000\u0000\u0090\u028f\u0001\u0000\u0000\u0000\u0092\u0095\u0003\u0002"+
		"\u0001\u0000\u0093\u0096\u0003\u0004\u0002\u0000\u0094\u0096\u0003\u0006"+
		"\u0003\u0000\u0095\u0093\u0001\u0000\u0000\u0000\u0095\u0094\u0001\u0000"+
		"\u0000\u0000\u0095\u0096\u0001\u0000\u0000\u0000\u0096\u0097\u0001\u0000"+
		"\u0000\u0000\u0097\u0098\u0005\u0000\u0000\u0001\u0098\u0001\u0001\u0000"+
		"\u0000\u0000\u0099\u009b\u0007\u0000\u0000\u0000\u009a\u0099\u0001\u0000"+
		"\u0000\u0000\u009b\u009e\u0001\u0000\u0000\u0000\u009c\u009a\u0001\u0000"+
		"\u0000\u0000\u009c\u009d\u0001\u0000\u0000\u0000\u009d\u0003\u0001\u0000"+
		"\u0000\u0000\u009e\u009c\u0001\u0000\u0000\u0000\u009f\u00a1\u0003\b\u0004"+
		"\u0000\u00a0\u009f\u0001\u0000\u0000\u0000\u00a1\u00a2\u0001\u0000\u0000"+
		"\u0000\u00a2\u00a0\u0001\u0000\u0000\u0000\u00a2\u00a3\u0001\u0000\u0000"+
		"\u0000\u00a3\u0005\u0001\u0000\u0000\u0000\u00a4\u00a6\u0003\n\u0005\u0000"+
		"\u00a5\u00a4\u0001\u0000\u0000\u0000\u00a6\u00a7\u0001\u0000\u0000\u0000"+
		"\u00a7\u00a5\u0001\u0000\u0000\u0000\u00a7\u00a8\u0001\u0000\u0000\u0000"+
		"\u00a8\u0007\u0001\u0000\u0000\u0000\u00a9\u00ad\u0003\u0016\u000b\u0000"+
		"\u00aa\u00ab\u00050\u0000\u0000\u00ab\u00ad\u00034\u001a\u0000\u00ac\u00a9"+
		"\u0001\u0000\u0000\u0000\u00ac\u00aa\u0001\u0000\u0000\u0000\u00ad\u00b1"+
		"\u0001\u0000\u0000\u0000\u00ae\u00b0\u0005b\u0000\u0000\u00af\u00ae\u0001"+
		"\u0000\u0000\u0000\u00b0\u00b3\u0001\u0000\u0000\u0000\u00b1\u00af\u0001"+
		"\u0000\u0000\u0000\u00b1\u00b2\u0001\u0000\u0000\u0000\u00b2\t\u0001\u0000"+
		"\u0000\u0000\u00b3\u00b1\u0001\u0000\u0000\u0000\u00b4\u00b5\u0005\f\u0000"+
		"\u0000\u00b5\u00b6\u0005b\u0000\u0000\u00b6\u00b7\u0003\"\u0011\u0000"+
		"\u00b7\u00b8\u0003\f\u0006\u0000\u00b8\u00bc\u0005\u0013\u0000\u0000\u00b9"+
		"\u00bb\u0005b\u0000\u0000\u00ba\u00b9\u0001\u0000\u0000\u0000\u00bb\u00be"+
		"\u0001\u0000\u0000\u0000\u00bc\u00ba\u0001\u0000\u0000\u0000\u00bc\u00bd"+
		"\u0001\u0000\u0000\u0000\u00bd\u000b\u0001\u0000\u0000\u0000\u00be\u00bc"+
		"\u0001\u0000\u0000\u0000\u00bf\u00c1\u0005b\u0000\u0000\u00c0\u00bf\u0001"+
		"\u0000\u0000\u0000\u00c1\u00c4\u0001\u0000\u0000\u0000\u00c2\u00c0\u0001"+
		"\u0000\u0000\u0000\u00c2\u00c3\u0001\u0000\u0000\u0000\u00c3\u00c5\u0001"+
		"\u0000\u0000\u0000\u00c4\u00c2\u0001\u0000\u0000\u0000\u00c5\u00c6\u0003"+
		"\u0010\b\u0000\u00c6\u00ca\u0005b\u0000\u0000\u00c7\u00c9\u0003\u000e"+
		"\u0007\u0000\u00c8\u00c7\u0001\u0000\u0000\u0000\u00c9\u00cc\u0001\u0000"+
		"\u0000\u0000\u00ca\u00c8\u0001\u0000\u0000\u0000\u00ca\u00cb\u0001\u0000"+
		"\u0000\u0000\u00cb\r\u0001\u0000\u0000\u0000\u00cc\u00ca\u0001\u0000\u0000"+
		"\u0000\u00cd\u00ce\u0003\u0010\b\u0000\u00ce\u00cf\u0005b\u0000\u0000"+
		"\u00cf\u00d2\u0001\u0000\u0000\u0000\u00d0\u00d2\u0005b\u0000\u0000\u00d1"+
		"\u00cd\u0001\u0000\u0000\u0000\u00d1\u00d0\u0001\u0000\u0000\u0000\u00d2"+
		"\u000f\u0001\u0000\u0000\u0000\u00d3\u00d7\u0003\u0014\n\u0000\u00d4\u00d7"+
		"\u0003.\u0017\u0000\u00d5\u00d7\u0003\u0012\t\u0000\u00d6\u00d3\u0001"+
		"\u0000\u0000\u0000\u00d6\u00d4\u0001\u0000\u0000\u0000\u00d6\u00d5\u0001"+
		"\u0000\u0000\u0000\u00d7\u0011\u0001\u0000\u0000\u0000\u00d8\u00d9\u0005"+
		"0\u0000\u0000\u00d9\u00da\u00032\u0019\u0000\u00da\u0013\u0001\u0000\u0000"+
		"\u0000\u00db\u00e7\u0003<\u001e\u0000\u00dc\u00e7\u0003>\u001f\u0000\u00dd"+
		"\u00e7\u0003B!\u0000\u00de\u00e7\u0003D\"\u0000\u00df\u00e7\u0003F#\u0000"+
		"\u00e0\u00e7\u0003H$\u0000\u00e1\u00e7\u00036\u001b\u0000\u00e2\u00e7"+
		"\u0003L&\u0000\u00e3\u00e7\u0003N\'\u0000\u00e4\u00e7\u0003P(\u0000\u00e5"+
		"\u00e7\u0003R)\u0000\u00e6\u00db\u0001\u0000\u0000\u0000\u00e6\u00dc\u0001"+
		"\u0000\u0000\u0000\u00e6\u00dd\u0001\u0000\u0000\u0000\u00e6\u00de\u0001"+
		"\u0000\u0000\u0000\u00e6\u00df\u0001\u0000\u0000\u0000\u00e6\u00e0\u0001"+
		"\u0000\u0000\u0000\u00e6\u00e1\u0001\u0000\u0000\u0000\u00e6\u00e2\u0001"+
		"\u0000\u0000\u0000\u00e6\u00e3\u0001\u0000\u0000\u0000\u00e6\u00e4\u0001"+
		"\u0000\u0000\u0000\u00e6\u00e5\u0001\u0000\u0000\u0000\u00e7\u0015\u0001"+
		"\u0000\u0000\u0000\u00e8\u00e9\u0005\u001f\u0000\u0000\u00e9\u00ea\u0005"+
		"E\u0000\u0000\u00ea\u00eb\u0003\u0018\f\u0000\u00eb\u00ec\u0005(\u0000"+
		"\u0000\u00ec\u00ed\u0005b\u0000\u0000\u00ed\u00ee\u0003\"\u0011\u0000"+
		"\u00ee\u00ef\u0003$\u0012\u0000\u00ef\u00f0\u0005\u0015\u0000\u0000\u00f0"+
		"\u0017\u0001\u0000\u0000\u0000\u00f1\u00f7\u0005Q\u0000\u0000\u00f2\u00f6"+
		"\u0003\u001a\r\u0000\u00f3\u00f6\u0003\u001c\u000e\u0000\u00f4\u00f6\u0003"+
		"\u001e\u000f\u0000\u00f5\u00f2\u0001\u0000\u0000\u0000\u00f5\u00f3\u0001"+
		"\u0000\u0000\u0000\u00f5\u00f4\u0001\u0000\u0000\u0000\u00f6\u00f9\u0001"+
		"\u0000\u0000\u0000\u00f7\u00f5\u0001\u0000\u0000\u0000\u00f7\u00f8\u0001"+
		"\u0000\u0000\u0000\u00f8\u00fa\u0001\u0000\u0000\u0000\u00f9\u00f7\u0001"+
		"\u0000\u0000\u0000\u00fa\u0104\u0005R\u0000\u0000\u00fb\u00ff\u0003\u001a"+
		"\r\u0000\u00fc\u00ff\u0003\u001c\u000e\u0000\u00fd\u00ff\u0003\u001e\u000f"+
		"\u0000\u00fe\u00fb\u0001\u0000\u0000\u0000\u00fe\u00fc\u0001\u0000\u0000"+
		"\u0000\u00fe\u00fd\u0001\u0000\u0000\u0000\u00ff\u0100\u0001\u0000\u0000"+
		"\u0000\u0100\u00fe\u0001\u0000\u0000\u0000\u0100\u0101\u0001\u0000\u0000"+
		"\u0000\u0101\u0104\u0001\u0000\u0000\u0000\u0102\u0104\u0001\u0000\u0000"+
		"\u0000\u0103\u00f1\u0001\u0000\u0000\u0000\u0103\u00fe\u0001\u0000\u0000"+
		"\u0000\u0103\u0102\u0001\u0000\u0000\u0000\u0104\u0019\u0001\u0000\u0000"+
		"\u0000\u0105\u0106\u0005,\u0000\u0000\u0106\u0107\u0003 \u0010\u0000\u0107"+
		"\u001b\u0001\u0000\u0000\u0000\u0108\u0109\u00051\u0000\u0000\u0109\u010a"+
		"\u0003 \u0010\u0000\u010a\u001d\u0001\u0000\u0000\u0000\u010b\u010c\u0005"+
		"5\u0000\u0000\u010c\u010d\u0003 \u0010\u0000\u010d\u001f\u0001\u0000\u0000"+
		"\u0000\u010e\u0113\u0005E\u0000\u0000\u010f\u0110\u0005S\u0000\u0000\u0110"+
		"\u0112\u0005E\u0000\u0000\u0111\u010f\u0001\u0000\u0000\u0000\u0112\u0115"+
		"\u0001\u0000\u0000\u0000\u0113\u0111\u0001\u0000\u0000\u0000\u0113\u0114"+
		"\u0001\u0000\u0000\u0000\u0114!\u0001\u0000\u0000\u0000\u0115\u0113\u0001"+
		"\u0000\u0000\u0000\u0116\u0117\u0005C\u0000\u0000\u0117\u0119\u0005b\u0000"+
		"\u0000\u0118\u0116\u0001\u0000\u0000\u0000\u0119\u011c\u0001\u0000\u0000"+
		"\u0000\u011a\u0118\u0001\u0000\u0000\u0000\u011a\u011b\u0001\u0000\u0000"+
		"\u0000\u011b#\u0001\u0000\u0000\u0000\u011c\u011a\u0001\u0000\u0000\u0000"+
		"\u011d\u011f\u0005b\u0000\u0000\u011e\u011d\u0001\u0000\u0000\u0000\u011f"+
		"\u0122\u0001\u0000\u0000\u0000\u0120\u011e\u0001\u0000\u0000\u0000\u0120"+
		"\u0121\u0001\u0000\u0000\u0000\u0121\u0123\u0001\u0000\u0000\u0000\u0122"+
		"\u0120\u0001\u0000\u0000\u0000\u0123\u0124\u0003*\u0015\u0000\u0124\u0128"+
		"\u0005b\u0000\u0000\u0125\u0127\u0003(\u0014\u0000\u0126\u0125\u0001\u0000"+
		"\u0000\u0000\u0127\u012a\u0001\u0000\u0000\u0000\u0128\u0126\u0001\u0000"+
		"\u0000\u0000\u0128\u0129\u0001\u0000\u0000\u0000\u0129%\u0001\u0000\u0000"+
		"\u0000\u012a\u0128\u0001\u0000\u0000\u0000\u012b\u012d\u0003(\u0014\u0000"+
		"\u012c\u012b\u0001\u0000\u0000\u0000\u012d\u0130\u0001\u0000\u0000\u0000"+
		"\u012e\u012c\u0001\u0000\u0000\u0000\u012e\u012f\u0001\u0000\u0000\u0000"+
		"\u012f\'\u0001\u0000\u0000\u0000\u0130\u012e\u0001\u0000\u0000\u0000\u0131"+
		"\u0132\u0003*\u0015\u0000\u0132\u0133\u0005b\u0000\u0000\u0133\u0136\u0001"+
		"\u0000\u0000\u0000\u0134\u0136\u0005b\u0000\u0000\u0135\u0131\u0001\u0000"+
		"\u0000\u0000\u0135\u0134\u0001\u0000\u0000\u0000\u0136)\u0001\u0000\u0000"+
		"\u0000\u0137\u013b\u0003,\u0016\u0000\u0138\u013b\u0003.\u0017\u0000\u0139"+
		"\u013b\u00030\u0018\u0000\u013a\u0137\u0001\u0000\u0000\u0000\u013a\u0138"+
		"\u0001\u0000\u0000\u0000\u013a\u0139\u0001\u0000\u0000\u0000\u013b+\u0001"+
		"\u0000\u0000\u0000\u013c\u014a\u0003<\u001e\u0000\u013d\u014a\u0003>\u001f"+
		"\u0000\u013e\u014a\u0003@ \u0000\u013f\u014a\u0003B!\u0000\u0140\u014a"+
		"\u0003D\"\u0000\u0141\u014a\u0003F#\u0000\u0142\u014a\u0003H$\u0000\u0143"+
		"\u014a\u0003J%\u0000\u0144\u014a\u00036\u001b\u0000\u0145\u014a\u0003"+
		"L&\u0000\u0146\u014a\u0003N\'\u0000\u0147\u014a\u0003P(\u0000\u0148\u014a"+
		"\u0003R)\u0000\u0149\u013c\u0001\u0000\u0000\u0000\u0149\u013d\u0001\u0000"+
		"\u0000\u0000\u0149\u013e\u0001\u0000\u0000\u0000\u0149\u013f\u0001\u0000"+
		"\u0000\u0000\u0149\u0140\u0001\u0000\u0000\u0000\u0149\u0141\u0001\u0000"+
		"\u0000\u0000\u0149\u0142\u0001\u0000\u0000\u0000\u0149\u0143\u0001\u0000"+
		"\u0000\u0000\u0149\u0144\u0001\u0000\u0000\u0000\u0149\u0145\u0001\u0000"+
		"\u0000\u0000\u0149\u0146\u0001\u0000\u0000\u0000\u0149\u0147\u0001\u0000"+
		"\u0000\u0000\u0149\u0148\u0001\u0000\u0000\u0000\u014a-\u0001\u0000\u0000"+
		"\u0000\u014b\u014f\u0003T*\u0000\u014c\u014f\u0003V+\u0000\u014d\u014f"+
		"\u0003X,\u0000\u014e\u014b\u0001\u0000\u0000\u0000\u014e\u014c\u0001\u0000"+
		"\u0000\u0000\u014e\u014d\u0001\u0000\u0000\u0000\u014f/\u0001\u0000\u0000"+
		"\u0000\u0150\u0153\u00050\u0000\u0000\u0151\u0154\u00032\u0019\u0000\u0152"+
		"\u0154\u00034\u001a\u0000\u0153\u0151\u0001\u0000\u0000\u0000\u0153\u0152"+
		"\u0001\u0000\u0000\u0000\u01541\u0001\u0000\u0000\u0000\u0155\u0156\u0005"+
		"\u0019\u0000\u0000\u0156\u0157\u0005\u000f\u0000\u0000\u0157\u0158\u0005"+
		"b\u0000\u0000\u0158\u0159\u0003$\u0012\u0000\u0159\u015a\u0005\u0017\u0000"+
		"\u0000\u015a3\u0001\u0000\u0000\u0000\u015b\u015c\u0005\u001b\u0000\u0000"+
		"\u015c\u015f\u0003^/\u0000\u015d\u015e\u0005+\u0000\u0000\u015e\u0160"+
		"\u0005A\u0000\u0000\u015f\u015d\u0001\u0000\u0000\u0000\u015f\u0160\u0001"+
		"\u0000\u0000\u0000\u0160\u0163\u0001\u0000\u0000\u0000\u0161\u0162\u0005"+
		"\u0004\u0000\u0000\u0162\u0164\u0005E\u0000\u0000\u0163\u0161\u0001\u0000"+
		"\u0000\u0000\u0163\u0164\u0001\u0000\u0000\u0000\u0164\u0165\u0001\u0000"+
		"\u0000\u0000\u0165\u0166\u0005\u000f\u0000\u0000\u0166\u0167\u0005b\u0000"+
		"\u0000\u0167\u0168\u0003$\u0012\u0000\u0168\u0169\u0005\u0017\u0000\u0000"+
		"\u01695\u0001\u0000\u0000\u0000\u016a\u016b\u0005\n\u0000\u0000\u016b"+
		"\u016f\u0005\u001b\u0000\u0000\u016c\u0170\u0003^/\u0000\u016d\u016e\u0005"+
		"+\u0000\u0000\u016e\u0170\u0005A\u0000\u0000\u016f\u016c\u0001\u0000\u0000"+
		"\u0000\u016f\u016d\u0001\u0000\u0000\u0000\u01707\u0001\u0000\u0000\u0000"+
		"\u0171\u017a\u0005E\u0000\u0000\u0172\u0173\u0005T\u0000\u0000\u0173\u0174"+
		"\u0003^/\u0000\u0174\u0175\u0005U\u0000\u0000\u0175\u0179\u0001\u0000"+
		"\u0000\u0000\u0176\u0177\u0005Y\u0000\u0000\u0177\u0179\u0005E\u0000\u0000"+
		"\u0178\u0172\u0001\u0000\u0000\u0000\u0178\u0176\u0001\u0000\u0000\u0000"+
		"\u0179\u017c\u0001\u0000\u0000\u0000\u017a\u0178\u0001\u0000\u0000\u0000"+
		"\u017a\u017b\u0001\u0000\u0000\u0000\u017b9\u0001\u0000\u0000\u0000\u017c"+
		"\u017a\u0001\u0000\u0000\u0000\u017d\u0182\u00038\u001c\u0000\u017e\u017f"+
		"\u0005S\u0000\u0000\u017f\u0181\u00038\u001c\u0000\u0180\u017e\u0001\u0000"+
		"\u0000\u0000\u0181\u0184\u0001\u0000\u0000\u0000\u0182\u0180\u0001\u0000"+
		"\u0000\u0000\u0182\u0183\u0001\u0000\u0000\u0000\u0183;\u0001\u0000\u0000"+
		"\u0000\u0184\u0182\u0001\u0000\u0000\u0000\u0185\u0186\u00056\u0000\u0000"+
		"\u0186\u0187\u0003:\u001d\u0000\u0187\u0188\u0005F\u0000\u0000\u0188\u0189"+
		"\u0003^/\u0000\u0189=\u0001\u0000\u0000\u0000\u018a\u018b\u0005\t\u0000"+
		"\u0000\u018b\u018c\u0003z=\u0000\u018c?\u0001\u0000\u0000\u0000\u018d"+
		"\u018f\u00054\u0000\u0000\u018e\u0190\u0003\u008aE\u0000\u018f\u018e\u0001"+
		"\u0000\u0000\u0000\u018f\u0190\u0001\u0000\u0000\u0000\u0190A\u0001\u0000"+
		"\u0000\u0000\u0191\u0192\u0005\u0012\u0000\u0000\u0192\u0193\u0003^/\u0000"+
		"\u0193C\u0001\u0000\u0000\u0000\u0194\u0195\u0005?\u0000\u0000\u0195\u0196"+
		"\u0003^/\u0000\u0196\u0197\u0005S\u0000\u0000\u0197\u0198\u0003^/\u0000"+
		"\u0198E\u0001\u0000\u0000\u0000\u0199\u019a\u0005)\u0000\u0000\u019a\u019b"+
		"\u0003^/\u0000\u019bG\u0001\u0000\u0000\u0000\u019c\u019e\u0005\u001c"+
		"\u0000\u0000\u019d\u019f\u0003^/\u0000\u019e\u019d\u0001\u0000\u0000\u0000"+
		"\u019e\u019f\u0001\u0000\u0000\u0000\u019fI\u0001\u0000\u0000\u0000\u01a0"+
		"\u01a1\u0005\u000b\u0000\u0000\u01a1K\u0001\u0000\u0000\u0000\u01a2\u01a3"+
		"\u0005\u0006\u0000\u0000\u01a3\u01a4\u0003^/\u0000\u01a4\u01a5\u0005S"+
		"\u0000\u0000\u01a5\u01a8\u0003^/\u0000\u01a6\u01a7\u0005@\u0000\u0000"+
		"\u01a7\u01a9\u0003^/\u0000\u01a8\u01a6\u0001\u0000\u0000\u0000\u01a8\u01a9"+
		"\u0001\u0000\u0000\u0000\u01a9\u01ac\u0001\u0000\u0000\u0000\u01aa\u01ab"+
		"\u0005#\u0000\u0000\u01ab\u01ad\u00038\u001c\u0000\u01ac\u01aa\u0001\u0000"+
		"\u0000\u0000\u01ac\u01ad\u0001\u0000\u0000\u0000\u01adM\u0001\u0000\u0000"+
		"\u0000\u01ae\u01af\u00053\u0000\u0000\u01af\u01b0\u0003^/\u0000\u01b0"+
		"\u01b1\u0005#\u0000\u0000\u01b1\u01b2\u00038\u001c\u0000\u01b2O\u0001"+
		"\u0000\u0000\u0000\u01b3\u01b4\u0005\b\u0000\u0000\u01b4Q\u0001\u0000"+
		"\u0000\u0000\u01b5\u01b6\u0005\r\u0000\u0000\u01b6S\u0001\u0000\u0000"+
		"\u0000\u01b7\u01b8\u0005!\u0000\u0000\u01b8\u01b9\u0003^/\u0000\u01b9"+
		"\u01ba\u0005b\u0000\u0000\u01ba\u01be\u0003$\u0012\u0000\u01bb\u01bc\u0005"+
		"\u0011\u0000\u0000\u01bc\u01bd\u0005b\u0000\u0000\u01bd\u01bf\u0003$\u0012"+
		"\u0000\u01be\u01bb\u0001\u0000\u0000\u0000\u01be\u01bf\u0001\u0000\u0000"+
		"\u0000\u01bf\u01c0\u0001\u0000\u0000\u0000\u01c0\u01c1\u0005\u0016\u0000"+
		"\u0000\u01c1U\u0001\u0000\u0000\u0000\u01c2\u01c3\u0005>\u0000\u0000\u01c3"+
		"\u01c4\u0003^/\u0000\u01c4\u01c5\u0005b\u0000\u0000\u01c5\u01c6\u0003"+
		"$\u0012\u0000\u01c6\u01c7\u0005\u0018\u0000\u0000\u01c7W\u0001\u0000\u0000"+
		"\u0000\u01c8\u01c9\u0005\u001e\u0000\u0000\u01c9\u01ca\u0005\u0010\u0000"+
		"\u0000\u01ca\u01cb\u0005E\u0000\u0000\u01cb\u01cc\u0005\"\u0000\u0000"+
		"\u01cc\u01cd\u0003^/\u0000\u01cd\u01ce\u0005b\u0000\u0000\u01ce\u01cf"+
		"\u0003$\u0012\u0000\u01cf\u01d0\u0005\u0014\u0000\u0000\u01d0Y\u0001\u0000"+
		"\u0000\u0000\u01d1\u01d6\u0005E\u0000\u0000\u01d2\u01d3\u0005Y\u0000\u0000"+
		"\u01d3\u01d5\u0005E\u0000\u0000\u01d4\u01d2\u0001\u0000\u0000\u0000\u01d5"+
		"\u01d8\u0001\u0000\u0000\u0000\u01d6\u01d4\u0001\u0000\u0000\u0000\u01d6"+
		"\u01d7\u0001\u0000\u0000\u0000\u01d7[\u0001\u0000\u0000\u0000\u01d8\u01d6"+
		"\u0001\u0000\u0000\u0000\u01d9\u01de\u0005E\u0000\u0000\u01da\u01db\u0005"+
		";\u0000\u0000\u01db\u01dc\u0005Y\u0000\u0000\u01dc\u01de\u0003Z-\u0000"+
		"\u01dd\u01d9\u0001\u0000\u0000\u0000\u01dd\u01da\u0001\u0000\u0000\u0000"+
		"\u01de]\u0001\u0000\u0000\u0000\u01df\u01e0\u0003`0\u0000\u01e0_\u0001"+
		"\u0000\u0000\u0000\u01e1\u01e6\u0003b1\u0000\u01e2\u01e3\u00052\u0000"+
		"\u0000\u01e3\u01e5\u0003b1\u0000\u01e4\u01e2\u0001\u0000\u0000\u0000\u01e5"+
		"\u01e8\u0001\u0000\u0000\u0000\u01e6\u01e4\u0001\u0000\u0000\u0000\u01e6"+
		"\u01e7\u0001\u0000\u0000\u0000\u01e7a\u0001\u0000\u0000\u0000\u01e8\u01e6"+
		"\u0001\u0000\u0000\u0000\u01e9\u01ee\u0003d2\u0000\u01ea\u01eb\u0005\u0003"+
		"\u0000\u0000\u01eb\u01ed\u0003d2\u0000\u01ec\u01ea\u0001\u0000\u0000\u0000"+
		"\u01ed\u01f0\u0001\u0000\u0000\u0000\u01ee\u01ec\u0001\u0000\u0000\u0000"+
		"\u01ee\u01ef\u0001\u0000\u0000\u0000\u01efc\u0001\u0000\u0000\u0000\u01f0"+
		"\u01ee\u0001\u0000\u0000\u0000\u01f1\u01f6\u0003f3\u0000\u01f2\u01f3\u0005"+
		"N\u0000\u0000\u01f3\u01f5\u0003f3\u0000\u01f4\u01f2\u0001\u0000\u0000"+
		"\u0000\u01f5\u01f8\u0001\u0000\u0000\u0000\u01f6\u01f4\u0001\u0000\u0000"+
		"\u0000\u01f6\u01f7\u0001\u0000\u0000\u0000\u01f7e\u0001\u0000\u0000\u0000"+
		"\u01f8\u01f6\u0001\u0000\u0000\u0000\u01f9\u01fe\u0003h4\u0000\u01fa\u01fb"+
		"\u0005O\u0000\u0000\u01fb\u01fd\u0003h4\u0000\u01fc\u01fa\u0001\u0000"+
		"\u0000\u0000\u01fd\u0200\u0001\u0000\u0000\u0000\u01fe\u01fc\u0001\u0000"+
		"\u0000\u0000\u01fe\u01ff\u0001\u0000\u0000\u0000\u01ffg\u0001\u0000\u0000"+
		"\u0000\u0200\u01fe\u0001\u0000\u0000\u0000\u0201\u0206\u0003j5\u0000\u0202"+
		"\u0203\u0005M\u0000\u0000\u0203\u0205\u0003j5\u0000\u0204\u0202\u0001"+
		"\u0000\u0000\u0000\u0205\u0208\u0001\u0000\u0000\u0000\u0206\u0204\u0001"+
		"\u0000\u0000\u0000\u0206\u0207\u0001\u0000\u0000\u0000\u0207i\u0001\u0000"+
		"\u0000\u0000\u0208\u0206\u0001\u0000\u0000\u0000\u0209\u020e\u0003l6\u0000"+
		"\u020a\u020b\u0007\u0001\u0000\u0000\u020b\u020d\u0003l6\u0000\u020c\u020a"+
		"\u0001\u0000\u0000\u0000\u020d\u0210\u0001\u0000\u0000\u0000\u020e\u020c"+
		"\u0001\u0000\u0000\u0000\u020e\u020f\u0001\u0000\u0000\u0000\u020fk\u0001"+
		"\u0000\u0000\u0000\u0210\u020e\u0001\u0000\u0000\u0000\u0211\u0216\u0003"+
		"n7\u0000\u0212\u0213\u0007\u0002\u0000\u0000\u0213\u0215\u0003n7\u0000"+
		"\u0214\u0212\u0001\u0000\u0000\u0000\u0215\u0218\u0001\u0000\u0000\u0000"+
		"\u0216\u0214\u0001\u0000\u0000\u0000\u0216\u0217\u0001\u0000\u0000\u0000"+
		"\u0217m\u0001\u0000\u0000\u0000\u0218\u0216\u0001\u0000\u0000\u0000\u0219"+
		"\u021e\u0003p8\u0000\u021a\u021b\u0007\u0003\u0000\u0000\u021b\u021d\u0003"+
		"p8\u0000\u021c\u021a\u0001\u0000\u0000\u0000\u021d\u0220\u0001\u0000\u0000"+
		"\u0000\u021e\u021c\u0001\u0000\u0000\u0000\u021e\u021f\u0001\u0000\u0000"+
		"\u0000\u021fo\u0001\u0000\u0000\u0000\u0220\u021e\u0001\u0000\u0000\u0000"+
		"\u0221\u0226\u0003r9\u0000\u0222\u0223\u0007\u0004\u0000\u0000\u0223\u0225"+
		"\u0003r9\u0000\u0224\u0222\u0001\u0000\u0000\u0000\u0225\u0228\u0001\u0000"+
		"\u0000\u0000\u0226\u0224\u0001\u0000\u0000\u0000\u0226\u0227\u0001\u0000"+
		"\u0000\u0000\u0227q\u0001\u0000\u0000\u0000\u0228\u0226\u0001\u0000\u0000"+
		"\u0000\u0229\u022a\u0007\u0005\u0000\u0000\u022a\u022f\u0003r9\u0000\u022b"+
		"\u022c\u0005=\u0000\u0000\u022c\u022f\u0003r9\u0000\u022d\u022f\u0003"+
		"t:\u0000\u022e\u0229\u0001\u0000\u0000\u0000\u022e\u022b\u0001\u0000\u0000"+
		"\u0000\u022e\u022d\u0001\u0000\u0000\u0000\u022fs\u0001\u0000\u0000\u0000"+
		"\u0230\u0233\u0003v;\u0000\u0231\u0232\u0005L\u0000\u0000\u0232\u0234"+
		"\u0003t:\u0000\u0233\u0231\u0001\u0000\u0000\u0000\u0233\u0234\u0001\u0000"+
		"\u0000\u0000\u0234u\u0001\u0000\u0000\u0000\u0235\u023c\u0003x<\u0000"+
		"\u0236\u0237\u0005T\u0000\u0000\u0237\u0238\u0003^/\u0000\u0238\u0239"+
		"\u0005U\u0000\u0000\u0239\u023b\u0001\u0000\u0000\u0000\u023a\u0236\u0001"+
		"\u0000\u0000\u0000\u023b\u023e\u0001\u0000\u0000\u0000\u023c\u023a\u0001"+
		"\u0000\u0000\u0000\u023c\u023d\u0001\u0000\u0000\u0000\u023dw\u0001\u0000"+
		"\u0000\u0000\u023e\u023c\u0001\u0000\u0000\u0000\u023f\u024e\u0003~?\u0000"+
		"\u0240\u024e\u0003|>\u0000\u0241\u024e\u0005E\u0000\u0000\u0242\u024e"+
		"\u0005$\u0000\u0000\u0243\u024e\u0003z=\u0000\u0244\u0245\u0005\u001a"+
		"\u0000\u0000\u0245\u0246\u0005Q\u0000\u0000\u0246\u0247\u0003^/\u0000"+
		"\u0247\u0248\u0005R\u0000\u0000\u0248\u024e\u0001\u0000\u0000\u0000\u0249"+
		"\u024a\u0005Q\u0000\u0000\u024a\u024b\u0003^/\u0000\u024b\u024c\u0005"+
		"R\u0000\u0000\u024c\u024e\u0001\u0000\u0000\u0000\u024d\u023f\u0001\u0000"+
		"\u0000\u0000\u024d\u0240\u0001\u0000\u0000\u0000\u024d\u0241\u0001\u0000"+
		"\u0000\u0000\u024d\u0242\u0001\u0000\u0000\u0000\u024d\u0243\u0001\u0000"+
		"\u0000\u0000\u024d\u0244\u0001\u0000\u0000\u0000\u024d\u0249\u0001\u0000"+
		"\u0000\u0000\u024ey\u0001\u0000\u0000\u0000\u024f\u025a\u0003\\.\u0000"+
		"\u0250\u025a\u0005&\u0000\u0000\u0251\u025a\u0005\'\u0000\u0000\u0252"+
		"\u025a\u00057\u0000\u0000\u0253\u025a\u0005\u000e\u0000\u0000\u0254\u025a"+
		"\u00059\u0000\u0000\u0255\u025a\u0005\u0005\u0000\u0000\u0256\u025a\u0005"+
		"\u0002\u0000\u0000\u0257\u025a\u0005\u0007\u0000\u0000\u0258\u025a\u0005"+
		"%\u0000\u0000\u0259\u024f\u0001\u0000\u0000\u0000\u0259\u0250\u0001\u0000"+
		"\u0000\u0000\u0259\u0251\u0001\u0000\u0000\u0000\u0259\u0252\u0001\u0000"+
		"\u0000\u0000\u0259\u0253\u0001\u0000\u0000\u0000\u0259\u0254\u0001\u0000"+
		"\u0000\u0000\u0259\u0255\u0001\u0000\u0000\u0000\u0259\u0256\u0001\u0000"+
		"\u0000\u0000\u0259\u0257\u0001\u0000\u0000\u0000\u0259\u0258\u0001\u0000"+
		"\u0000\u0000\u025a\u025b\u0001\u0000\u0000\u0000\u025b\u025c\u0005Q\u0000"+
		"\u0000\u025c\u025d\u0003\u0088D\u0000\u025d\u025e\u0005R\u0000\u0000\u025e"+
		"{\u0001\u0000\u0000\u0000\u025f\u0260\u0005Z\u0000\u0000\u0260\u0261\u0007"+
		"\u0006\u0000\u0000\u0261\u0262\u0005W\u0000\u0000\u0262\u0263\u0005W\u0000"+
		"\u0000\u0263}\u0001\u0000\u0000\u0000\u0264\u026c\u0005A\u0000\u0000\u0265"+
		"\u026c\u0005B\u0000\u0000\u0266\u026c\u0005D\u0000\u0000\u0267\u026c\u0003"+
		"\u0084B\u0000\u0268\u026c\u0003\u0086C\u0000\u0269\u026c\u0003\u0082A"+
		"\u0000\u026a\u026c\u0003\u0080@\u0000\u026b\u0264\u0001\u0000\u0000\u0000"+
		"\u026b\u0265\u0001\u0000\u0000\u0000\u026b\u0266\u0001\u0000\u0000\u0000"+
		"\u026b\u0267\u0001\u0000\u0000\u0000\u026b\u0268\u0001\u0000\u0000\u0000"+
		"\u026b\u0269\u0001\u0000\u0000\u0000\u026b\u026a\u0001\u0000\u0000\u0000"+
		"\u026c\u007f\u0001\u0000\u0000\u0000\u026d\u026e\u0005-\u0000\u0000\u026e"+
		"\u0081\u0001\u0000\u0000\u0000\u026f\u0270\u0007\u0007\u0000\u0000\u0270"+
		"\u0083\u0001\u0000\u0000\u0000\u0271\u0272\u0005T\u0000\u0000\u0272\u0273"+
		"\u0003\u0088D\u0000\u0273\u0274\u0005U\u0000\u0000\u0274\u0085\u0001\u0000"+
		"\u0000\u0000\u0275\u0276\u0005V\u0000\u0000\u0276\u0277\u0003\u008cF\u0000"+
		"\u0277\u0278\u0005W\u0000\u0000\u0278\u0087\u0001\u0000\u0000\u0000\u0279"+
		"\u027b\u0003\u008aE\u0000\u027a\u0279\u0001\u0000\u0000\u0000\u027a\u027b"+
		"\u0001\u0000\u0000\u0000\u027b\u0089\u0001\u0000\u0000\u0000\u027c\u0281"+
		"\u0003^/\u0000\u027d\u027e\u0005S\u0000\u0000\u027e\u0280\u0003^/\u0000"+
		"\u027f\u027d\u0001\u0000\u0000\u0000\u0280\u0283\u0001\u0000\u0000\u0000"+
		"\u0281\u027f\u0001\u0000\u0000\u0000\u0281\u0282\u0001\u0000\u0000\u0000"+
		"\u0282\u008b\u0001\u0000\u0000\u0000\u0283\u0281\u0001\u0000\u0000\u0000"+
		"\u0284\u0286\u0003\u008eG\u0000\u0285\u0284\u0001\u0000\u0000\u0000\u0285"+
		"\u0286\u0001\u0000\u0000\u0000\u0286\u008d\u0001\u0000\u0000\u0000\u0287"+
		"\u028c\u0003\u0090H\u0000\u0288\u0289\u0005S\u0000\u0000\u0289\u028b\u0003"+
		"\u0090H\u0000\u028a\u0288\u0001\u0000\u0000\u0000\u028b\u028e\u0001\u0000"+
		"\u0000\u0000\u028c\u028a\u0001\u0000\u0000\u0000\u028c\u028d\u0001\u0000"+
		"\u0000\u0000\u028d\u008f\u0001\u0000\u0000\u0000\u028e\u028c\u0001\u0000"+
		"\u0000\u0000\u028f\u0290\u0005A\u0000\u0000\u0290\u0291\u0005X\u0000\u0000"+
		"\u0291\u0292\u0003^/\u0000\u0292\u0091\u0001\u0000\u0000\u0000;\u0095"+
		"\u009c\u00a2\u00a7\u00ac\u00b1\u00bc\u00c2\u00ca\u00d1\u00d6\u00e6\u00f5"+
		"\u00f7\u00fe\u0100\u0103\u0113\u011a\u0120\u0128\u012e\u0135\u013a\u0149"+
		"\u014e\u0153\u015f\u0163\u016f\u0178\u017a\u0182\u018f\u019e\u01a8\u01ac"+
		"\u01be\u01d6\u01dd\u01e6\u01ee\u01f6\u01fe\u0206\u020e\u0216\u021e\u0226"+
		"\u022e\u0233\u023c\u024d\u0259\u026b\u027a\u0281\u0285\u028c";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}