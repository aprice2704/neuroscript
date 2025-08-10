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
		KW_WHILE=62, KW_WITH=63, STRING_LIT=64, TRIPLE_BACKTICK_STRING=65, METADATA_LINE=66, 
		NUMBER_LIT=67, IDENTIFIER=68, ASSIGN=69, PLUS=70, MINUS=71, STAR=72, SLASH=73, 
		PERCENT=74, STAR_STAR=75, AMPERSAND=76, PIPE=77, CARET=78, TILDE=79, LPAREN=80, 
		RPAREN=81, COMMA=82, LBRACK=83, RBRACK=84, LBRACE=85, RBRACE=86, COLON=87, 
		DOT=88, PLACEHOLDER_START=89, PLACEHOLDER_END=90, EQ=91, NEQ=92, GT=93, 
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
		RULE_emit_statement = 33, RULE_must_statement = 34, RULE_fail_statement = 35, 
		RULE_clearErrorStmt = 36, RULE_ask_stmt = 37, RULE_promptuser_stmt = 38, 
		RULE_break_statement = 39, RULE_continue_statement = 40, RULE_if_statement = 41, 
		RULE_while_statement = 42, RULE_for_each_statement = 43, RULE_qualified_identifier = 44, 
		RULE_call_target = 45, RULE_expression = 46, RULE_logical_or_expr = 47, 
		RULE_logical_and_expr = 48, RULE_bitwise_or_expr = 49, RULE_bitwise_xor_expr = 50, 
		RULE_bitwise_and_expr = 51, RULE_equality_expr = 52, RULE_relational_expr = 53, 
		RULE_additive_expr = 54, RULE_multiplicative_expr = 55, RULE_unary_expr = 56, 
		RULE_power_expr = 57, RULE_accessor_expr = 58, RULE_primary = 59, RULE_callable_expr = 60, 
		RULE_placeholder = 61, RULE_literal = 62, RULE_nil_literal = 63, RULE_boolean_literal = 64, 
		RULE_list_literal = 65, RULE_map_literal = 66, RULE_expression_list_opt = 67, 
		RULE_expression_list = 68, RULE_map_entry_list_opt = 69, RULE_map_entry_list = 70, 
		RULE_map_entry = 71;
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
			"must_statement", "fail_statement", "clearErrorStmt", "ask_stmt", "promptuser_stmt", 
			"break_statement", "continue_statement", "if_statement", "while_statement", 
			"for_each_statement", "qualified_identifier", "call_target", "expression", 
			"logical_or_expr", "logical_and_expr", "bitwise_or_expr", "bitwise_xor_expr", 
			"bitwise_and_expr", "equality_expr", "relational_expr", "additive_expr", 
			"multiplicative_expr", "unary_expr", "power_expr", "accessor_expr", "primary", 
			"callable_expr", "placeholder", "literal", "nil_literal", "boolean_literal", 
			"list_literal", "map_literal", "expression_list_opt", "expression_list", 
			"map_entry_list_opt", "map_entry_list", "map_entry"
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
			"'tool'", "'true'", "'typeof'", "'while'", "'with'", null, null, null, 
			null, null, "'='", "'+'", "'-'", "'*'", "'/'", "'%'", "'**'", "'&'", 
			"'|'", "'^'", "'~'", "'('", "')'", "','", "'['", "']'", "'{'", "'}'", 
			"':'", "'.'", "'{{'", "'}}'", "'=='", "'!='", "'>'", "'<'", "'>='", "'<='"
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
			"KW_WITH", "STRING_LIT", "TRIPLE_BACKTICK_STRING", "METADATA_LINE", "NUMBER_LIT", 
			"IDENTIFIER", "ASSIGN", "PLUS", "MINUS", "STAR", "SLASH", "PERCENT", 
			"STAR_STAR", "AMPERSAND", "PIPE", "CARET", "TILDE", "LPAREN", "RPAREN", 
			"COMMA", "LBRACK", "RBRACK", "LBRACE", "RBRACE", "COLON", "DOT", "PLACEHOLDER_START", 
			"PLACEHOLDER_END", "EQ", "NEQ", "GT", "LT", "GTE", "LTE", "LINE_COMMENT", 
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
			setState(144);
			file_header();
			setState(147);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_FUNC:
			case KW_ON:
				{
				setState(145);
				library_script();
				}
				break;
			case KW_COMMAND:
				{
				setState(146);
				command_script();
				}
				break;
			case EOF:
				break;
			default:
				break;
			}
			setState(149);
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
			setState(154);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==METADATA_LINE || _la==NEWLINE) {
				{
				{
				setState(151);
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
				setState(156);
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
			setState(158); 
			_errHandler.sync(this);
			_la = _input.LA(1);
			do {
				{
				{
				setState(157);
				library_block();
				}
				}
				setState(160); 
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
			setState(163); 
			_errHandler.sync(this);
			_la = _input.LA(1);
			do {
				{
				{
				setState(162);
				command_block();
				}
				}
				setState(165); 
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
			setState(170);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_FUNC:
				{
				setState(167);
				procedure_definition();
				}
				break;
			case KW_ON:
				{
				setState(168);
				match(KW_ON);
				setState(169);
				event_handler();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			setState(175);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==NEWLINE) {
				{
				{
				setState(172);
				match(NEWLINE);
				}
				}
				setState(177);
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
			setState(178);
			match(KW_COMMAND);
			setState(179);
			match(NEWLINE);
			setState(180);
			metadata_block();
			setState(181);
			command_statement_list();
			setState(182);
			match(KW_ENDCOMMAND);
			setState(186);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==NEWLINE) {
				{
				{
				setState(183);
				match(NEWLINE);
				}
				}
				setState(188);
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
			setState(192);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==NEWLINE) {
				{
				{
				setState(189);
				match(NEWLINE);
				}
				}
				setState(194);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(195);
			command_statement();
			setState(196);
			match(NEWLINE);
			setState(200);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while ((((_la) & ~0x3f) == 0 && ((1L << _la) & 4632235900682905408L) != 0) || _la==NEWLINE) {
				{
				{
				setState(197);
				command_body_line();
				}
				}
				setState(202);
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
			setState(207);
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
				enterOuterAlt(_localctx, 1);
				{
				setState(203);
				command_statement();
				setState(204);
				match(NEWLINE);
				}
				break;
			case NEWLINE:
				enterOuterAlt(_localctx, 2);
				{
				setState(206);
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
			setState(212);
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
				enterOuterAlt(_localctx, 1);
				{
				setState(209);
				simple_command_statement();
				}
				break;
			case KW_FOR:
			case KW_IF:
			case KW_WHILE:
				enterOuterAlt(_localctx, 2);
				{
				setState(210);
				block_statement();
				}
				break;
			case KW_ON:
				enterOuterAlt(_localctx, 3);
				{
				setState(211);
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
			setState(214);
			match(KW_ON);
			setState(215);
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
			setState(227);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_SET:
				enterOuterAlt(_localctx, 1);
				{
				setState(217);
				set_statement();
				}
				break;
			case KW_CALL:
				enterOuterAlt(_localctx, 2);
				{
				setState(218);
				call_statement();
				}
				break;
			case KW_EMIT:
				enterOuterAlt(_localctx, 3);
				{
				setState(219);
				emit_statement();
				}
				break;
			case KW_MUST:
				enterOuterAlt(_localctx, 4);
				{
				setState(220);
				must_statement();
				}
				break;
			case KW_FAIL:
				enterOuterAlt(_localctx, 5);
				{
				setState(221);
				fail_statement();
				}
				break;
			case KW_CLEAR:
				enterOuterAlt(_localctx, 6);
				{
				setState(222);
				clearEventStmt();
				}
				break;
			case KW_ASK:
				enterOuterAlt(_localctx, 7);
				{
				setState(223);
				ask_stmt();
				}
				break;
			case KW_PROMPTUSER:
				enterOuterAlt(_localctx, 8);
				{
				setState(224);
				promptuser_stmt();
				}
				break;
			case KW_BREAK:
				enterOuterAlt(_localctx, 9);
				{
				setState(225);
				break_statement();
				}
				break;
			case KW_CONTINUE:
				enterOuterAlt(_localctx, 10);
				{
				setState(226);
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
			setState(229);
			match(KW_FUNC);
			setState(230);
			match(IDENTIFIER);
			setState(231);
			signature_part();
			setState(232);
			match(KW_MEANS);
			setState(233);
			match(NEWLINE);
			setState(234);
			metadata_block();
			setState(235);
			non_empty_statement_list();
			setState(236);
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
			setState(256);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case LPAREN:
				enterOuterAlt(_localctx, 1);
				{
				setState(238);
				match(LPAREN);
				setState(244);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while ((((_la) & ~0x3f) == 0 && ((1L << _la) & 9587741394206720L) != 0)) {
					{
					setState(242);
					_errHandler.sync(this);
					switch (_input.LA(1)) {
					case KW_NEEDS:
						{
						setState(239);
						needs_clause();
						}
						break;
					case KW_OPTIONAL:
						{
						setState(240);
						optional_clause();
						}
						break;
					case KW_RETURNS:
						{
						setState(241);
						returns_clause();
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					}
					setState(246);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(247);
				match(RPAREN);
				}
				break;
			case KW_NEEDS:
			case KW_OPTIONAL:
			case KW_RETURNS:
				enterOuterAlt(_localctx, 2);
				{
				setState(251); 
				_errHandler.sync(this);
				_la = _input.LA(1);
				do {
					{
					setState(251);
					_errHandler.sync(this);
					switch (_input.LA(1)) {
					case KW_NEEDS:
						{
						setState(248);
						needs_clause();
						}
						break;
					case KW_OPTIONAL:
						{
						setState(249);
						optional_clause();
						}
						break;
					case KW_RETURNS:
						{
						setState(250);
						returns_clause();
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					}
					setState(253); 
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
			setState(258);
			match(KW_NEEDS);
			setState(259);
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
			setState(261);
			match(KW_OPTIONAL);
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
			setState(264);
			match(KW_RETURNS);
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
			setState(267);
			match(IDENTIFIER);
			setState(272);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(268);
				match(COMMA);
				setState(269);
				match(IDENTIFIER);
				}
				}
				setState(274);
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
			setState(279);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==METADATA_LINE) {
				{
				{
				setState(275);
				match(METADATA_LINE);
				setState(276);
				match(NEWLINE);
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
			setState(285);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==NEWLINE) {
				{
				{
				setState(282);
				match(NEWLINE);
				}
				}
				setState(287);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(288);
			statement();
			setState(289);
			match(NEWLINE);
			setState(293);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while ((((_la) & ~0x3f) == 0 && ((1L << _la) & 4636739500310277952L) != 0) || _la==NEWLINE) {
				{
				{
				setState(290);
				body_line();
				}
				}
				setState(295);
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
			setState(299);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while ((((_la) & ~0x3f) == 0 && ((1L << _la) & 4636739500310277952L) != 0) || _la==NEWLINE) {
				{
				{
				setState(296);
				body_line();
				}
				}
				setState(301);
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
			setState(306);
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
				enterOuterAlt(_localctx, 1);
				{
				setState(302);
				statement();
				setState(303);
				match(NEWLINE);
				}
				break;
			case NEWLINE:
				enterOuterAlt(_localctx, 2);
				{
				setState(305);
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
			setState(311);
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
				enterOuterAlt(_localctx, 1);
				{
				setState(308);
				simple_statement();
				}
				break;
			case KW_FOR:
			case KW_IF:
			case KW_WHILE:
				enterOuterAlt(_localctx, 2);
				{
				setState(309);
				block_statement();
				}
				break;
			case KW_ON:
				enterOuterAlt(_localctx, 3);
				{
				setState(310);
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
			setState(325);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_SET:
				enterOuterAlt(_localctx, 1);
				{
				setState(313);
				set_statement();
				}
				break;
			case KW_CALL:
				enterOuterAlt(_localctx, 2);
				{
				setState(314);
				call_statement();
				}
				break;
			case KW_RETURN:
				enterOuterAlt(_localctx, 3);
				{
				setState(315);
				return_statement();
				}
				break;
			case KW_EMIT:
				enterOuterAlt(_localctx, 4);
				{
				setState(316);
				emit_statement();
				}
				break;
			case KW_MUST:
				enterOuterAlt(_localctx, 5);
				{
				setState(317);
				must_statement();
				}
				break;
			case KW_FAIL:
				enterOuterAlt(_localctx, 6);
				{
				setState(318);
				fail_statement();
				}
				break;
			case KW_CLEAR_ERROR:
				enterOuterAlt(_localctx, 7);
				{
				setState(319);
				clearErrorStmt();
				}
				break;
			case KW_CLEAR:
				enterOuterAlt(_localctx, 8);
				{
				setState(320);
				clearEventStmt();
				}
				break;
			case KW_ASK:
				enterOuterAlt(_localctx, 9);
				{
				setState(321);
				ask_stmt();
				}
				break;
			case KW_PROMPTUSER:
				enterOuterAlt(_localctx, 10);
				{
				setState(322);
				promptuser_stmt();
				}
				break;
			case KW_BREAK:
				enterOuterAlt(_localctx, 11);
				{
				setState(323);
				break_statement();
				}
				break;
			case KW_CONTINUE:
				enterOuterAlt(_localctx, 12);
				{
				setState(324);
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
			setState(330);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_IF:
				enterOuterAlt(_localctx, 1);
				{
				setState(327);
				if_statement();
				}
				break;
			case KW_WHILE:
				enterOuterAlt(_localctx, 2);
				{
				setState(328);
				while_statement();
				}
				break;
			case KW_FOR:
				enterOuterAlt(_localctx, 3);
				{
				setState(329);
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
			setState(332);
			match(KW_ON);
			setState(335);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_ERROR:
				{
				setState(333);
				error_handler();
				}
				break;
			case KW_EVENT:
				{
				setState(334);
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
			setState(337);
			match(KW_ERROR);
			setState(338);
			match(KW_DO);
			setState(339);
			match(NEWLINE);
			setState(340);
			non_empty_statement_list();
			setState(341);
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
			setState(343);
			match(KW_EVENT);
			setState(344);
			expression();
			setState(347);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==KW_NAMED) {
				{
				setState(345);
				match(KW_NAMED);
				setState(346);
				match(STRING_LIT);
				}
			}

			setState(351);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==KW_AS) {
				{
				setState(349);
				match(KW_AS);
				setState(350);
				match(IDENTIFIER);
				}
			}

			setState(353);
			match(KW_DO);
			setState(354);
			match(NEWLINE);
			setState(355);
			non_empty_statement_list();
			setState(356);
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
			setState(358);
			match(KW_CLEAR);
			setState(359);
			match(KW_EVENT);
			setState(363);
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
				setState(360);
				expression();
				}
				break;
			case KW_NAMED:
				{
				setState(361);
				match(KW_NAMED);
				setState(362);
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
			setState(365);
			match(IDENTIFIER);
			setState(374);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==LBRACK || _la==DOT) {
				{
				setState(372);
				_errHandler.sync(this);
				switch (_input.LA(1)) {
				case LBRACK:
					{
					setState(366);
					match(LBRACK);
					setState(367);
					expression();
					setState(368);
					match(RBRACK);
					}
					break;
				case DOT:
					{
					setState(370);
					match(DOT);
					setState(371);
					match(IDENTIFIER);
					}
					break;
				default:
					throw new NoViableAltException(this);
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
			setState(377);
			lvalue();
			setState(382);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(378);
				match(COMMA);
				setState(379);
				lvalue();
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
			setState(385);
			match(KW_SET);
			setState(386);
			lvalue_list();
			setState(387);
			match(ASSIGN);
			setState(388);
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
			setState(390);
			match(KW_CALL);
			setState(391);
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
			setState(393);
			match(KW_RETURN);
			setState(395);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if ((((_la) & ~0x3f) == 0 && ((1L << _la) & 4287674167257481380L) != 0) || ((((_la - 64)) & ~0x3f) == 0 && ((1L << (_la - 64)) & 36274331L) != 0)) {
				{
				setState(394);
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
			setState(397);
			match(KW_EMIT);
			setState(398);
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
		enterRule(_localctx, 68, RULE_must_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(400);
			match(KW_MUST);
			setState(401);
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
		enterRule(_localctx, 70, RULE_fail_statement);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(403);
			match(KW_FAIL);
			setState(405);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if ((((_la) & ~0x3f) == 0 && ((1L << _la) & 4287674167257481380L) != 0) || ((((_la - 64)) & ~0x3f) == 0 && ((1L << (_la - 64)) & 36274331L) != 0)) {
				{
				setState(404);
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
		enterRule(_localctx, 72, RULE_clearErrorStmt);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(407);
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
		enterRule(_localctx, 74, RULE_ask_stmt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(409);
			match(KW_ASK);
			setState(410);
			expression();
			setState(411);
			match(COMMA);
			setState(412);
			expression();
			setState(415);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==KW_WITH) {
				{
				setState(413);
				match(KW_WITH);
				setState(414);
				expression();
				}
			}

			setState(419);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==KW_INTO) {
				{
				setState(417);
				match(KW_INTO);
				setState(418);
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
		enterRule(_localctx, 76, RULE_promptuser_stmt);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(421);
			match(KW_PROMPTUSER);
			setState(422);
			expression();
			setState(423);
			match(KW_INTO);
			setState(424);
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
		enterRule(_localctx, 78, RULE_break_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(426);
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
		enterRule(_localctx, 80, RULE_continue_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(428);
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
		enterRule(_localctx, 82, RULE_if_statement);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(430);
			match(KW_IF);
			setState(431);
			expression();
			setState(432);
			match(NEWLINE);
			setState(433);
			non_empty_statement_list();
			setState(437);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==KW_ELSE) {
				{
				setState(434);
				match(KW_ELSE);
				setState(435);
				match(NEWLINE);
				setState(436);
				non_empty_statement_list();
				}
			}

			setState(439);
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
		enterRule(_localctx, 84, RULE_while_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(441);
			match(KW_WHILE);
			setState(442);
			expression();
			setState(443);
			match(NEWLINE);
			setState(444);
			non_empty_statement_list();
			setState(445);
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
		enterRule(_localctx, 86, RULE_for_each_statement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(447);
			match(KW_FOR);
			setState(448);
			match(KW_EACH);
			setState(449);
			match(IDENTIFIER);
			setState(450);
			match(KW_IN);
			setState(451);
			expression();
			setState(452);
			match(NEWLINE);
			setState(453);
			non_empty_statement_list();
			setState(454);
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
		enterRule(_localctx, 88, RULE_qualified_identifier);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(456);
			match(IDENTIFIER);
			setState(461);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==DOT) {
				{
				{
				setState(457);
				match(DOT);
				setState(458);
				match(IDENTIFIER);
				}
				}
				setState(463);
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
		enterRule(_localctx, 90, RULE_call_target);
		try {
			setState(468);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case IDENTIFIER:
				enterOuterAlt(_localctx, 1);
				{
				setState(464);
				match(IDENTIFIER);
				}
				break;
			case KW_TOOL:
				enterOuterAlt(_localctx, 2);
				{
				setState(465);
				match(KW_TOOL);
				setState(466);
				match(DOT);
				setState(467);
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
		enterRule(_localctx, 92, RULE_expression);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(470);
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
		enterRule(_localctx, 94, RULE_logical_or_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(472);
			logical_and_expr();
			setState(477);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==KW_OR) {
				{
				{
				setState(473);
				match(KW_OR);
				setState(474);
				logical_and_expr();
				}
				}
				setState(479);
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
		enterRule(_localctx, 96, RULE_logical_and_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(480);
			bitwise_or_expr();
			setState(485);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==KW_AND) {
				{
				{
				setState(481);
				match(KW_AND);
				setState(482);
				bitwise_or_expr();
				}
				}
				setState(487);
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
		enterRule(_localctx, 98, RULE_bitwise_or_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(488);
			bitwise_xor_expr();
			setState(493);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==PIPE) {
				{
				{
				setState(489);
				match(PIPE);
				setState(490);
				bitwise_xor_expr();
				}
				}
				setState(495);
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
		enterRule(_localctx, 100, RULE_bitwise_xor_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(496);
			bitwise_and_expr();
			setState(501);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==CARET) {
				{
				{
				setState(497);
				match(CARET);
				setState(498);
				bitwise_and_expr();
				}
				}
				setState(503);
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
		enterRule(_localctx, 102, RULE_bitwise_and_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(504);
			equality_expr();
			setState(509);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==AMPERSAND) {
				{
				{
				setState(505);
				match(AMPERSAND);
				setState(506);
				equality_expr();
				}
				}
				setState(511);
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
		enterRule(_localctx, 104, RULE_equality_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(512);
			relational_expr();
			setState(517);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==EQ || _la==NEQ) {
				{
				{
				setState(513);
				_la = _input.LA(1);
				if ( !(_la==EQ || _la==NEQ) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(514);
				relational_expr();
				}
				}
				setState(519);
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
		enterRule(_localctx, 106, RULE_relational_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(520);
			additive_expr();
			setState(525);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (((((_la - 93)) & ~0x3f) == 0 && ((1L << (_la - 93)) & 15L) != 0)) {
				{
				{
				setState(521);
				_la = _input.LA(1);
				if ( !(((((_la - 93)) & ~0x3f) == 0 && ((1L << (_la - 93)) & 15L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(522);
				additive_expr();
				}
				}
				setState(527);
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
		enterRule(_localctx, 108, RULE_additive_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(528);
			multiplicative_expr();
			setState(533);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==PLUS || _la==MINUS) {
				{
				{
				setState(529);
				_la = _input.LA(1);
				if ( !(_la==PLUS || _la==MINUS) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(530);
				multiplicative_expr();
				}
				}
				setState(535);
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
		enterRule(_localctx, 110, RULE_multiplicative_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(536);
			unary_expr();
			setState(541);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (((((_la - 72)) & ~0x3f) == 0 && ((1L << (_la - 72)) & 7L) != 0)) {
				{
				{
				setState(537);
				_la = _input.LA(1);
				if ( !(((((_la - 72)) & ~0x3f) == 0 && ((1L << (_la - 72)) & 7L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(538);
				unary_expr();
				}
				}
				setState(543);
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
		enterRule(_localctx, 112, RULE_unary_expr);
		int _la;
		try {
			setState(549);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_NO:
			case KW_NOT:
			case KW_SOME:
			case MINUS:
			case TILDE:
				enterOuterAlt(_localctx, 1);
				{
				setState(544);
				_la = _input.LA(1);
				if ( !(((((_la - 46)) & ~0x3f) == 0 && ((1L << (_la - 46)) & 8623490051L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(545);
				unary_expr();
				}
				break;
			case KW_TYPEOF:
				enterOuterAlt(_localctx, 2);
				{
				setState(546);
				match(KW_TYPEOF);
				setState(547);
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
				setState(548);
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
		enterRule(_localctx, 114, RULE_power_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(551);
			accessor_expr();
			setState(554);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==STAR_STAR) {
				{
				setState(552);
				match(STAR_STAR);
				setState(553);
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
		enterRule(_localctx, 116, RULE_accessor_expr);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(556);
			primary();
			setState(563);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==LBRACK) {
				{
				{
				setState(557);
				match(LBRACK);
				setState(558);
				expression();
				setState(559);
				match(RBRACK);
				}
				}
				setState(565);
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
		enterRule(_localctx, 118, RULE_primary);
		try {
			setState(580);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,52,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(566);
				literal();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(567);
				placeholder();
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(568);
				match(IDENTIFIER);
				}
				break;
			case 4:
				enterOuterAlt(_localctx, 4);
				{
				setState(569);
				match(KW_LAST);
				}
				break;
			case 5:
				enterOuterAlt(_localctx, 5);
				{
				setState(570);
				callable_expr();
				}
				break;
			case 6:
				enterOuterAlt(_localctx, 6);
				{
				setState(571);
				match(KW_EVAL);
				setState(572);
				match(LPAREN);
				setState(573);
				expression();
				setState(574);
				match(RPAREN);
				}
				break;
			case 7:
				enterOuterAlt(_localctx, 7);
				{
				setState(576);
				match(LPAREN);
				setState(577);
				expression();
				setState(578);
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
		enterRule(_localctx, 120, RULE_callable_expr);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(592);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case KW_TOOL:
			case IDENTIFIER:
				{
				setState(582);
				call_target();
				}
				break;
			case KW_LN:
				{
				setState(583);
				match(KW_LN);
				}
				break;
			case KW_LOG:
				{
				setState(584);
				match(KW_LOG);
				}
				break;
			case KW_SIN:
				{
				setState(585);
				match(KW_SIN);
				}
				break;
			case KW_COS:
				{
				setState(586);
				match(KW_COS);
				}
				break;
			case KW_TAN:
				{
				setState(587);
				match(KW_TAN);
				}
				break;
			case KW_ASIN:
				{
				setState(588);
				match(KW_ASIN);
				}
				break;
			case KW_ACOS:
				{
				setState(589);
				match(KW_ACOS);
				}
				break;
			case KW_ATAN:
				{
				setState(590);
				match(KW_ATAN);
				}
				break;
			case KW_LEN:
				{
				setState(591);
				match(KW_LEN);
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			setState(594);
			match(LPAREN);
			setState(595);
			expression_list_opt();
			setState(596);
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
		enterRule(_localctx, 122, RULE_placeholder);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(598);
			match(PLACEHOLDER_START);
			setState(599);
			_la = _input.LA(1);
			if ( !(_la==KW_LAST || _la==IDENTIFIER) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			setState(600);
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
		enterRule(_localctx, 124, RULE_literal);
		try {
			setState(609);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case STRING_LIT:
				enterOuterAlt(_localctx, 1);
				{
				setState(602);
				match(STRING_LIT);
				}
				break;
			case TRIPLE_BACKTICK_STRING:
				enterOuterAlt(_localctx, 2);
				{
				setState(603);
				match(TRIPLE_BACKTICK_STRING);
				}
				break;
			case NUMBER_LIT:
				enterOuterAlt(_localctx, 3);
				{
				setState(604);
				match(NUMBER_LIT);
				}
				break;
			case LBRACK:
				enterOuterAlt(_localctx, 4);
				{
				setState(605);
				list_literal();
				}
				break;
			case LBRACE:
				enterOuterAlt(_localctx, 5);
				{
				setState(606);
				map_literal();
				}
				break;
			case KW_FALSE:
			case KW_TRUE:
				enterOuterAlt(_localctx, 6);
				{
				setState(607);
				boolean_literal();
				}
				break;
			case KW_NIL:
				enterOuterAlt(_localctx, 7);
				{
				setState(608);
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
		enterRule(_localctx, 126, RULE_nil_literal);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(611);
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
		enterRule(_localctx, 128, RULE_boolean_literal);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(613);
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
		enterRule(_localctx, 130, RULE_list_literal);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(615);
			match(LBRACK);
			setState(616);
			expression_list_opt();
			setState(617);
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
		enterRule(_localctx, 132, RULE_map_literal);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(619);
			match(LBRACE);
			setState(620);
			map_entry_list_opt();
			setState(621);
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
		enterRule(_localctx, 134, RULE_expression_list_opt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(624);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if ((((_la) & ~0x3f) == 0 && ((1L << _la) & 4287674167257481380L) != 0) || ((((_la - 64)) & ~0x3f) == 0 && ((1L << (_la - 64)) & 36274331L) != 0)) {
				{
				setState(623);
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
		enterRule(_localctx, 136, RULE_expression_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(626);
			expression();
			setState(631);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(627);
				match(COMMA);
				setState(628);
				expression();
				}
				}
				setState(633);
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
		enterRule(_localctx, 138, RULE_map_entry_list_opt);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(635);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==STRING_LIT) {
				{
				setState(634);
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
		enterRule(_localctx, 140, RULE_map_entry_list);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(637);
			map_entry();
			setState(642);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(638);
				match(COMMA);
				setState(639);
				map_entry();
				}
				}
				setState(644);
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
		enterRule(_localctx, 142, RULE_map_entry);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(645);
			match(STRING_LIT);
			setState(646);
			match(COLON);
			setState(647);
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
		"\u0004\u0001c\u028a\u0002\u0000\u0007\u0000\u0002\u0001\u0007\u0001\u0002"+
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
		"F\u0007F\u0002G\u0007G\u0001\u0000\u0001\u0000\u0001\u0000\u0003\u0000"+
		"\u0094\b\u0000\u0001\u0000\u0001\u0000\u0001\u0001\u0005\u0001\u0099\b"+
		"\u0001\n\u0001\f\u0001\u009c\t\u0001\u0001\u0002\u0004\u0002\u009f\b\u0002"+
		"\u000b\u0002\f\u0002\u00a0\u0001\u0003\u0004\u0003\u00a4\b\u0003\u000b"+
		"\u0003\f\u0003\u00a5\u0001\u0004\u0001\u0004\u0001\u0004\u0003\u0004\u00ab"+
		"\b\u0004\u0001\u0004\u0005\u0004\u00ae\b\u0004\n\u0004\f\u0004\u00b1\t"+
		"\u0004\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001"+
		"\u0005\u0005\u0005\u00b9\b\u0005\n\u0005\f\u0005\u00bc\t\u0005\u0001\u0006"+
		"\u0005\u0006\u00bf\b\u0006\n\u0006\f\u0006\u00c2\t\u0006\u0001\u0006\u0001"+
		"\u0006\u0001\u0006\u0005\u0006\u00c7\b\u0006\n\u0006\f\u0006\u00ca\t\u0006"+
		"\u0001\u0007\u0001\u0007\u0001\u0007\u0001\u0007\u0003\u0007\u00d0\b\u0007"+
		"\u0001\b\u0001\b\u0001\b\u0003\b\u00d5\b\b\u0001\t\u0001\t\u0001\t\u0001"+
		"\n\u0001\n\u0001\n\u0001\n\u0001\n\u0001\n\u0001\n\u0001\n\u0001\n\u0001"+
		"\n\u0003\n\u00e4\b\n\u0001\u000b\u0001\u000b\u0001\u000b\u0001\u000b\u0001"+
		"\u000b\u0001\u000b\u0001\u000b\u0001\u000b\u0001\u000b\u0001\f\u0001\f"+
		"\u0001\f\u0001\f\u0005\f\u00f3\b\f\n\f\f\f\u00f6\t\f\u0001\f\u0001\f\u0001"+
		"\f\u0001\f\u0004\f\u00fc\b\f\u000b\f\f\f\u00fd\u0001\f\u0003\f\u0101\b"+
		"\f\u0001\r\u0001\r\u0001\r\u0001\u000e\u0001\u000e\u0001\u000e\u0001\u000f"+
		"\u0001\u000f\u0001\u000f\u0001\u0010\u0001\u0010\u0001\u0010\u0005\u0010"+
		"\u010f\b\u0010\n\u0010\f\u0010\u0112\t\u0010\u0001\u0011\u0001\u0011\u0005"+
		"\u0011\u0116\b\u0011\n\u0011\f\u0011\u0119\t\u0011\u0001\u0012\u0005\u0012"+
		"\u011c\b\u0012\n\u0012\f\u0012\u011f\t\u0012\u0001\u0012\u0001\u0012\u0001"+
		"\u0012\u0005\u0012\u0124\b\u0012\n\u0012\f\u0012\u0127\t\u0012\u0001\u0013"+
		"\u0005\u0013\u012a\b\u0013\n\u0013\f\u0013\u012d\t\u0013\u0001\u0014\u0001"+
		"\u0014\u0001\u0014\u0001\u0014\u0003\u0014\u0133\b\u0014\u0001\u0015\u0001"+
		"\u0015\u0001\u0015\u0003\u0015\u0138\b\u0015\u0001\u0016\u0001\u0016\u0001"+
		"\u0016\u0001\u0016\u0001\u0016\u0001\u0016\u0001\u0016\u0001\u0016\u0001"+
		"\u0016\u0001\u0016\u0001\u0016\u0001\u0016\u0003\u0016\u0146\b\u0016\u0001"+
		"\u0017\u0001\u0017\u0001\u0017\u0003\u0017\u014b\b\u0017\u0001\u0018\u0001"+
		"\u0018\u0001\u0018\u0003\u0018\u0150\b\u0018\u0001\u0019\u0001\u0019\u0001"+
		"\u0019\u0001\u0019\u0001\u0019\u0001\u0019\u0001\u001a\u0001\u001a\u0001"+
		"\u001a\u0001\u001a\u0003\u001a\u015c\b\u001a\u0001\u001a\u0001\u001a\u0003"+
		"\u001a\u0160\b\u001a\u0001\u001a\u0001\u001a\u0001\u001a\u0001\u001a\u0001"+
		"\u001a\u0001\u001b\u0001\u001b\u0001\u001b\u0001\u001b\u0001\u001b\u0003"+
		"\u001b\u016c\b\u001b\u0001\u001c\u0001\u001c\u0001\u001c\u0001\u001c\u0001"+
		"\u001c\u0001\u001c\u0001\u001c\u0005\u001c\u0175\b\u001c\n\u001c\f\u001c"+
		"\u0178\t\u001c\u0001\u001d\u0001\u001d\u0001\u001d\u0005\u001d\u017d\b"+
		"\u001d\n\u001d\f\u001d\u0180\t\u001d\u0001\u001e\u0001\u001e\u0001\u001e"+
		"\u0001\u001e\u0001\u001e\u0001\u001f\u0001\u001f\u0001\u001f\u0001 \u0001"+
		" \u0003 \u018c\b \u0001!\u0001!\u0001!\u0001\"\u0001\"\u0001\"\u0001#"+
		"\u0001#\u0003#\u0196\b#\u0001$\u0001$\u0001%\u0001%\u0001%\u0001%\u0001"+
		"%\u0001%\u0003%\u01a0\b%\u0001%\u0001%\u0003%\u01a4\b%\u0001&\u0001&\u0001"+
		"&\u0001&\u0001&\u0001\'\u0001\'\u0001(\u0001(\u0001)\u0001)\u0001)\u0001"+
		")\u0001)\u0001)\u0001)\u0003)\u01b6\b)\u0001)\u0001)\u0001*\u0001*\u0001"+
		"*\u0001*\u0001*\u0001*\u0001+\u0001+\u0001+\u0001+\u0001+\u0001+\u0001"+
		"+\u0001+\u0001+\u0001,\u0001,\u0001,\u0005,\u01cc\b,\n,\f,\u01cf\t,\u0001"+
		"-\u0001-\u0001-\u0001-\u0003-\u01d5\b-\u0001.\u0001.\u0001/\u0001/\u0001"+
		"/\u0005/\u01dc\b/\n/\f/\u01df\t/\u00010\u00010\u00010\u00050\u01e4\b0"+
		"\n0\f0\u01e7\t0\u00011\u00011\u00011\u00051\u01ec\b1\n1\f1\u01ef\t1\u0001"+
		"2\u00012\u00012\u00052\u01f4\b2\n2\f2\u01f7\t2\u00013\u00013\u00013\u0005"+
		"3\u01fc\b3\n3\f3\u01ff\t3\u00014\u00014\u00014\u00054\u0204\b4\n4\f4\u0207"+
		"\t4\u00015\u00015\u00015\u00055\u020c\b5\n5\f5\u020f\t5\u00016\u00016"+
		"\u00016\u00056\u0214\b6\n6\f6\u0217\t6\u00017\u00017\u00017\u00057\u021c"+
		"\b7\n7\f7\u021f\t7\u00018\u00018\u00018\u00018\u00018\u00038\u0226\b8"+
		"\u00019\u00019\u00019\u00039\u022b\b9\u0001:\u0001:\u0001:\u0001:\u0001"+
		":\u0005:\u0232\b:\n:\f:\u0235\t:\u0001;\u0001;\u0001;\u0001;\u0001;\u0001"+
		";\u0001;\u0001;\u0001;\u0001;\u0001;\u0001;\u0001;\u0001;\u0003;\u0245"+
		"\b;\u0001<\u0001<\u0001<\u0001<\u0001<\u0001<\u0001<\u0001<\u0001<\u0001"+
		"<\u0003<\u0251\b<\u0001<\u0001<\u0001<\u0001<\u0001=\u0001=\u0001=\u0001"+
		"=\u0001>\u0001>\u0001>\u0001>\u0001>\u0001>\u0001>\u0003>\u0262\b>\u0001"+
		"?\u0001?\u0001@\u0001@\u0001A\u0001A\u0001A\u0001A\u0001B\u0001B\u0001"+
		"B\u0001B\u0001C\u0003C\u0271\bC\u0001D\u0001D\u0001D\u0005D\u0276\bD\n"+
		"D\fD\u0279\tD\u0001E\u0003E\u027c\bE\u0001F\u0001F\u0001F\u0005F\u0281"+
		"\bF\nF\fF\u0284\tF\u0001G\u0001G\u0001G\u0001G\u0001G\u0000\u0000H\u0000"+
		"\u0002\u0004\u0006\b\n\f\u000e\u0010\u0012\u0014\u0016\u0018\u001a\u001c"+
		"\u001e \"$&(*,.02468:<>@BDFHJLNPRTVXZ\\^`bdfhjlnprtvxz|~\u0080\u0082\u0084"+
		"\u0086\u0088\u008a\u008c\u008e\u0000\b\u0002\u0000BBbb\u0001\u0000[\\"+
		"\u0001\u0000]`\u0001\u0000FG\u0001\u0000HJ\u0004\u0000./88GGOO\u0002\u0000"+
		"$$DD\u0002\u0000\u001d\u001d<<\u02a8\u0000\u0090\u0001\u0000\u0000\u0000"+
		"\u0002\u009a\u0001\u0000\u0000\u0000\u0004\u009e\u0001\u0000\u0000\u0000"+
		"\u0006\u00a3\u0001\u0000\u0000\u0000\b\u00aa\u0001\u0000\u0000\u0000\n"+
		"\u00b2\u0001\u0000\u0000\u0000\f\u00c0\u0001\u0000\u0000\u0000\u000e\u00cf"+
		"\u0001\u0000\u0000\u0000\u0010\u00d4\u0001\u0000\u0000\u0000\u0012\u00d6"+
		"\u0001\u0000\u0000\u0000\u0014\u00e3\u0001\u0000\u0000\u0000\u0016\u00e5"+
		"\u0001\u0000\u0000\u0000\u0018\u0100\u0001\u0000\u0000\u0000\u001a\u0102"+
		"\u0001\u0000\u0000\u0000\u001c\u0105\u0001\u0000\u0000\u0000\u001e\u0108"+
		"\u0001\u0000\u0000\u0000 \u010b\u0001\u0000\u0000\u0000\"\u0117\u0001"+
		"\u0000\u0000\u0000$\u011d\u0001\u0000\u0000\u0000&\u012b\u0001\u0000\u0000"+
		"\u0000(\u0132\u0001\u0000\u0000\u0000*\u0137\u0001\u0000\u0000\u0000,"+
		"\u0145\u0001\u0000\u0000\u0000.\u014a\u0001\u0000\u0000\u00000\u014c\u0001"+
		"\u0000\u0000\u00002\u0151\u0001\u0000\u0000\u00004\u0157\u0001\u0000\u0000"+
		"\u00006\u0166\u0001\u0000\u0000\u00008\u016d\u0001\u0000\u0000\u0000:"+
		"\u0179\u0001\u0000\u0000\u0000<\u0181\u0001\u0000\u0000\u0000>\u0186\u0001"+
		"\u0000\u0000\u0000@\u0189\u0001\u0000\u0000\u0000B\u018d\u0001\u0000\u0000"+
		"\u0000D\u0190\u0001\u0000\u0000\u0000F\u0193\u0001\u0000\u0000\u0000H"+
		"\u0197\u0001\u0000\u0000\u0000J\u0199\u0001\u0000\u0000\u0000L\u01a5\u0001"+
		"\u0000\u0000\u0000N\u01aa\u0001\u0000\u0000\u0000P\u01ac\u0001\u0000\u0000"+
		"\u0000R\u01ae\u0001\u0000\u0000\u0000T\u01b9\u0001\u0000\u0000\u0000V"+
		"\u01bf\u0001\u0000\u0000\u0000X\u01c8\u0001\u0000\u0000\u0000Z\u01d4\u0001"+
		"\u0000\u0000\u0000\\\u01d6\u0001\u0000\u0000\u0000^\u01d8\u0001\u0000"+
		"\u0000\u0000`\u01e0\u0001\u0000\u0000\u0000b\u01e8\u0001\u0000\u0000\u0000"+
		"d\u01f0\u0001\u0000\u0000\u0000f\u01f8\u0001\u0000\u0000\u0000h\u0200"+
		"\u0001\u0000\u0000\u0000j\u0208\u0001\u0000\u0000\u0000l\u0210\u0001\u0000"+
		"\u0000\u0000n\u0218\u0001\u0000\u0000\u0000p\u0225\u0001\u0000\u0000\u0000"+
		"r\u0227\u0001\u0000\u0000\u0000t\u022c\u0001\u0000\u0000\u0000v\u0244"+
		"\u0001\u0000\u0000\u0000x\u0250\u0001\u0000\u0000\u0000z\u0256\u0001\u0000"+
		"\u0000\u0000|\u0261\u0001\u0000\u0000\u0000~\u0263\u0001\u0000\u0000\u0000"+
		"\u0080\u0265\u0001\u0000\u0000\u0000\u0082\u0267\u0001\u0000\u0000\u0000"+
		"\u0084\u026b\u0001\u0000\u0000\u0000\u0086\u0270\u0001\u0000\u0000\u0000"+
		"\u0088\u0272\u0001\u0000\u0000\u0000\u008a\u027b\u0001\u0000\u0000\u0000"+
		"\u008c\u027d\u0001\u0000\u0000\u0000\u008e\u0285\u0001\u0000\u0000\u0000"+
		"\u0090\u0093\u0003\u0002\u0001\u0000\u0091\u0094\u0003\u0004\u0002\u0000"+
		"\u0092\u0094\u0003\u0006\u0003\u0000\u0093\u0091\u0001\u0000\u0000\u0000"+
		"\u0093\u0092\u0001\u0000\u0000\u0000\u0093\u0094\u0001\u0000\u0000\u0000"+
		"\u0094\u0095\u0001\u0000\u0000\u0000\u0095\u0096\u0005\u0000\u0000\u0001"+
		"\u0096\u0001\u0001\u0000\u0000\u0000\u0097\u0099\u0007\u0000\u0000\u0000"+
		"\u0098\u0097\u0001\u0000\u0000\u0000\u0099\u009c\u0001\u0000\u0000\u0000"+
		"\u009a\u0098\u0001\u0000\u0000\u0000\u009a\u009b\u0001\u0000\u0000\u0000"+
		"\u009b\u0003\u0001\u0000\u0000\u0000\u009c\u009a\u0001\u0000\u0000\u0000"+
		"\u009d\u009f\u0003\b\u0004\u0000\u009e\u009d\u0001\u0000\u0000\u0000\u009f"+
		"\u00a0\u0001\u0000\u0000\u0000\u00a0\u009e\u0001\u0000\u0000\u0000\u00a0"+
		"\u00a1\u0001\u0000\u0000\u0000\u00a1\u0005\u0001\u0000\u0000\u0000\u00a2"+
		"\u00a4\u0003\n\u0005\u0000\u00a3\u00a2\u0001\u0000\u0000\u0000\u00a4\u00a5"+
		"\u0001\u0000\u0000\u0000\u00a5\u00a3\u0001\u0000\u0000\u0000\u00a5\u00a6"+
		"\u0001\u0000\u0000\u0000\u00a6\u0007\u0001\u0000\u0000\u0000\u00a7\u00ab"+
		"\u0003\u0016\u000b\u0000\u00a8\u00a9\u00050\u0000\u0000\u00a9\u00ab\u0003"+
		"4\u001a\u0000\u00aa\u00a7\u0001\u0000\u0000\u0000\u00aa\u00a8\u0001\u0000"+
		"\u0000\u0000\u00ab\u00af\u0001\u0000\u0000\u0000\u00ac\u00ae\u0005b\u0000"+
		"\u0000\u00ad\u00ac\u0001\u0000\u0000\u0000\u00ae\u00b1\u0001\u0000\u0000"+
		"\u0000\u00af\u00ad\u0001\u0000\u0000\u0000\u00af\u00b0\u0001\u0000\u0000"+
		"\u0000\u00b0\t\u0001\u0000\u0000\u0000\u00b1\u00af\u0001\u0000\u0000\u0000"+
		"\u00b2\u00b3\u0005\f\u0000\u0000\u00b3\u00b4\u0005b\u0000\u0000\u00b4"+
		"\u00b5\u0003\"\u0011\u0000\u00b5\u00b6\u0003\f\u0006\u0000\u00b6\u00ba"+
		"\u0005\u0013\u0000\u0000\u00b7\u00b9\u0005b\u0000\u0000\u00b8\u00b7\u0001"+
		"\u0000\u0000\u0000\u00b9\u00bc\u0001\u0000\u0000\u0000\u00ba\u00b8\u0001"+
		"\u0000\u0000\u0000\u00ba\u00bb\u0001\u0000\u0000\u0000\u00bb\u000b\u0001"+
		"\u0000\u0000\u0000\u00bc\u00ba\u0001\u0000\u0000\u0000\u00bd\u00bf\u0005"+
		"b\u0000\u0000\u00be\u00bd\u0001\u0000\u0000\u0000\u00bf\u00c2\u0001\u0000"+
		"\u0000\u0000\u00c0\u00be\u0001\u0000\u0000\u0000\u00c0\u00c1\u0001\u0000"+
		"\u0000\u0000\u00c1\u00c3\u0001\u0000\u0000\u0000\u00c2\u00c0\u0001\u0000"+
		"\u0000\u0000\u00c3\u00c4\u0003\u0010\b\u0000\u00c4\u00c8\u0005b\u0000"+
		"\u0000\u00c5\u00c7\u0003\u000e\u0007\u0000\u00c6\u00c5\u0001\u0000\u0000"+
		"\u0000\u00c7\u00ca\u0001\u0000\u0000\u0000\u00c8\u00c6\u0001\u0000\u0000"+
		"\u0000\u00c8\u00c9\u0001\u0000\u0000\u0000\u00c9\r\u0001\u0000\u0000\u0000"+
		"\u00ca\u00c8\u0001\u0000\u0000\u0000\u00cb\u00cc\u0003\u0010\b\u0000\u00cc"+
		"\u00cd\u0005b\u0000\u0000\u00cd\u00d0\u0001\u0000\u0000\u0000\u00ce\u00d0"+
		"\u0005b\u0000\u0000\u00cf\u00cb\u0001\u0000\u0000\u0000\u00cf\u00ce\u0001"+
		"\u0000\u0000\u0000\u00d0\u000f\u0001\u0000\u0000\u0000\u00d1\u00d5\u0003"+
		"\u0014\n\u0000\u00d2\u00d5\u0003.\u0017\u0000\u00d3\u00d5\u0003\u0012"+
		"\t\u0000\u00d4\u00d1\u0001\u0000\u0000\u0000\u00d4\u00d2\u0001\u0000\u0000"+
		"\u0000\u00d4\u00d3\u0001\u0000\u0000\u0000\u00d5\u0011\u0001\u0000\u0000"+
		"\u0000\u00d6\u00d7\u00050\u0000\u0000\u00d7\u00d8\u00032\u0019\u0000\u00d8"+
		"\u0013\u0001\u0000\u0000\u0000\u00d9\u00e4\u0003<\u001e\u0000\u00da\u00e4"+
		"\u0003>\u001f\u0000\u00db\u00e4\u0003B!\u0000\u00dc\u00e4\u0003D\"\u0000"+
		"\u00dd\u00e4\u0003F#\u0000\u00de\u00e4\u00036\u001b\u0000\u00df\u00e4"+
		"\u0003J%\u0000\u00e0\u00e4\u0003L&\u0000\u00e1\u00e4\u0003N\'\u0000\u00e2"+
		"\u00e4\u0003P(\u0000\u00e3\u00d9\u0001\u0000\u0000\u0000\u00e3\u00da\u0001"+
		"\u0000\u0000\u0000\u00e3\u00db\u0001\u0000\u0000\u0000\u00e3\u00dc\u0001"+
		"\u0000\u0000\u0000\u00e3\u00dd\u0001\u0000\u0000\u0000\u00e3\u00de\u0001"+
		"\u0000\u0000\u0000\u00e3\u00df\u0001\u0000\u0000\u0000\u00e3\u00e0\u0001"+
		"\u0000\u0000\u0000\u00e3\u00e1\u0001\u0000\u0000\u0000\u00e3\u00e2\u0001"+
		"\u0000\u0000\u0000\u00e4\u0015\u0001\u0000\u0000\u0000\u00e5\u00e6\u0005"+
		"\u001f\u0000\u0000\u00e6\u00e7\u0005D\u0000\u0000\u00e7\u00e8\u0003\u0018"+
		"\f\u0000\u00e8\u00e9\u0005(\u0000\u0000\u00e9\u00ea\u0005b\u0000\u0000"+
		"\u00ea\u00eb\u0003\"\u0011\u0000\u00eb\u00ec\u0003$\u0012\u0000\u00ec"+
		"\u00ed\u0005\u0015\u0000\u0000\u00ed\u0017\u0001\u0000\u0000\u0000\u00ee"+
		"\u00f4\u0005P\u0000\u0000\u00ef\u00f3\u0003\u001a\r\u0000\u00f0\u00f3"+
		"\u0003\u001c\u000e\u0000\u00f1\u00f3\u0003\u001e\u000f\u0000\u00f2\u00ef"+
		"\u0001\u0000\u0000\u0000\u00f2\u00f0\u0001\u0000\u0000\u0000\u00f2\u00f1"+
		"\u0001\u0000\u0000\u0000\u00f3\u00f6\u0001\u0000\u0000\u0000\u00f4\u00f2"+
		"\u0001\u0000\u0000\u0000\u00f4\u00f5\u0001\u0000\u0000\u0000\u00f5\u00f7"+
		"\u0001\u0000\u0000\u0000\u00f6\u00f4\u0001\u0000\u0000\u0000\u00f7\u0101"+
		"\u0005Q\u0000\u0000\u00f8\u00fc\u0003\u001a\r\u0000\u00f9\u00fc\u0003"+
		"\u001c\u000e\u0000\u00fa\u00fc\u0003\u001e\u000f\u0000\u00fb\u00f8\u0001"+
		"\u0000\u0000\u0000\u00fb\u00f9\u0001\u0000\u0000\u0000\u00fb\u00fa\u0001"+
		"\u0000\u0000\u0000\u00fc\u00fd\u0001\u0000\u0000\u0000\u00fd\u00fb\u0001"+
		"\u0000\u0000\u0000\u00fd\u00fe\u0001\u0000\u0000\u0000\u00fe\u0101\u0001"+
		"\u0000\u0000\u0000\u00ff\u0101\u0001\u0000\u0000\u0000\u0100\u00ee\u0001"+
		"\u0000\u0000\u0000\u0100\u00fb\u0001\u0000\u0000\u0000\u0100\u00ff\u0001"+
		"\u0000\u0000\u0000\u0101\u0019\u0001\u0000\u0000\u0000\u0102\u0103\u0005"+
		",\u0000\u0000\u0103\u0104\u0003 \u0010\u0000\u0104\u001b\u0001\u0000\u0000"+
		"\u0000\u0105\u0106\u00051\u0000\u0000\u0106\u0107\u0003 \u0010\u0000\u0107"+
		"\u001d\u0001\u0000\u0000\u0000\u0108\u0109\u00055\u0000\u0000\u0109\u010a"+
		"\u0003 \u0010\u0000\u010a\u001f\u0001\u0000\u0000\u0000\u010b\u0110\u0005"+
		"D\u0000\u0000\u010c\u010d\u0005R\u0000\u0000\u010d\u010f\u0005D\u0000"+
		"\u0000\u010e\u010c\u0001\u0000\u0000\u0000\u010f\u0112\u0001\u0000\u0000"+
		"\u0000\u0110\u010e\u0001\u0000\u0000\u0000\u0110\u0111\u0001\u0000\u0000"+
		"\u0000\u0111!\u0001\u0000\u0000\u0000\u0112\u0110\u0001\u0000\u0000\u0000"+
		"\u0113\u0114\u0005B\u0000\u0000\u0114\u0116\u0005b\u0000\u0000\u0115\u0113"+
		"\u0001\u0000\u0000\u0000\u0116\u0119\u0001\u0000\u0000\u0000\u0117\u0115"+
		"\u0001\u0000\u0000\u0000\u0117\u0118\u0001\u0000\u0000\u0000\u0118#\u0001"+
		"\u0000\u0000\u0000\u0119\u0117\u0001\u0000\u0000\u0000\u011a\u011c\u0005"+
		"b\u0000\u0000\u011b\u011a\u0001\u0000\u0000\u0000\u011c\u011f\u0001\u0000"+
		"\u0000\u0000\u011d\u011b\u0001\u0000\u0000\u0000\u011d\u011e\u0001\u0000"+
		"\u0000\u0000\u011e\u0120\u0001\u0000\u0000\u0000\u011f\u011d\u0001\u0000"+
		"\u0000\u0000\u0120\u0121\u0003*\u0015\u0000\u0121\u0125\u0005b\u0000\u0000"+
		"\u0122\u0124\u0003(\u0014\u0000\u0123\u0122\u0001\u0000\u0000\u0000\u0124"+
		"\u0127\u0001\u0000\u0000\u0000\u0125\u0123\u0001\u0000\u0000\u0000\u0125"+
		"\u0126\u0001\u0000\u0000\u0000\u0126%\u0001\u0000\u0000\u0000\u0127\u0125"+
		"\u0001\u0000\u0000\u0000\u0128\u012a\u0003(\u0014\u0000\u0129\u0128\u0001"+
		"\u0000\u0000\u0000\u012a\u012d\u0001\u0000\u0000\u0000\u012b\u0129\u0001"+
		"\u0000\u0000\u0000\u012b\u012c\u0001\u0000\u0000\u0000\u012c\'\u0001\u0000"+
		"\u0000\u0000\u012d\u012b\u0001\u0000\u0000\u0000\u012e\u012f\u0003*\u0015"+
		"\u0000\u012f\u0130\u0005b\u0000\u0000\u0130\u0133\u0001\u0000\u0000\u0000"+
		"\u0131\u0133\u0005b\u0000\u0000\u0132\u012e\u0001\u0000\u0000\u0000\u0132"+
		"\u0131\u0001\u0000\u0000\u0000\u0133)\u0001\u0000\u0000\u0000\u0134\u0138"+
		"\u0003,\u0016\u0000\u0135\u0138\u0003.\u0017\u0000\u0136\u0138\u00030"+
		"\u0018\u0000\u0137\u0134\u0001\u0000\u0000\u0000\u0137\u0135\u0001\u0000"+
		"\u0000\u0000\u0137\u0136\u0001\u0000\u0000\u0000\u0138+\u0001\u0000\u0000"+
		"\u0000\u0139\u0146\u0003<\u001e\u0000\u013a\u0146\u0003>\u001f\u0000\u013b"+
		"\u0146\u0003@ \u0000\u013c\u0146\u0003B!\u0000\u013d\u0146\u0003D\"\u0000"+
		"\u013e\u0146\u0003F#\u0000\u013f\u0146\u0003H$\u0000\u0140\u0146\u0003"+
		"6\u001b\u0000\u0141\u0146\u0003J%\u0000\u0142\u0146\u0003L&\u0000\u0143"+
		"\u0146\u0003N\'\u0000\u0144\u0146\u0003P(\u0000\u0145\u0139\u0001\u0000"+
		"\u0000\u0000\u0145\u013a\u0001\u0000\u0000\u0000\u0145\u013b\u0001\u0000"+
		"\u0000\u0000\u0145\u013c\u0001\u0000\u0000\u0000\u0145\u013d\u0001\u0000"+
		"\u0000\u0000\u0145\u013e\u0001\u0000\u0000\u0000\u0145\u013f\u0001\u0000"+
		"\u0000\u0000\u0145\u0140\u0001\u0000\u0000\u0000\u0145\u0141\u0001\u0000"+
		"\u0000\u0000\u0145\u0142\u0001\u0000\u0000\u0000\u0145\u0143\u0001\u0000"+
		"\u0000\u0000\u0145\u0144\u0001\u0000\u0000\u0000\u0146-\u0001\u0000\u0000"+
		"\u0000\u0147\u014b\u0003R)\u0000\u0148\u014b\u0003T*\u0000\u0149\u014b"+
		"\u0003V+\u0000\u014a\u0147\u0001\u0000\u0000\u0000\u014a\u0148\u0001\u0000"+
		"\u0000\u0000\u014a\u0149\u0001\u0000\u0000\u0000\u014b/\u0001\u0000\u0000"+
		"\u0000\u014c\u014f\u00050\u0000\u0000\u014d\u0150\u00032\u0019\u0000\u014e"+
		"\u0150\u00034\u001a\u0000\u014f\u014d\u0001\u0000\u0000\u0000\u014f\u014e"+
		"\u0001\u0000\u0000\u0000\u01501\u0001\u0000\u0000\u0000\u0151\u0152\u0005"+
		"\u0019\u0000\u0000\u0152\u0153\u0005\u000f\u0000\u0000\u0153\u0154\u0005"+
		"b\u0000\u0000\u0154\u0155\u0003$\u0012\u0000\u0155\u0156\u0005\u0017\u0000"+
		"\u0000\u01563\u0001\u0000\u0000\u0000\u0157\u0158\u0005\u001b\u0000\u0000"+
		"\u0158\u015b\u0003\\.\u0000\u0159\u015a\u0005+\u0000\u0000\u015a\u015c"+
		"\u0005@\u0000\u0000\u015b\u0159\u0001\u0000\u0000\u0000\u015b\u015c\u0001"+
		"\u0000\u0000\u0000\u015c\u015f\u0001\u0000\u0000\u0000\u015d\u015e\u0005"+
		"\u0004\u0000\u0000\u015e\u0160\u0005D\u0000\u0000\u015f\u015d\u0001\u0000"+
		"\u0000\u0000\u015f\u0160\u0001\u0000\u0000\u0000\u0160\u0161\u0001\u0000"+
		"\u0000\u0000\u0161\u0162\u0005\u000f\u0000\u0000\u0162\u0163\u0005b\u0000"+
		"\u0000\u0163\u0164\u0003$\u0012\u0000\u0164\u0165\u0005\u0017\u0000\u0000"+
		"\u01655\u0001\u0000\u0000\u0000\u0166\u0167\u0005\n\u0000\u0000\u0167"+
		"\u016b\u0005\u001b\u0000\u0000\u0168\u016c\u0003\\.\u0000\u0169\u016a"+
		"\u0005+\u0000\u0000\u016a\u016c\u0005@\u0000\u0000\u016b\u0168\u0001\u0000"+
		"\u0000\u0000\u016b\u0169\u0001\u0000\u0000\u0000\u016c7\u0001\u0000\u0000"+
		"\u0000\u016d\u0176\u0005D\u0000\u0000\u016e\u016f\u0005S\u0000\u0000\u016f"+
		"\u0170\u0003\\.\u0000\u0170\u0171\u0005T\u0000\u0000\u0171\u0175\u0001"+
		"\u0000\u0000\u0000\u0172\u0173\u0005X\u0000\u0000\u0173\u0175\u0005D\u0000"+
		"\u0000\u0174\u016e\u0001\u0000\u0000\u0000\u0174\u0172\u0001\u0000\u0000"+
		"\u0000\u0175\u0178\u0001\u0000\u0000\u0000\u0176\u0174\u0001\u0000\u0000"+
		"\u0000\u0176\u0177\u0001\u0000\u0000\u0000\u01779\u0001\u0000\u0000\u0000"+
		"\u0178\u0176\u0001\u0000\u0000\u0000\u0179\u017e\u00038\u001c\u0000\u017a"+
		"\u017b\u0005R\u0000\u0000\u017b\u017d\u00038\u001c\u0000\u017c\u017a\u0001"+
		"\u0000\u0000\u0000\u017d\u0180\u0001\u0000\u0000\u0000\u017e\u017c\u0001"+
		"\u0000\u0000\u0000\u017e\u017f\u0001\u0000\u0000\u0000\u017f;\u0001\u0000"+
		"\u0000\u0000\u0180\u017e\u0001\u0000\u0000\u0000\u0181\u0182\u00056\u0000"+
		"\u0000\u0182\u0183\u0003:\u001d\u0000\u0183\u0184\u0005E\u0000\u0000\u0184"+
		"\u0185\u0003\\.\u0000\u0185=\u0001\u0000\u0000\u0000\u0186\u0187\u0005"+
		"\t\u0000\u0000\u0187\u0188\u0003x<\u0000\u0188?\u0001\u0000\u0000\u0000"+
		"\u0189\u018b\u00054\u0000\u0000\u018a\u018c\u0003\u0088D\u0000\u018b\u018a"+
		"\u0001\u0000\u0000\u0000\u018b\u018c\u0001\u0000\u0000\u0000\u018cA\u0001"+
		"\u0000\u0000\u0000\u018d\u018e\u0005\u0012\u0000\u0000\u018e\u018f\u0003"+
		"\\.\u0000\u018fC\u0001\u0000\u0000\u0000\u0190\u0191\u0005)\u0000\u0000"+
		"\u0191\u0192\u0003\\.\u0000\u0192E\u0001\u0000\u0000\u0000\u0193\u0195"+
		"\u0005\u001c\u0000\u0000\u0194\u0196\u0003\\.\u0000\u0195\u0194\u0001"+
		"\u0000\u0000\u0000\u0195\u0196\u0001\u0000\u0000\u0000\u0196G\u0001\u0000"+
		"\u0000\u0000\u0197\u0198\u0005\u000b\u0000\u0000\u0198I\u0001\u0000\u0000"+
		"\u0000\u0199\u019a\u0005\u0006\u0000\u0000\u019a\u019b\u0003\\.\u0000"+
		"\u019b\u019c\u0005R\u0000\u0000\u019c\u019f\u0003\\.\u0000\u019d\u019e"+
		"\u0005?\u0000\u0000\u019e\u01a0\u0003\\.\u0000\u019f\u019d\u0001\u0000"+
		"\u0000\u0000\u019f\u01a0\u0001\u0000\u0000\u0000\u01a0\u01a3\u0001\u0000"+
		"\u0000\u0000\u01a1\u01a2\u0005#\u0000\u0000\u01a2\u01a4\u00038\u001c\u0000"+
		"\u01a3\u01a1\u0001\u0000\u0000\u0000\u01a3\u01a4\u0001\u0000\u0000\u0000"+
		"\u01a4K\u0001\u0000\u0000\u0000\u01a5\u01a6\u00053\u0000\u0000\u01a6\u01a7"+
		"\u0003\\.\u0000\u01a7\u01a8\u0005#\u0000\u0000\u01a8\u01a9\u00038\u001c"+
		"\u0000\u01a9M\u0001\u0000\u0000\u0000\u01aa\u01ab\u0005\b\u0000\u0000"+
		"\u01abO\u0001\u0000\u0000\u0000\u01ac\u01ad\u0005\r\u0000\u0000\u01ad"+
		"Q\u0001\u0000\u0000\u0000\u01ae\u01af\u0005!\u0000\u0000\u01af\u01b0\u0003"+
		"\\.\u0000\u01b0\u01b1\u0005b\u0000\u0000\u01b1\u01b5\u0003$\u0012\u0000"+
		"\u01b2\u01b3\u0005\u0011\u0000\u0000\u01b3\u01b4\u0005b\u0000\u0000\u01b4"+
		"\u01b6\u0003$\u0012\u0000\u01b5\u01b2\u0001\u0000\u0000\u0000\u01b5\u01b6"+
		"\u0001\u0000\u0000\u0000\u01b6\u01b7\u0001\u0000\u0000\u0000\u01b7\u01b8"+
		"\u0005\u0016\u0000\u0000\u01b8S\u0001\u0000\u0000\u0000\u01b9\u01ba\u0005"+
		">\u0000\u0000\u01ba\u01bb\u0003\\.\u0000\u01bb\u01bc\u0005b\u0000\u0000"+
		"\u01bc\u01bd\u0003$\u0012\u0000\u01bd\u01be\u0005\u0018\u0000\u0000\u01be"+
		"U\u0001\u0000\u0000\u0000\u01bf\u01c0\u0005\u001e\u0000\u0000\u01c0\u01c1"+
		"\u0005\u0010\u0000\u0000\u01c1\u01c2\u0005D\u0000\u0000\u01c2\u01c3\u0005"+
		"\"\u0000\u0000\u01c3\u01c4\u0003\\.\u0000\u01c4\u01c5\u0005b\u0000\u0000"+
		"\u01c5\u01c6\u0003$\u0012\u0000\u01c6\u01c7\u0005\u0014\u0000\u0000\u01c7"+
		"W\u0001\u0000\u0000\u0000\u01c8\u01cd\u0005D\u0000\u0000\u01c9\u01ca\u0005"+
		"X\u0000\u0000\u01ca\u01cc\u0005D\u0000\u0000\u01cb\u01c9\u0001\u0000\u0000"+
		"\u0000\u01cc\u01cf\u0001\u0000\u0000\u0000\u01cd\u01cb\u0001\u0000\u0000"+
		"\u0000\u01cd\u01ce\u0001\u0000\u0000\u0000\u01ceY\u0001\u0000\u0000\u0000"+
		"\u01cf\u01cd\u0001\u0000\u0000\u0000\u01d0\u01d5\u0005D\u0000\u0000\u01d1"+
		"\u01d2\u0005;\u0000\u0000\u01d2\u01d3\u0005X\u0000\u0000\u01d3\u01d5\u0003"+
		"X,\u0000\u01d4\u01d0\u0001\u0000\u0000\u0000\u01d4\u01d1\u0001\u0000\u0000"+
		"\u0000\u01d5[\u0001\u0000\u0000\u0000\u01d6\u01d7\u0003^/\u0000\u01d7"+
		"]\u0001\u0000\u0000\u0000\u01d8\u01dd\u0003`0\u0000\u01d9\u01da\u0005"+
		"2\u0000\u0000\u01da\u01dc\u0003`0\u0000\u01db\u01d9\u0001\u0000\u0000"+
		"\u0000\u01dc\u01df\u0001\u0000\u0000\u0000\u01dd\u01db\u0001\u0000\u0000"+
		"\u0000\u01dd\u01de\u0001\u0000\u0000\u0000\u01de_\u0001\u0000\u0000\u0000"+
		"\u01df\u01dd\u0001\u0000\u0000\u0000\u01e0\u01e5\u0003b1\u0000\u01e1\u01e2"+
		"\u0005\u0003\u0000\u0000\u01e2\u01e4\u0003b1\u0000\u01e3\u01e1\u0001\u0000"+
		"\u0000\u0000\u01e4\u01e7\u0001\u0000\u0000\u0000\u01e5\u01e3\u0001\u0000"+
		"\u0000\u0000\u01e5\u01e6\u0001\u0000\u0000\u0000\u01e6a\u0001\u0000\u0000"+
		"\u0000\u01e7\u01e5\u0001\u0000\u0000\u0000\u01e8\u01ed\u0003d2\u0000\u01e9"+
		"\u01ea\u0005M\u0000\u0000\u01ea\u01ec\u0003d2\u0000\u01eb\u01e9\u0001"+
		"\u0000\u0000\u0000\u01ec\u01ef\u0001\u0000\u0000\u0000\u01ed\u01eb\u0001"+
		"\u0000\u0000\u0000\u01ed\u01ee\u0001\u0000\u0000\u0000\u01eec\u0001\u0000"+
		"\u0000\u0000\u01ef\u01ed\u0001\u0000\u0000\u0000\u01f0\u01f5\u0003f3\u0000"+
		"\u01f1\u01f2\u0005N\u0000\u0000\u01f2\u01f4\u0003f3\u0000\u01f3\u01f1"+
		"\u0001\u0000\u0000\u0000\u01f4\u01f7\u0001\u0000\u0000\u0000\u01f5\u01f3"+
		"\u0001\u0000\u0000\u0000\u01f5\u01f6\u0001\u0000\u0000\u0000\u01f6e\u0001"+
		"\u0000\u0000\u0000\u01f7\u01f5\u0001\u0000\u0000\u0000\u01f8\u01fd\u0003"+
		"h4\u0000\u01f9\u01fa\u0005L\u0000\u0000\u01fa\u01fc\u0003h4\u0000\u01fb"+
		"\u01f9\u0001\u0000\u0000\u0000\u01fc\u01ff\u0001\u0000\u0000\u0000\u01fd"+
		"\u01fb\u0001\u0000\u0000\u0000\u01fd\u01fe\u0001\u0000\u0000\u0000\u01fe"+
		"g\u0001\u0000\u0000\u0000\u01ff\u01fd\u0001\u0000\u0000\u0000\u0200\u0205"+
		"\u0003j5\u0000\u0201\u0202\u0007\u0001\u0000\u0000\u0202\u0204\u0003j"+
		"5\u0000\u0203\u0201\u0001\u0000\u0000\u0000\u0204\u0207\u0001\u0000\u0000"+
		"\u0000\u0205\u0203\u0001\u0000\u0000\u0000\u0205\u0206\u0001\u0000\u0000"+
		"\u0000\u0206i\u0001\u0000\u0000\u0000\u0207\u0205\u0001\u0000\u0000\u0000"+
		"\u0208\u020d\u0003l6\u0000\u0209\u020a\u0007\u0002\u0000\u0000\u020a\u020c"+
		"\u0003l6\u0000\u020b\u0209\u0001\u0000\u0000\u0000\u020c\u020f\u0001\u0000"+
		"\u0000\u0000\u020d\u020b\u0001\u0000\u0000\u0000\u020d\u020e\u0001\u0000"+
		"\u0000\u0000\u020ek\u0001\u0000\u0000\u0000\u020f\u020d\u0001\u0000\u0000"+
		"\u0000\u0210\u0215\u0003n7\u0000\u0211\u0212\u0007\u0003\u0000\u0000\u0212"+
		"\u0214\u0003n7\u0000\u0213\u0211\u0001\u0000\u0000\u0000\u0214\u0217\u0001"+
		"\u0000\u0000\u0000\u0215\u0213\u0001\u0000\u0000\u0000\u0215\u0216\u0001"+
		"\u0000\u0000\u0000\u0216m\u0001\u0000\u0000\u0000\u0217\u0215\u0001\u0000"+
		"\u0000\u0000\u0218\u021d\u0003p8\u0000\u0219\u021a\u0007\u0004\u0000\u0000"+
		"\u021a\u021c\u0003p8\u0000\u021b\u0219\u0001\u0000\u0000\u0000\u021c\u021f"+
		"\u0001\u0000\u0000\u0000\u021d\u021b\u0001\u0000\u0000\u0000\u021d\u021e"+
		"\u0001\u0000\u0000\u0000\u021eo\u0001\u0000\u0000\u0000\u021f\u021d\u0001"+
		"\u0000\u0000\u0000\u0220\u0221\u0007\u0005\u0000\u0000\u0221\u0226\u0003"+
		"p8\u0000\u0222\u0223\u0005=\u0000\u0000\u0223\u0226\u0003p8\u0000\u0224"+
		"\u0226\u0003r9\u0000\u0225\u0220\u0001\u0000\u0000\u0000\u0225\u0222\u0001"+
		"\u0000\u0000\u0000\u0225\u0224\u0001\u0000\u0000\u0000\u0226q\u0001\u0000"+
		"\u0000\u0000\u0227\u022a\u0003t:\u0000\u0228\u0229\u0005K\u0000\u0000"+
		"\u0229\u022b\u0003r9\u0000\u022a\u0228\u0001\u0000\u0000\u0000\u022a\u022b"+
		"\u0001\u0000\u0000\u0000\u022bs\u0001\u0000\u0000\u0000\u022c\u0233\u0003"+
		"v;\u0000\u022d\u022e\u0005S\u0000\u0000\u022e\u022f\u0003\\.\u0000\u022f"+
		"\u0230\u0005T\u0000\u0000\u0230\u0232\u0001\u0000\u0000\u0000\u0231\u022d"+
		"\u0001\u0000\u0000\u0000\u0232\u0235\u0001\u0000\u0000\u0000\u0233\u0231"+
		"\u0001\u0000\u0000\u0000\u0233\u0234\u0001\u0000\u0000\u0000\u0234u\u0001"+
		"\u0000\u0000\u0000\u0235\u0233\u0001\u0000\u0000\u0000\u0236\u0245\u0003"+
		"|>\u0000\u0237\u0245\u0003z=\u0000\u0238\u0245\u0005D\u0000\u0000\u0239"+
		"\u0245\u0005$\u0000\u0000\u023a\u0245\u0003x<\u0000\u023b\u023c\u0005"+
		"\u001a\u0000\u0000\u023c\u023d\u0005P\u0000\u0000\u023d\u023e\u0003\\"+
		".\u0000\u023e\u023f\u0005Q\u0000\u0000\u023f\u0245\u0001\u0000\u0000\u0000"+
		"\u0240\u0241\u0005P\u0000\u0000\u0241\u0242\u0003\\.\u0000\u0242\u0243"+
		"\u0005Q\u0000\u0000\u0243\u0245\u0001\u0000\u0000\u0000\u0244\u0236\u0001"+
		"\u0000\u0000\u0000\u0244\u0237\u0001\u0000\u0000\u0000\u0244\u0238\u0001"+
		"\u0000\u0000\u0000\u0244\u0239\u0001\u0000\u0000\u0000\u0244\u023a\u0001"+
		"\u0000\u0000\u0000\u0244\u023b\u0001\u0000\u0000\u0000\u0244\u0240\u0001"+
		"\u0000\u0000\u0000\u0245w\u0001\u0000\u0000\u0000\u0246\u0251\u0003Z-"+
		"\u0000\u0247\u0251\u0005&\u0000\u0000\u0248\u0251\u0005\'\u0000\u0000"+
		"\u0249\u0251\u00057\u0000\u0000\u024a\u0251\u0005\u000e\u0000\u0000\u024b"+
		"\u0251\u00059\u0000\u0000\u024c\u0251\u0005\u0005\u0000\u0000\u024d\u0251"+
		"\u0005\u0002\u0000\u0000\u024e\u0251\u0005\u0007\u0000\u0000\u024f\u0251"+
		"\u0005%\u0000\u0000\u0250\u0246\u0001\u0000\u0000\u0000\u0250\u0247\u0001"+
		"\u0000\u0000\u0000\u0250\u0248\u0001\u0000\u0000\u0000\u0250\u0249\u0001"+
		"\u0000\u0000\u0000\u0250\u024a\u0001\u0000\u0000\u0000\u0250\u024b\u0001"+
		"\u0000\u0000\u0000\u0250\u024c\u0001\u0000\u0000\u0000\u0250\u024d\u0001"+
		"\u0000\u0000\u0000\u0250\u024e\u0001\u0000\u0000\u0000\u0250\u024f\u0001"+
		"\u0000\u0000\u0000\u0251\u0252\u0001\u0000\u0000\u0000\u0252\u0253\u0005"+
		"P\u0000\u0000\u0253\u0254\u0003\u0086C\u0000\u0254\u0255\u0005Q\u0000"+
		"\u0000\u0255y\u0001\u0000\u0000\u0000\u0256\u0257\u0005Y\u0000\u0000\u0257"+
		"\u0258\u0007\u0006\u0000\u0000\u0258\u0259\u0005Z\u0000\u0000\u0259{\u0001"+
		"\u0000\u0000\u0000\u025a\u0262\u0005@\u0000\u0000\u025b\u0262\u0005A\u0000"+
		"\u0000\u025c\u0262\u0005C\u0000\u0000\u025d\u0262\u0003\u0082A\u0000\u025e"+
		"\u0262\u0003\u0084B\u0000\u025f\u0262\u0003\u0080@\u0000\u0260\u0262\u0003"+
		"~?\u0000\u0261\u025a\u0001\u0000\u0000\u0000\u0261\u025b\u0001\u0000\u0000"+
		"\u0000\u0261\u025c\u0001\u0000\u0000\u0000\u0261\u025d\u0001\u0000\u0000"+
		"\u0000\u0261\u025e\u0001\u0000\u0000\u0000\u0261\u025f\u0001\u0000\u0000"+
		"\u0000\u0261\u0260\u0001\u0000\u0000\u0000\u0262}\u0001\u0000\u0000\u0000"+
		"\u0263\u0264\u0005-\u0000\u0000\u0264\u007f\u0001\u0000\u0000\u0000\u0265"+
		"\u0266\u0007\u0007\u0000\u0000\u0266\u0081\u0001\u0000\u0000\u0000\u0267"+
		"\u0268\u0005S\u0000\u0000\u0268\u0269\u0003\u0086C\u0000\u0269\u026a\u0005"+
		"T\u0000\u0000\u026a\u0083\u0001\u0000\u0000\u0000\u026b\u026c\u0005U\u0000"+
		"\u0000\u026c\u026d\u0003\u008aE\u0000\u026d\u026e\u0005V\u0000\u0000\u026e"+
		"\u0085\u0001\u0000\u0000\u0000\u026f\u0271\u0003\u0088D\u0000\u0270\u026f"+
		"\u0001\u0000\u0000\u0000\u0270\u0271\u0001\u0000\u0000\u0000\u0271\u0087"+
		"\u0001\u0000\u0000\u0000\u0272\u0277\u0003\\.\u0000\u0273\u0274\u0005"+
		"R\u0000\u0000\u0274\u0276\u0003\\.\u0000\u0275\u0273\u0001\u0000\u0000"+
		"\u0000\u0276\u0279\u0001\u0000\u0000\u0000\u0277\u0275\u0001\u0000\u0000"+
		"\u0000\u0277\u0278\u0001\u0000\u0000\u0000\u0278\u0089\u0001\u0000\u0000"+
		"\u0000\u0279\u0277\u0001\u0000\u0000\u0000\u027a\u027c\u0003\u008cF\u0000"+
		"\u027b\u027a\u0001\u0000\u0000\u0000\u027b\u027c\u0001\u0000\u0000\u0000"+
		"\u027c\u008b\u0001\u0000\u0000\u0000\u027d\u0282\u0003\u008eG\u0000\u027e"+
		"\u027f\u0005R\u0000\u0000\u027f\u0281\u0003\u008eG\u0000\u0280\u027e\u0001"+
		"\u0000\u0000\u0000\u0281\u0284\u0001\u0000\u0000\u0000\u0282\u0280\u0001"+
		"\u0000\u0000\u0000\u0282\u0283\u0001\u0000\u0000\u0000\u0283\u008d\u0001"+
		"\u0000\u0000\u0000\u0284\u0282\u0001\u0000\u0000\u0000\u0285\u0286\u0005"+
		"@\u0000\u0000\u0286\u0287\u0005W\u0000\u0000\u0287\u0288\u0003\\.\u0000"+
		"\u0288\u008f\u0001\u0000\u0000\u0000;\u0093\u009a\u00a0\u00a5\u00aa\u00af"+
		"\u00ba\u00c0\u00c8\u00cf\u00d4\u00e3\u00f2\u00f4\u00fb\u00fd\u0100\u0110"+
		"\u0117\u011d\u0125\u012b\u0132\u0137\u0145\u014a\u014f\u015b\u015f\u016b"+
		"\u0174\u0176\u017e\u018b\u0195\u019f\u01a3\u01b5\u01cd\u01d4\u01dd\u01e5"+
		"\u01ed\u01f5\u01fd\u0205\u020d\u0215\u021d\u0225\u022a\u0233\u0244\u0250"+
		"\u0261\u0270\u0277\u027b\u0282";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}