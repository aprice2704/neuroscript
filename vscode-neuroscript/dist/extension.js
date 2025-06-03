"use strict";
var __create = Object.create;
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __getProtoOf = Object.getPrototypeOf;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __commonJS = (cb, mod) => function __require() {
  return mod || (0, cb[__getOwnPropNames(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
};
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toESM = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(
  // If the importer is in node compatibility mode or this is not an ESM
  // file that has been converted to a CommonJS file using a Babel-
  // compatible transform (i.e. "__esModule" has not been set), then set
  // "default" to the CommonJS "module.exports" for node compatibility.
  isNodeMode || !mod || !mod.__esModule ? __defProp(target, "default", { value: mod, enumerable: true }) : target,
  mod
));
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// node_modules/semver/internal/constants.js
var require_constants = __commonJS({
  "node_modules/semver/internal/constants.js"(exports2, module2) {
    "use strict";
    var SEMVER_SPEC_VERSION = "2.0.0";
    var MAX_LENGTH = 256;
    var MAX_SAFE_INTEGER = Number.MAX_SAFE_INTEGER || /* istanbul ignore next */
    9007199254740991;
    var MAX_SAFE_COMPONENT_LENGTH = 16;
    var MAX_SAFE_BUILD_LENGTH = MAX_LENGTH - 6;
    var RELEASE_TYPES = [
      "major",
      "premajor",
      "minor",
      "preminor",
      "patch",
      "prepatch",
      "prerelease"
    ];
    module2.exports = {
      MAX_LENGTH,
      MAX_SAFE_COMPONENT_LENGTH,
      MAX_SAFE_BUILD_LENGTH,
      MAX_SAFE_INTEGER,
      RELEASE_TYPES,
      SEMVER_SPEC_VERSION,
      FLAG_INCLUDE_PRERELEASE: 1,
      FLAG_LOOSE: 2
    };
  }
});

// node_modules/semver/internal/debug.js
var require_debug = __commonJS({
  "node_modules/semver/internal/debug.js"(exports2, module2) {
    "use strict";
    var debug = typeof process === "object" && process.env && process.env.NODE_DEBUG && /\bsemver\b/i.test(process.env.NODE_DEBUG) ? (...args) => console.error("SEMVER", ...args) : () => {
    };
    module2.exports = debug;
  }
});

// node_modules/semver/internal/re.js
var require_re = __commonJS({
  "node_modules/semver/internal/re.js"(exports2, module2) {
    "use strict";
    var {
      MAX_SAFE_COMPONENT_LENGTH,
      MAX_SAFE_BUILD_LENGTH,
      MAX_LENGTH
    } = require_constants();
    var debug = require_debug();
    exports2 = module2.exports = {};
    var re = exports2.re = [];
    var safeRe = exports2.safeRe = [];
    var src = exports2.src = [];
    var safeSrc = exports2.safeSrc = [];
    var t = exports2.t = {};
    var R = 0;
    var LETTERDASHNUMBER = "[a-zA-Z0-9-]";
    var safeRegexReplacements = [
      ["\\s", 1],
      ["\\d", MAX_LENGTH],
      [LETTERDASHNUMBER, MAX_SAFE_BUILD_LENGTH]
    ];
    var makeSafeRegex = (value) => {
      for (const [token, max] of safeRegexReplacements) {
        value = value.split(`${token}*`).join(`${token}{0,${max}}`).split(`${token}+`).join(`${token}{1,${max}}`);
      }
      return value;
    };
    var createToken = (name, value, isGlobal) => {
      const safe = makeSafeRegex(value);
      const index = R++;
      debug(name, index, value);
      t[name] = index;
      src[index] = value;
      safeSrc[index] = safe;
      re[index] = new RegExp(value, isGlobal ? "g" : void 0);
      safeRe[index] = new RegExp(safe, isGlobal ? "g" : void 0);
    };
    createToken("NUMERICIDENTIFIER", "0|[1-9]\\d*");
    createToken("NUMERICIDENTIFIERLOOSE", "\\d+");
    createToken("NONNUMERICIDENTIFIER", `\\d*[a-zA-Z-]${LETTERDASHNUMBER}*`);
    createToken("MAINVERSION", `(${src[t.NUMERICIDENTIFIER]})\\.(${src[t.NUMERICIDENTIFIER]})\\.(${src[t.NUMERICIDENTIFIER]})`);
    createToken("MAINVERSIONLOOSE", `(${src[t.NUMERICIDENTIFIERLOOSE]})\\.(${src[t.NUMERICIDENTIFIERLOOSE]})\\.(${src[t.NUMERICIDENTIFIERLOOSE]})`);
    createToken("PRERELEASEIDENTIFIER", `(?:${src[t.NONNUMERICIDENTIFIER]}|${src[t.NUMERICIDENTIFIER]})`);
    createToken("PRERELEASEIDENTIFIERLOOSE", `(?:${src[t.NONNUMERICIDENTIFIER]}|${src[t.NUMERICIDENTIFIERLOOSE]})`);
    createToken("PRERELEASE", `(?:-(${src[t.PRERELEASEIDENTIFIER]}(?:\\.${src[t.PRERELEASEIDENTIFIER]})*))`);
    createToken("PRERELEASELOOSE", `(?:-?(${src[t.PRERELEASEIDENTIFIERLOOSE]}(?:\\.${src[t.PRERELEASEIDENTIFIERLOOSE]})*))`);
    createToken("BUILDIDENTIFIER", `${LETTERDASHNUMBER}+`);
    createToken("BUILD", `(?:\\+(${src[t.BUILDIDENTIFIER]}(?:\\.${src[t.BUILDIDENTIFIER]})*))`);
    createToken("FULLPLAIN", `v?${src[t.MAINVERSION]}${src[t.PRERELEASE]}?${src[t.BUILD]}?`);
    createToken("FULL", `^${src[t.FULLPLAIN]}$`);
    createToken("LOOSEPLAIN", `[v=\\s]*${src[t.MAINVERSIONLOOSE]}${src[t.PRERELEASELOOSE]}?${src[t.BUILD]}?`);
    createToken("LOOSE", `^${src[t.LOOSEPLAIN]}$`);
    createToken("GTLT", "((?:<|>)?=?)");
    createToken("XRANGEIDENTIFIERLOOSE", `${src[t.NUMERICIDENTIFIERLOOSE]}|x|X|\\*`);
    createToken("XRANGEIDENTIFIER", `${src[t.NUMERICIDENTIFIER]}|x|X|\\*`);
    createToken("XRANGEPLAIN", `[v=\\s]*(${src[t.XRANGEIDENTIFIER]})(?:\\.(${src[t.XRANGEIDENTIFIER]})(?:\\.(${src[t.XRANGEIDENTIFIER]})(?:${src[t.PRERELEASE]})?${src[t.BUILD]}?)?)?`);
    createToken("XRANGEPLAINLOOSE", `[v=\\s]*(${src[t.XRANGEIDENTIFIERLOOSE]})(?:\\.(${src[t.XRANGEIDENTIFIERLOOSE]})(?:\\.(${src[t.XRANGEIDENTIFIERLOOSE]})(?:${src[t.PRERELEASELOOSE]})?${src[t.BUILD]}?)?)?`);
    createToken("XRANGE", `^${src[t.GTLT]}\\s*${src[t.XRANGEPLAIN]}$`);
    createToken("XRANGELOOSE", `^${src[t.GTLT]}\\s*${src[t.XRANGEPLAINLOOSE]}$`);
    createToken("COERCEPLAIN", `${"(^|[^\\d])(\\d{1,"}${MAX_SAFE_COMPONENT_LENGTH}})(?:\\.(\\d{1,${MAX_SAFE_COMPONENT_LENGTH}}))?(?:\\.(\\d{1,${MAX_SAFE_COMPONENT_LENGTH}}))?`);
    createToken("COERCE", `${src[t.COERCEPLAIN]}(?:$|[^\\d])`);
    createToken("COERCEFULL", src[t.COERCEPLAIN] + `(?:${src[t.PRERELEASE]})?(?:${src[t.BUILD]})?(?:$|[^\\d])`);
    createToken("COERCERTL", src[t.COERCE], true);
    createToken("COERCERTLFULL", src[t.COERCEFULL], true);
    createToken("LONETILDE", "(?:~>?)");
    createToken("TILDETRIM", `(\\s*)${src[t.LONETILDE]}\\s+`, true);
    exports2.tildeTrimReplace = "$1~";
    createToken("TILDE", `^${src[t.LONETILDE]}${src[t.XRANGEPLAIN]}$`);
    createToken("TILDELOOSE", `^${src[t.LONETILDE]}${src[t.XRANGEPLAINLOOSE]}$`);
    createToken("LONECARET", "(?:\\^)");
    createToken("CARETTRIM", `(\\s*)${src[t.LONECARET]}\\s+`, true);
    exports2.caretTrimReplace = "$1^";
    createToken("CARET", `^${src[t.LONECARET]}${src[t.XRANGEPLAIN]}$`);
    createToken("CARETLOOSE", `^${src[t.LONECARET]}${src[t.XRANGEPLAINLOOSE]}$`);
    createToken("COMPARATORLOOSE", `^${src[t.GTLT]}\\s*(${src[t.LOOSEPLAIN]})$|^$`);
    createToken("COMPARATOR", `^${src[t.GTLT]}\\s*(${src[t.FULLPLAIN]})$|^$`);
    createToken("COMPARATORTRIM", `(\\s*)${src[t.GTLT]}\\s*(${src[t.LOOSEPLAIN]}|${src[t.XRANGEPLAIN]})`, true);
    exports2.comparatorTrimReplace = "$1$2$3";
    createToken("HYPHENRANGE", `^\\s*(${src[t.XRANGEPLAIN]})\\s+-\\s+(${src[t.XRANGEPLAIN]})\\s*$`);
    createToken("HYPHENRANGELOOSE", `^\\s*(${src[t.XRANGEPLAINLOOSE]})\\s+-\\s+(${src[t.XRANGEPLAINLOOSE]})\\s*$`);
    createToken("STAR", "(<|>)?=?\\s*\\*");
    createToken("GTE0", "^\\s*>=\\s*0\\.0\\.0\\s*$");
    createToken("GTE0PRE", "^\\s*>=\\s*0\\.0\\.0-0\\s*$");
  }
});

// node_modules/semver/internal/parse-options.js
var require_parse_options = __commonJS({
  "node_modules/semver/internal/parse-options.js"(exports2, module2) {
    "use strict";
    var looseOption = Object.freeze({ loose: true });
    var emptyOpts = Object.freeze({});
    var parseOptions = (options) => {
      if (!options) {
        return emptyOpts;
      }
      if (typeof options !== "object") {
        return looseOption;
      }
      return options;
    };
    module2.exports = parseOptions;
  }
});

// node_modules/semver/internal/identifiers.js
var require_identifiers = __commonJS({
  "node_modules/semver/internal/identifiers.js"(exports2, module2) {
    "use strict";
    var numeric = /^[0-9]+$/;
    var compareIdentifiers = (a, b) => {
      const anum = numeric.test(a);
      const bnum = numeric.test(b);
      if (anum && bnum) {
        a = +a;
        b = +b;
      }
      return a === b ? 0 : anum && !bnum ? -1 : bnum && !anum ? 1 : a < b ? -1 : 1;
    };
    var rcompareIdentifiers = (a, b) => compareIdentifiers(b, a);
    module2.exports = {
      compareIdentifiers,
      rcompareIdentifiers
    };
  }
});

// node_modules/semver/classes/semver.js
var require_semver = __commonJS({
  "node_modules/semver/classes/semver.js"(exports2, module2) {
    "use strict";
    var debug = require_debug();
    var { MAX_LENGTH, MAX_SAFE_INTEGER } = require_constants();
    var { safeRe: re, t } = require_re();
    var parseOptions = require_parse_options();
    var { compareIdentifiers } = require_identifiers();
    var SemVer = class _SemVer {
      constructor(version, options) {
        options = parseOptions(options);
        if (version instanceof _SemVer) {
          if (version.loose === !!options.loose && version.includePrerelease === !!options.includePrerelease) {
            return version;
          } else {
            version = version.version;
          }
        } else if (typeof version !== "string") {
          throw new TypeError(`Invalid version. Must be a string. Got type "${typeof version}".`);
        }
        if (version.length > MAX_LENGTH) {
          throw new TypeError(
            `version is longer than ${MAX_LENGTH} characters`
          );
        }
        debug("SemVer", version, options);
        this.options = options;
        this.loose = !!options.loose;
        this.includePrerelease = !!options.includePrerelease;
        const m = version.trim().match(options.loose ? re[t.LOOSE] : re[t.FULL]);
        if (!m) {
          throw new TypeError(`Invalid Version: ${version}`);
        }
        this.raw = version;
        this.major = +m[1];
        this.minor = +m[2];
        this.patch = +m[3];
        if (this.major > MAX_SAFE_INTEGER || this.major < 0) {
          throw new TypeError("Invalid major version");
        }
        if (this.minor > MAX_SAFE_INTEGER || this.minor < 0) {
          throw new TypeError("Invalid minor version");
        }
        if (this.patch > MAX_SAFE_INTEGER || this.patch < 0) {
          throw new TypeError("Invalid patch version");
        }
        if (!m[4]) {
          this.prerelease = [];
        } else {
          this.prerelease = m[4].split(".").map((id) => {
            if (/^[0-9]+$/.test(id)) {
              const num = +id;
              if (num >= 0 && num < MAX_SAFE_INTEGER) {
                return num;
              }
            }
            return id;
          });
        }
        this.build = m[5] ? m[5].split(".") : [];
        this.format();
      }
      format() {
        this.version = `${this.major}.${this.minor}.${this.patch}`;
        if (this.prerelease.length) {
          this.version += `-${this.prerelease.join(".")}`;
        }
        return this.version;
      }
      toString() {
        return this.version;
      }
      compare(other) {
        debug("SemVer.compare", this.version, this.options, other);
        if (!(other instanceof _SemVer)) {
          if (typeof other === "string" && other === this.version) {
            return 0;
          }
          other = new _SemVer(other, this.options);
        }
        if (other.version === this.version) {
          return 0;
        }
        return this.compareMain(other) || this.comparePre(other);
      }
      compareMain(other) {
        if (!(other instanceof _SemVer)) {
          other = new _SemVer(other, this.options);
        }
        return compareIdentifiers(this.major, other.major) || compareIdentifiers(this.minor, other.minor) || compareIdentifiers(this.patch, other.patch);
      }
      comparePre(other) {
        if (!(other instanceof _SemVer)) {
          other = new _SemVer(other, this.options);
        }
        if (this.prerelease.length && !other.prerelease.length) {
          return -1;
        } else if (!this.prerelease.length && other.prerelease.length) {
          return 1;
        } else if (!this.prerelease.length && !other.prerelease.length) {
          return 0;
        }
        let i = 0;
        do {
          const a = this.prerelease[i];
          const b = other.prerelease[i];
          debug("prerelease compare", i, a, b);
          if (a === void 0 && b === void 0) {
            return 0;
          } else if (b === void 0) {
            return 1;
          } else if (a === void 0) {
            return -1;
          } else if (a === b) {
            continue;
          } else {
            return compareIdentifiers(a, b);
          }
        } while (++i);
      }
      compareBuild(other) {
        if (!(other instanceof _SemVer)) {
          other = new _SemVer(other, this.options);
        }
        let i = 0;
        do {
          const a = this.build[i];
          const b = other.build[i];
          debug("build compare", i, a, b);
          if (a === void 0 && b === void 0) {
            return 0;
          } else if (b === void 0) {
            return 1;
          } else if (a === void 0) {
            return -1;
          } else if (a === b) {
            continue;
          } else {
            return compareIdentifiers(a, b);
          }
        } while (++i);
      }
      // preminor will bump the version up to the next minor release, and immediately
      // down to pre-release. premajor and prepatch work the same way.
      inc(release, identifier, identifierBase) {
        if (release.startsWith("pre")) {
          if (!identifier && identifierBase === false) {
            throw new Error("invalid increment argument: identifier is empty");
          }
          if (identifier) {
            const match = `-${identifier}`.match(this.options.loose ? re[t.PRERELEASELOOSE] : re[t.PRERELEASE]);
            if (!match || match[1] !== identifier) {
              throw new Error(`invalid identifier: ${identifier}`);
            }
          }
        }
        switch (release) {
          case "premajor":
            this.prerelease.length = 0;
            this.patch = 0;
            this.minor = 0;
            this.major++;
            this.inc("pre", identifier, identifierBase);
            break;
          case "preminor":
            this.prerelease.length = 0;
            this.patch = 0;
            this.minor++;
            this.inc("pre", identifier, identifierBase);
            break;
          case "prepatch":
            this.prerelease.length = 0;
            this.inc("patch", identifier, identifierBase);
            this.inc("pre", identifier, identifierBase);
            break;
          // If the input is a non-prerelease version, this acts the same as
          // prepatch.
          case "prerelease":
            if (this.prerelease.length === 0) {
              this.inc("patch", identifier, identifierBase);
            }
            this.inc("pre", identifier, identifierBase);
            break;
          case "release":
            if (this.prerelease.length === 0) {
              throw new Error(`version ${this.raw} is not a prerelease`);
            }
            this.prerelease.length = 0;
            break;
          case "major":
            if (this.minor !== 0 || this.patch !== 0 || this.prerelease.length === 0) {
              this.major++;
            }
            this.minor = 0;
            this.patch = 0;
            this.prerelease = [];
            break;
          case "minor":
            if (this.patch !== 0 || this.prerelease.length === 0) {
              this.minor++;
            }
            this.patch = 0;
            this.prerelease = [];
            break;
          case "patch":
            if (this.prerelease.length === 0) {
              this.patch++;
            }
            this.prerelease = [];
            break;
          // This probably shouldn't be used publicly.
          // 1.0.0 'pre' would become 1.0.0-0 which is the wrong direction.
          case "pre": {
            const base = Number(identifierBase) ? 1 : 0;
            if (this.prerelease.length === 0) {
              this.prerelease = [base];
            } else {
              let i = this.prerelease.length;
              while (--i >= 0) {
                if (typeof this.prerelease[i] === "number") {
                  this.prerelease[i]++;
                  i = -2;
                }
              }
              if (i === -1) {
                if (identifier === this.prerelease.join(".") && identifierBase === false) {
                  throw new Error("invalid increment argument: identifier already exists");
                }
                this.prerelease.push(base);
              }
            }
            if (identifier) {
              let prerelease = [identifier, base];
              if (identifierBase === false) {
                prerelease = [identifier];
              }
              if (compareIdentifiers(this.prerelease[0], identifier) === 0) {
                if (isNaN(this.prerelease[1])) {
                  this.prerelease = prerelease;
                }
              } else {
                this.prerelease = prerelease;
              }
            }
            break;
          }
          default:
            throw new Error(`invalid increment argument: ${release}`);
        }
        this.raw = this.format();
        if (this.build.length) {
          this.raw += `+${this.build.join(".")}`;
        }
        return this;
      }
    };
    module2.exports = SemVer;
  }
});

// node_modules/semver/functions/parse.js
var require_parse = __commonJS({
  "node_modules/semver/functions/parse.js"(exports2, module2) {
    "use strict";
    var SemVer = require_semver();
    var parse = (version, options, throwErrors = false) => {
      if (version instanceof SemVer) {
        return version;
      }
      try {
        return new SemVer(version, options);
      } catch (er) {
        if (!throwErrors) {
          return null;
        }
        throw er;
      }
    };
    module2.exports = parse;
  }
});

// node_modules/semver/functions/valid.js
var require_valid = __commonJS({
  "node_modules/semver/functions/valid.js"(exports2, module2) {
    "use strict";
    var parse = require_parse();
    var valid = (version, options) => {
      const v = parse(version, options);
      return v ? v.version : null;
    };
    module2.exports = valid;
  }
});

// node_modules/semver/functions/clean.js
var require_clean = __commonJS({
  "node_modules/semver/functions/clean.js"(exports2, module2) {
    "use strict";
    var parse = require_parse();
    var clean = (version, options) => {
      const s = parse(version.trim().replace(/^[=v]+/, ""), options);
      return s ? s.version : null;
    };
    module2.exports = clean;
  }
});

// node_modules/semver/functions/inc.js
var require_inc = __commonJS({
  "node_modules/semver/functions/inc.js"(exports2, module2) {
    "use strict";
    var SemVer = require_semver();
    var inc = (version, release, options, identifier, identifierBase) => {
      if (typeof options === "string") {
        identifierBase = identifier;
        identifier = options;
        options = void 0;
      }
      try {
        return new SemVer(
          version instanceof SemVer ? version.version : version,
          options
        ).inc(release, identifier, identifierBase).version;
      } catch (er) {
        return null;
      }
    };
    module2.exports = inc;
  }
});

// node_modules/semver/functions/diff.js
var require_diff = __commonJS({
  "node_modules/semver/functions/diff.js"(exports2, module2) {
    "use strict";
    var parse = require_parse();
    var diff = (version1, version2) => {
      const v1 = parse(version1, null, true);
      const v2 = parse(version2, null, true);
      const comparison = v1.compare(v2);
      if (comparison === 0) {
        return null;
      }
      const v1Higher = comparison > 0;
      const highVersion = v1Higher ? v1 : v2;
      const lowVersion = v1Higher ? v2 : v1;
      const highHasPre = !!highVersion.prerelease.length;
      const lowHasPre = !!lowVersion.prerelease.length;
      if (lowHasPre && !highHasPre) {
        if (!lowVersion.patch && !lowVersion.minor) {
          return "major";
        }
        if (lowVersion.compareMain(highVersion) === 0) {
          if (lowVersion.minor && !lowVersion.patch) {
            return "minor";
          }
          return "patch";
        }
      }
      const prefix = highHasPre ? "pre" : "";
      if (v1.major !== v2.major) {
        return prefix + "major";
      }
      if (v1.minor !== v2.minor) {
        return prefix + "minor";
      }
      if (v1.patch !== v2.patch) {
        return prefix + "patch";
      }
      return "prerelease";
    };
    module2.exports = diff;
  }
});

// node_modules/semver/functions/major.js
var require_major = __commonJS({
  "node_modules/semver/functions/major.js"(exports2, module2) {
    "use strict";
    var SemVer = require_semver();
    var major = (a, loose) => new SemVer(a, loose).major;
    module2.exports = major;
  }
});

// node_modules/semver/functions/minor.js
var require_minor = __commonJS({
  "node_modules/semver/functions/minor.js"(exports2, module2) {
    "use strict";
    var SemVer = require_semver();
    var minor = (a, loose) => new SemVer(a, loose).minor;
    module2.exports = minor;
  }
});

// node_modules/semver/functions/patch.js
var require_patch = __commonJS({
  "node_modules/semver/functions/patch.js"(exports2, module2) {
    "use strict";
    var SemVer = require_semver();
    var patch = (a, loose) => new SemVer(a, loose).patch;
    module2.exports = patch;
  }
});

// node_modules/semver/functions/prerelease.js
var require_prerelease = __commonJS({
  "node_modules/semver/functions/prerelease.js"(exports2, module2) {
    "use strict";
    var parse = require_parse();
    var prerelease = (version, options) => {
      const parsed = parse(version, options);
      return parsed && parsed.prerelease.length ? parsed.prerelease : null;
    };
    module2.exports = prerelease;
  }
});

// node_modules/semver/functions/compare.js
var require_compare = __commonJS({
  "node_modules/semver/functions/compare.js"(exports2, module2) {
    "use strict";
    var SemVer = require_semver();
    var compare = (a, b, loose) => new SemVer(a, loose).compare(new SemVer(b, loose));
    module2.exports = compare;
  }
});

// node_modules/semver/functions/rcompare.js
var require_rcompare = __commonJS({
  "node_modules/semver/functions/rcompare.js"(exports2, module2) {
    "use strict";
    var compare = require_compare();
    var rcompare = (a, b, loose) => compare(b, a, loose);
    module2.exports = rcompare;
  }
});

// node_modules/semver/functions/compare-loose.js
var require_compare_loose = __commonJS({
  "node_modules/semver/functions/compare-loose.js"(exports2, module2) {
    "use strict";
    var compare = require_compare();
    var compareLoose = (a, b) => compare(a, b, true);
    module2.exports = compareLoose;
  }
});

// node_modules/semver/functions/compare-build.js
var require_compare_build = __commonJS({
  "node_modules/semver/functions/compare-build.js"(exports2, module2) {
    "use strict";
    var SemVer = require_semver();
    var compareBuild = (a, b, loose) => {
      const versionA = new SemVer(a, loose);
      const versionB = new SemVer(b, loose);
      return versionA.compare(versionB) || versionA.compareBuild(versionB);
    };
    module2.exports = compareBuild;
  }
});

// node_modules/semver/functions/sort.js
var require_sort = __commonJS({
  "node_modules/semver/functions/sort.js"(exports2, module2) {
    "use strict";
    var compareBuild = require_compare_build();
    var sort = (list, loose) => list.sort((a, b) => compareBuild(a, b, loose));
    module2.exports = sort;
  }
});

// node_modules/semver/functions/rsort.js
var require_rsort = __commonJS({
  "node_modules/semver/functions/rsort.js"(exports2, module2) {
    "use strict";
    var compareBuild = require_compare_build();
    var rsort = (list, loose) => list.sort((a, b) => compareBuild(b, a, loose));
    module2.exports = rsort;
  }
});

// node_modules/semver/functions/gt.js
var require_gt = __commonJS({
  "node_modules/semver/functions/gt.js"(exports2, module2) {
    "use strict";
    var compare = require_compare();
    var gt = (a, b, loose) => compare(a, b, loose) > 0;
    module2.exports = gt;
  }
});

// node_modules/semver/functions/lt.js
var require_lt = __commonJS({
  "node_modules/semver/functions/lt.js"(exports2, module2) {
    "use strict";
    var compare = require_compare();
    var lt = (a, b, loose) => compare(a, b, loose) < 0;
    module2.exports = lt;
  }
});

// node_modules/semver/functions/eq.js
var require_eq = __commonJS({
  "node_modules/semver/functions/eq.js"(exports2, module2) {
    "use strict";
    var compare = require_compare();
    var eq = (a, b, loose) => compare(a, b, loose) === 0;
    module2.exports = eq;
  }
});

// node_modules/semver/functions/neq.js
var require_neq = __commonJS({
  "node_modules/semver/functions/neq.js"(exports2, module2) {
    "use strict";
    var compare = require_compare();
    var neq = (a, b, loose) => compare(a, b, loose) !== 0;
    module2.exports = neq;
  }
});

// node_modules/semver/functions/gte.js
var require_gte = __commonJS({
  "node_modules/semver/functions/gte.js"(exports2, module2) {
    "use strict";
    var compare = require_compare();
    var gte = (a, b, loose) => compare(a, b, loose) >= 0;
    module2.exports = gte;
  }
});

// node_modules/semver/functions/lte.js
var require_lte = __commonJS({
  "node_modules/semver/functions/lte.js"(exports2, module2) {
    "use strict";
    var compare = require_compare();
    var lte = (a, b, loose) => compare(a, b, loose) <= 0;
    module2.exports = lte;
  }
});

// node_modules/semver/functions/cmp.js
var require_cmp = __commonJS({
  "node_modules/semver/functions/cmp.js"(exports2, module2) {
    "use strict";
    var eq = require_eq();
    var neq = require_neq();
    var gt = require_gt();
    var gte = require_gte();
    var lt = require_lt();
    var lte = require_lte();
    var cmp = (a, op, b, loose) => {
      switch (op) {
        case "===":
          if (typeof a === "object") {
            a = a.version;
          }
          if (typeof b === "object") {
            b = b.version;
          }
          return a === b;
        case "!==":
          if (typeof a === "object") {
            a = a.version;
          }
          if (typeof b === "object") {
            b = b.version;
          }
          return a !== b;
        case "":
        case "=":
        case "==":
          return eq(a, b, loose);
        case "!=":
          return neq(a, b, loose);
        case ">":
          return gt(a, b, loose);
        case ">=":
          return gte(a, b, loose);
        case "<":
          return lt(a, b, loose);
        case "<=":
          return lte(a, b, loose);
        default:
          throw new TypeError(`Invalid operator: ${op}`);
      }
    };
    module2.exports = cmp;
  }
});

// node_modules/semver/functions/coerce.js
var require_coerce = __commonJS({
  "node_modules/semver/functions/coerce.js"(exports2, module2) {
    "use strict";
    var SemVer = require_semver();
    var parse = require_parse();
    var { safeRe: re, t } = require_re();
    var coerce = (version, options) => {
      if (version instanceof SemVer) {
        return version;
      }
      if (typeof version === "number") {
        version = String(version);
      }
      if (typeof version !== "string") {
        return null;
      }
      options = options || {};
      let match = null;
      if (!options.rtl) {
        match = version.match(options.includePrerelease ? re[t.COERCEFULL] : re[t.COERCE]);
      } else {
        const coerceRtlRegex = options.includePrerelease ? re[t.COERCERTLFULL] : re[t.COERCERTL];
        let next;
        while ((next = coerceRtlRegex.exec(version)) && (!match || match.index + match[0].length !== version.length)) {
          if (!match || next.index + next[0].length !== match.index + match[0].length) {
            match = next;
          }
          coerceRtlRegex.lastIndex = next.index + next[1].length + next[2].length;
        }
        coerceRtlRegex.lastIndex = -1;
      }
      if (match === null) {
        return null;
      }
      const major = match[2];
      const minor = match[3] || "0";
      const patch = match[4] || "0";
      const prerelease = options.includePrerelease && match[5] ? `-${match[5]}` : "";
      const build = options.includePrerelease && match[6] ? `+${match[6]}` : "";
      return parse(`${major}.${minor}.${patch}${prerelease}${build}`, options);
    };
    module2.exports = coerce;
  }
});

// node_modules/semver/internal/lrucache.js
var require_lrucache = __commonJS({
  "node_modules/semver/internal/lrucache.js"(exports2, module2) {
    "use strict";
    var LRUCache = class {
      constructor() {
        this.max = 1e3;
        this.map = /* @__PURE__ */ new Map();
      }
      get(key) {
        const value = this.map.get(key);
        if (value === void 0) {
          return void 0;
        } else {
          this.map.delete(key);
          this.map.set(key, value);
          return value;
        }
      }
      delete(key) {
        return this.map.delete(key);
      }
      set(key, value) {
        const deleted = this.delete(key);
        if (!deleted && value !== void 0) {
          if (this.map.size >= this.max) {
            const firstKey = this.map.keys().next().value;
            this.delete(firstKey);
          }
          this.map.set(key, value);
        }
        return this;
      }
    };
    module2.exports = LRUCache;
  }
});

// node_modules/semver/classes/range.js
var require_range = __commonJS({
  "node_modules/semver/classes/range.js"(exports2, module2) {
    "use strict";
    var SPACE_CHARACTERS = /\s+/g;
    var Range = class _Range {
      constructor(range, options) {
        options = parseOptions(options);
        if (range instanceof _Range) {
          if (range.loose === !!options.loose && range.includePrerelease === !!options.includePrerelease) {
            return range;
          } else {
            return new _Range(range.raw, options);
          }
        }
        if (range instanceof Comparator) {
          this.raw = range.value;
          this.set = [[range]];
          this.formatted = void 0;
          return this;
        }
        this.options = options;
        this.loose = !!options.loose;
        this.includePrerelease = !!options.includePrerelease;
        this.raw = range.trim().replace(SPACE_CHARACTERS, " ");
        this.set = this.raw.split("||").map((r) => this.parseRange(r.trim())).filter((c) => c.length);
        if (!this.set.length) {
          throw new TypeError(`Invalid SemVer Range: ${this.raw}`);
        }
        if (this.set.length > 1) {
          const first = this.set[0];
          this.set = this.set.filter((c) => !isNullSet(c[0]));
          if (this.set.length === 0) {
            this.set = [first];
          } else if (this.set.length > 1) {
            for (const c of this.set) {
              if (c.length === 1 && isAny(c[0])) {
                this.set = [c];
                break;
              }
            }
          }
        }
        this.formatted = void 0;
      }
      get range() {
        if (this.formatted === void 0) {
          this.formatted = "";
          for (let i = 0; i < this.set.length; i++) {
            if (i > 0) {
              this.formatted += "||";
            }
            const comps = this.set[i];
            for (let k = 0; k < comps.length; k++) {
              if (k > 0) {
                this.formatted += " ";
              }
              this.formatted += comps[k].toString().trim();
            }
          }
        }
        return this.formatted;
      }
      format() {
        return this.range;
      }
      toString() {
        return this.range;
      }
      parseRange(range) {
        const memoOpts = (this.options.includePrerelease && FLAG_INCLUDE_PRERELEASE) | (this.options.loose && FLAG_LOOSE);
        const memoKey = memoOpts + ":" + range;
        const cached = cache.get(memoKey);
        if (cached) {
          return cached;
        }
        const loose = this.options.loose;
        const hr = loose ? re[t.HYPHENRANGELOOSE] : re[t.HYPHENRANGE];
        range = range.replace(hr, hyphenReplace(this.options.includePrerelease));
        debug("hyphen replace", range);
        range = range.replace(re[t.COMPARATORTRIM], comparatorTrimReplace);
        debug("comparator trim", range);
        range = range.replace(re[t.TILDETRIM], tildeTrimReplace);
        debug("tilde trim", range);
        range = range.replace(re[t.CARETTRIM], caretTrimReplace);
        debug("caret trim", range);
        let rangeList = range.split(" ").map((comp) => parseComparator(comp, this.options)).join(" ").split(/\s+/).map((comp) => replaceGTE0(comp, this.options));
        if (loose) {
          rangeList = rangeList.filter((comp) => {
            debug("loose invalid filter", comp, this.options);
            return !!comp.match(re[t.COMPARATORLOOSE]);
          });
        }
        debug("range list", rangeList);
        const rangeMap = /* @__PURE__ */ new Map();
        const comparators = rangeList.map((comp) => new Comparator(comp, this.options));
        for (const comp of comparators) {
          if (isNullSet(comp)) {
            return [comp];
          }
          rangeMap.set(comp.value, comp);
        }
        if (rangeMap.size > 1 && rangeMap.has("")) {
          rangeMap.delete("");
        }
        const result = [...rangeMap.values()];
        cache.set(memoKey, result);
        return result;
      }
      intersects(range, options) {
        if (!(range instanceof _Range)) {
          throw new TypeError("a Range is required");
        }
        return this.set.some((thisComparators) => {
          return isSatisfiable(thisComparators, options) && range.set.some((rangeComparators) => {
            return isSatisfiable(rangeComparators, options) && thisComparators.every((thisComparator) => {
              return rangeComparators.every((rangeComparator) => {
                return thisComparator.intersects(rangeComparator, options);
              });
            });
          });
        });
      }
      // if ANY of the sets match ALL of its comparators, then pass
      test(version) {
        if (!version) {
          return false;
        }
        if (typeof version === "string") {
          try {
            version = new SemVer(version, this.options);
          } catch (er) {
            return false;
          }
        }
        for (let i = 0; i < this.set.length; i++) {
          if (testSet(this.set[i], version, this.options)) {
            return true;
          }
        }
        return false;
      }
    };
    module2.exports = Range;
    var LRU = require_lrucache();
    var cache = new LRU();
    var parseOptions = require_parse_options();
    var Comparator = require_comparator();
    var debug = require_debug();
    var SemVer = require_semver();
    var {
      safeRe: re,
      t,
      comparatorTrimReplace,
      tildeTrimReplace,
      caretTrimReplace
    } = require_re();
    var { FLAG_INCLUDE_PRERELEASE, FLAG_LOOSE } = require_constants();
    var isNullSet = (c) => c.value === "<0.0.0-0";
    var isAny = (c) => c.value === "";
    var isSatisfiable = (comparators, options) => {
      let result = true;
      const remainingComparators = comparators.slice();
      let testComparator = remainingComparators.pop();
      while (result && remainingComparators.length) {
        result = remainingComparators.every((otherComparator) => {
          return testComparator.intersects(otherComparator, options);
        });
        testComparator = remainingComparators.pop();
      }
      return result;
    };
    var parseComparator = (comp, options) => {
      debug("comp", comp, options);
      comp = replaceCarets(comp, options);
      debug("caret", comp);
      comp = replaceTildes(comp, options);
      debug("tildes", comp);
      comp = replaceXRanges(comp, options);
      debug("xrange", comp);
      comp = replaceStars(comp, options);
      debug("stars", comp);
      return comp;
    };
    var isX = (id) => !id || id.toLowerCase() === "x" || id === "*";
    var replaceTildes = (comp, options) => {
      return comp.trim().split(/\s+/).map((c) => replaceTilde(c, options)).join(" ");
    };
    var replaceTilde = (comp, options) => {
      const r = options.loose ? re[t.TILDELOOSE] : re[t.TILDE];
      return comp.replace(r, (_, M, m, p, pr) => {
        debug("tilde", comp, _, M, m, p, pr);
        let ret;
        if (isX(M)) {
          ret = "";
        } else if (isX(m)) {
          ret = `>=${M}.0.0 <${+M + 1}.0.0-0`;
        } else if (isX(p)) {
          ret = `>=${M}.${m}.0 <${M}.${+m + 1}.0-0`;
        } else if (pr) {
          debug("replaceTilde pr", pr);
          ret = `>=${M}.${m}.${p}-${pr} <${M}.${+m + 1}.0-0`;
        } else {
          ret = `>=${M}.${m}.${p} <${M}.${+m + 1}.0-0`;
        }
        debug("tilde return", ret);
        return ret;
      });
    };
    var replaceCarets = (comp, options) => {
      return comp.trim().split(/\s+/).map((c) => replaceCaret(c, options)).join(" ");
    };
    var replaceCaret = (comp, options) => {
      debug("caret", comp, options);
      const r = options.loose ? re[t.CARETLOOSE] : re[t.CARET];
      const z = options.includePrerelease ? "-0" : "";
      return comp.replace(r, (_, M, m, p, pr) => {
        debug("caret", comp, _, M, m, p, pr);
        let ret;
        if (isX(M)) {
          ret = "";
        } else if (isX(m)) {
          ret = `>=${M}.0.0${z} <${+M + 1}.0.0-0`;
        } else if (isX(p)) {
          if (M === "0") {
            ret = `>=${M}.${m}.0${z} <${M}.${+m + 1}.0-0`;
          } else {
            ret = `>=${M}.${m}.0${z} <${+M + 1}.0.0-0`;
          }
        } else if (pr) {
          debug("replaceCaret pr", pr);
          if (M === "0") {
            if (m === "0") {
              ret = `>=${M}.${m}.${p}-${pr} <${M}.${m}.${+p + 1}-0`;
            } else {
              ret = `>=${M}.${m}.${p}-${pr} <${M}.${+m + 1}.0-0`;
            }
          } else {
            ret = `>=${M}.${m}.${p}-${pr} <${+M + 1}.0.0-0`;
          }
        } else {
          debug("no pr");
          if (M === "0") {
            if (m === "0") {
              ret = `>=${M}.${m}.${p}${z} <${M}.${m}.${+p + 1}-0`;
            } else {
              ret = `>=${M}.${m}.${p}${z} <${M}.${+m + 1}.0-0`;
            }
          } else {
            ret = `>=${M}.${m}.${p} <${+M + 1}.0.0-0`;
          }
        }
        debug("caret return", ret);
        return ret;
      });
    };
    var replaceXRanges = (comp, options) => {
      debug("replaceXRanges", comp, options);
      return comp.split(/\s+/).map((c) => replaceXRange(c, options)).join(" ");
    };
    var replaceXRange = (comp, options) => {
      comp = comp.trim();
      const r = options.loose ? re[t.XRANGELOOSE] : re[t.XRANGE];
      return comp.replace(r, (ret, gtlt, M, m, p, pr) => {
        debug("xRange", comp, ret, gtlt, M, m, p, pr);
        const xM = isX(M);
        const xm = xM || isX(m);
        const xp = xm || isX(p);
        const anyX = xp;
        if (gtlt === "=" && anyX) {
          gtlt = "";
        }
        pr = options.includePrerelease ? "-0" : "";
        if (xM) {
          if (gtlt === ">" || gtlt === "<") {
            ret = "<0.0.0-0";
          } else {
            ret = "*";
          }
        } else if (gtlt && anyX) {
          if (xm) {
            m = 0;
          }
          p = 0;
          if (gtlt === ">") {
            gtlt = ">=";
            if (xm) {
              M = +M + 1;
              m = 0;
              p = 0;
            } else {
              m = +m + 1;
              p = 0;
            }
          } else if (gtlt === "<=") {
            gtlt = "<";
            if (xm) {
              M = +M + 1;
            } else {
              m = +m + 1;
            }
          }
          if (gtlt === "<") {
            pr = "-0";
          }
          ret = `${gtlt + M}.${m}.${p}${pr}`;
        } else if (xm) {
          ret = `>=${M}.0.0${pr} <${+M + 1}.0.0-0`;
        } else if (xp) {
          ret = `>=${M}.${m}.0${pr} <${M}.${+m + 1}.0-0`;
        }
        debug("xRange return", ret);
        return ret;
      });
    };
    var replaceStars = (comp, options) => {
      debug("replaceStars", comp, options);
      return comp.trim().replace(re[t.STAR], "");
    };
    var replaceGTE0 = (comp, options) => {
      debug("replaceGTE0", comp, options);
      return comp.trim().replace(re[options.includePrerelease ? t.GTE0PRE : t.GTE0], "");
    };
    var hyphenReplace = (incPr) => ($0, from, fM, fm, fp, fpr, fb, to, tM, tm, tp, tpr) => {
      if (isX(fM)) {
        from = "";
      } else if (isX(fm)) {
        from = `>=${fM}.0.0${incPr ? "-0" : ""}`;
      } else if (isX(fp)) {
        from = `>=${fM}.${fm}.0${incPr ? "-0" : ""}`;
      } else if (fpr) {
        from = `>=${from}`;
      } else {
        from = `>=${from}${incPr ? "-0" : ""}`;
      }
      if (isX(tM)) {
        to = "";
      } else if (isX(tm)) {
        to = `<${+tM + 1}.0.0-0`;
      } else if (isX(tp)) {
        to = `<${tM}.${+tm + 1}.0-0`;
      } else if (tpr) {
        to = `<=${tM}.${tm}.${tp}-${tpr}`;
      } else if (incPr) {
        to = `<${tM}.${tm}.${+tp + 1}-0`;
      } else {
        to = `<=${to}`;
      }
      return `${from} ${to}`.trim();
    };
    var testSet = (set, version, options) => {
      for (let i = 0; i < set.length; i++) {
        if (!set[i].test(version)) {
          return false;
        }
      }
      if (version.prerelease.length && !options.includePrerelease) {
        for (let i = 0; i < set.length; i++) {
          debug(set[i].semver);
          if (set[i].semver === Comparator.ANY) {
            continue;
          }
          if (set[i].semver.prerelease.length > 0) {
            const allowed = set[i].semver;
            if (allowed.major === version.major && allowed.minor === version.minor && allowed.patch === version.patch) {
              return true;
            }
          }
        }
        return false;
      }
      return true;
    };
  }
});

// node_modules/semver/classes/comparator.js
var require_comparator = __commonJS({
  "node_modules/semver/classes/comparator.js"(exports2, module2) {
    "use strict";
    var ANY = Symbol("SemVer ANY");
    var Comparator = class _Comparator {
      static get ANY() {
        return ANY;
      }
      constructor(comp, options) {
        options = parseOptions(options);
        if (comp instanceof _Comparator) {
          if (comp.loose === !!options.loose) {
            return comp;
          } else {
            comp = comp.value;
          }
        }
        comp = comp.trim().split(/\s+/).join(" ");
        debug("comparator", comp, options);
        this.options = options;
        this.loose = !!options.loose;
        this.parse(comp);
        if (this.semver === ANY) {
          this.value = "";
        } else {
          this.value = this.operator + this.semver.version;
        }
        debug("comp", this);
      }
      parse(comp) {
        const r = this.options.loose ? re[t.COMPARATORLOOSE] : re[t.COMPARATOR];
        const m = comp.match(r);
        if (!m) {
          throw new TypeError(`Invalid comparator: ${comp}`);
        }
        this.operator = m[1] !== void 0 ? m[1] : "";
        if (this.operator === "=") {
          this.operator = "";
        }
        if (!m[2]) {
          this.semver = ANY;
        } else {
          this.semver = new SemVer(m[2], this.options.loose);
        }
      }
      toString() {
        return this.value;
      }
      test(version) {
        debug("Comparator.test", version, this.options.loose);
        if (this.semver === ANY || version === ANY) {
          return true;
        }
        if (typeof version === "string") {
          try {
            version = new SemVer(version, this.options);
          } catch (er) {
            return false;
          }
        }
        return cmp(version, this.operator, this.semver, this.options);
      }
      intersects(comp, options) {
        if (!(comp instanceof _Comparator)) {
          throw new TypeError("a Comparator is required");
        }
        if (this.operator === "") {
          if (this.value === "") {
            return true;
          }
          return new Range(comp.value, options).test(this.value);
        } else if (comp.operator === "") {
          if (comp.value === "") {
            return true;
          }
          return new Range(this.value, options).test(comp.semver);
        }
        options = parseOptions(options);
        if (options.includePrerelease && (this.value === "<0.0.0-0" || comp.value === "<0.0.0-0")) {
          return false;
        }
        if (!options.includePrerelease && (this.value.startsWith("<0.0.0") || comp.value.startsWith("<0.0.0"))) {
          return false;
        }
        if (this.operator.startsWith(">") && comp.operator.startsWith(">")) {
          return true;
        }
        if (this.operator.startsWith("<") && comp.operator.startsWith("<")) {
          return true;
        }
        if (this.semver.version === comp.semver.version && this.operator.includes("=") && comp.operator.includes("=")) {
          return true;
        }
        if (cmp(this.semver, "<", comp.semver, options) && this.operator.startsWith(">") && comp.operator.startsWith("<")) {
          return true;
        }
        if (cmp(this.semver, ">", comp.semver, options) && this.operator.startsWith("<") && comp.operator.startsWith(">")) {
          return true;
        }
        return false;
      }
    };
    module2.exports = Comparator;
    var parseOptions = require_parse_options();
    var { safeRe: re, t } = require_re();
    var cmp = require_cmp();
    var debug = require_debug();
    var SemVer = require_semver();
    var Range = require_range();
  }
});

// node_modules/semver/functions/satisfies.js
var require_satisfies = __commonJS({
  "node_modules/semver/functions/satisfies.js"(exports2, module2) {
    "use strict";
    var Range = require_range();
    var satisfies = (version, range, options) => {
      try {
        range = new Range(range, options);
      } catch (er) {
        return false;
      }
      return range.test(version);
    };
    module2.exports = satisfies;
  }
});

// node_modules/semver/ranges/to-comparators.js
var require_to_comparators = __commonJS({
  "node_modules/semver/ranges/to-comparators.js"(exports2, module2) {
    "use strict";
    var Range = require_range();
    var toComparators = (range, options) => new Range(range, options).set.map((comp) => comp.map((c) => c.value).join(" ").trim().split(" "));
    module2.exports = toComparators;
  }
});

// node_modules/semver/ranges/max-satisfying.js
var require_max_satisfying = __commonJS({
  "node_modules/semver/ranges/max-satisfying.js"(exports2, module2) {
    "use strict";
    var SemVer = require_semver();
    var Range = require_range();
    var maxSatisfying = (versions, range, options) => {
      let max = null;
      let maxSV = null;
      let rangeObj = null;
      try {
        rangeObj = new Range(range, options);
      } catch (er) {
        return null;
      }
      versions.forEach((v) => {
        if (rangeObj.test(v)) {
          if (!max || maxSV.compare(v) === -1) {
            max = v;
            maxSV = new SemVer(max, options);
          }
        }
      });
      return max;
    };
    module2.exports = maxSatisfying;
  }
});

// node_modules/semver/ranges/min-satisfying.js
var require_min_satisfying = __commonJS({
  "node_modules/semver/ranges/min-satisfying.js"(exports2, module2) {
    "use strict";
    var SemVer = require_semver();
    var Range = require_range();
    var minSatisfying = (versions, range, options) => {
      let min = null;
      let minSV = null;
      let rangeObj = null;
      try {
        rangeObj = new Range(range, options);
      } catch (er) {
        return null;
      }
      versions.forEach((v) => {
        if (rangeObj.test(v)) {
          if (!min || minSV.compare(v) === 1) {
            min = v;
            minSV = new SemVer(min, options);
          }
        }
      });
      return min;
    };
    module2.exports = minSatisfying;
  }
});

// node_modules/semver/ranges/min-version.js
var require_min_version = __commonJS({
  "node_modules/semver/ranges/min-version.js"(exports2, module2) {
    "use strict";
    var SemVer = require_semver();
    var Range = require_range();
    var gt = require_gt();
    var minVersion = (range, loose) => {
      range = new Range(range, loose);
      let minver = new SemVer("0.0.0");
      if (range.test(minver)) {
        return minver;
      }
      minver = new SemVer("0.0.0-0");
      if (range.test(minver)) {
        return minver;
      }
      minver = null;
      for (let i = 0; i < range.set.length; ++i) {
        const comparators = range.set[i];
        let setMin = null;
        comparators.forEach((comparator) => {
          const compver = new SemVer(comparator.semver.version);
          switch (comparator.operator) {
            case ">":
              if (compver.prerelease.length === 0) {
                compver.patch++;
              } else {
                compver.prerelease.push(0);
              }
              compver.raw = compver.format();
            /* fallthrough */
            case "":
            case ">=":
              if (!setMin || gt(compver, setMin)) {
                setMin = compver;
              }
              break;
            case "<":
            case "<=":
              break;
            /* istanbul ignore next */
            default:
              throw new Error(`Unexpected operation: ${comparator.operator}`);
          }
        });
        if (setMin && (!minver || gt(minver, setMin))) {
          minver = setMin;
        }
      }
      if (minver && range.test(minver)) {
        return minver;
      }
      return null;
    };
    module2.exports = minVersion;
  }
});

// node_modules/semver/ranges/valid.js
var require_valid2 = __commonJS({
  "node_modules/semver/ranges/valid.js"(exports2, module2) {
    "use strict";
    var Range = require_range();
    var validRange = (range, options) => {
      try {
        return new Range(range, options).range || "*";
      } catch (er) {
        return null;
      }
    };
    module2.exports = validRange;
  }
});

// node_modules/semver/ranges/outside.js
var require_outside = __commonJS({
  "node_modules/semver/ranges/outside.js"(exports2, module2) {
    "use strict";
    var SemVer = require_semver();
    var Comparator = require_comparator();
    var { ANY } = Comparator;
    var Range = require_range();
    var satisfies = require_satisfies();
    var gt = require_gt();
    var lt = require_lt();
    var lte = require_lte();
    var gte = require_gte();
    var outside = (version, range, hilo, options) => {
      version = new SemVer(version, options);
      range = new Range(range, options);
      let gtfn, ltefn, ltfn, comp, ecomp;
      switch (hilo) {
        case ">":
          gtfn = gt;
          ltefn = lte;
          ltfn = lt;
          comp = ">";
          ecomp = ">=";
          break;
        case "<":
          gtfn = lt;
          ltefn = gte;
          ltfn = gt;
          comp = "<";
          ecomp = "<=";
          break;
        default:
          throw new TypeError('Must provide a hilo val of "<" or ">"');
      }
      if (satisfies(version, range, options)) {
        return false;
      }
      for (let i = 0; i < range.set.length; ++i) {
        const comparators = range.set[i];
        let high = null;
        let low = null;
        comparators.forEach((comparator) => {
          if (comparator.semver === ANY) {
            comparator = new Comparator(">=0.0.0");
          }
          high = high || comparator;
          low = low || comparator;
          if (gtfn(comparator.semver, high.semver, options)) {
            high = comparator;
          } else if (ltfn(comparator.semver, low.semver, options)) {
            low = comparator;
          }
        });
        if (high.operator === comp || high.operator === ecomp) {
          return false;
        }
        if ((!low.operator || low.operator === comp) && ltefn(version, low.semver)) {
          return false;
        } else if (low.operator === ecomp && ltfn(version, low.semver)) {
          return false;
        }
      }
      return true;
    };
    module2.exports = outside;
  }
});

// node_modules/semver/ranges/gtr.js
var require_gtr = __commonJS({
  "node_modules/semver/ranges/gtr.js"(exports2, module2) {
    "use strict";
    var outside = require_outside();
    var gtr = (version, range, options) => outside(version, range, ">", options);
    module2.exports = gtr;
  }
});

// node_modules/semver/ranges/ltr.js
var require_ltr = __commonJS({
  "node_modules/semver/ranges/ltr.js"(exports2, module2) {
    "use strict";
    var outside = require_outside();
    var ltr = (version, range, options) => outside(version, range, "<", options);
    module2.exports = ltr;
  }
});

// node_modules/semver/ranges/intersects.js
var require_intersects = __commonJS({
  "node_modules/semver/ranges/intersects.js"(exports2, module2) {
    "use strict";
    var Range = require_range();
    var intersects = (r1, r2, options) => {
      r1 = new Range(r1, options);
      r2 = new Range(r2, options);
      return r1.intersects(r2, options);
    };
    module2.exports = intersects;
  }
});

// node_modules/semver/ranges/simplify.js
var require_simplify = __commonJS({
  "node_modules/semver/ranges/simplify.js"(exports2, module2) {
    "use strict";
    var satisfies = require_satisfies();
    var compare = require_compare();
    module2.exports = (versions, range, options) => {
      const set = [];
      let first = null;
      let prev = null;
      const v = versions.sort((a, b) => compare(a, b, options));
      for (const version of v) {
        const included = satisfies(version, range, options);
        if (included) {
          prev = version;
          if (!first) {
            first = version;
          }
        } else {
          if (prev) {
            set.push([first, prev]);
          }
          prev = null;
          first = null;
        }
      }
      if (first) {
        set.push([first, null]);
      }
      const ranges = [];
      for (const [min, max] of set) {
        if (min === max) {
          ranges.push(min);
        } else if (!max && min === v[0]) {
          ranges.push("*");
        } else if (!max) {
          ranges.push(`>=${min}`);
        } else if (min === v[0]) {
          ranges.push(`<=${max}`);
        } else {
          ranges.push(`${min} - ${max}`);
        }
      }
      const simplified = ranges.join(" || ");
      const original = typeof range.raw === "string" ? range.raw : String(range);
      return simplified.length < original.length ? simplified : range;
    };
  }
});

// node_modules/semver/ranges/subset.js
var require_subset = __commonJS({
  "node_modules/semver/ranges/subset.js"(exports2, module2) {
    "use strict";
    var Range = require_range();
    var Comparator = require_comparator();
    var { ANY } = Comparator;
    var satisfies = require_satisfies();
    var compare = require_compare();
    var subset = (sub, dom, options = {}) => {
      if (sub === dom) {
        return true;
      }
      sub = new Range(sub, options);
      dom = new Range(dom, options);
      let sawNonNull = false;
      OUTER: for (const simpleSub of sub.set) {
        for (const simpleDom of dom.set) {
          const isSub = simpleSubset(simpleSub, simpleDom, options);
          sawNonNull = sawNonNull || isSub !== null;
          if (isSub) {
            continue OUTER;
          }
        }
        if (sawNonNull) {
          return false;
        }
      }
      return true;
    };
    var minimumVersionWithPreRelease = [new Comparator(">=0.0.0-0")];
    var minimumVersion = [new Comparator(">=0.0.0")];
    var simpleSubset = (sub, dom, options) => {
      if (sub === dom) {
        return true;
      }
      if (sub.length === 1 && sub[0].semver === ANY) {
        if (dom.length === 1 && dom[0].semver === ANY) {
          return true;
        } else if (options.includePrerelease) {
          sub = minimumVersionWithPreRelease;
        } else {
          sub = minimumVersion;
        }
      }
      if (dom.length === 1 && dom[0].semver === ANY) {
        if (options.includePrerelease) {
          return true;
        } else {
          dom = minimumVersion;
        }
      }
      const eqSet = /* @__PURE__ */ new Set();
      let gt, lt;
      for (const c of sub) {
        if (c.operator === ">" || c.operator === ">=") {
          gt = higherGT(gt, c, options);
        } else if (c.operator === "<" || c.operator === "<=") {
          lt = lowerLT(lt, c, options);
        } else {
          eqSet.add(c.semver);
        }
      }
      if (eqSet.size > 1) {
        return null;
      }
      let gtltComp;
      if (gt && lt) {
        gtltComp = compare(gt.semver, lt.semver, options);
        if (gtltComp > 0) {
          return null;
        } else if (gtltComp === 0 && (gt.operator !== ">=" || lt.operator !== "<=")) {
          return null;
        }
      }
      for (const eq of eqSet) {
        if (gt && !satisfies(eq, String(gt), options)) {
          return null;
        }
        if (lt && !satisfies(eq, String(lt), options)) {
          return null;
        }
        for (const c of dom) {
          if (!satisfies(eq, String(c), options)) {
            return false;
          }
        }
        return true;
      }
      let higher, lower;
      let hasDomLT, hasDomGT;
      let needDomLTPre = lt && !options.includePrerelease && lt.semver.prerelease.length ? lt.semver : false;
      let needDomGTPre = gt && !options.includePrerelease && gt.semver.prerelease.length ? gt.semver : false;
      if (needDomLTPre && needDomLTPre.prerelease.length === 1 && lt.operator === "<" && needDomLTPre.prerelease[0] === 0) {
        needDomLTPre = false;
      }
      for (const c of dom) {
        hasDomGT = hasDomGT || c.operator === ">" || c.operator === ">=";
        hasDomLT = hasDomLT || c.operator === "<" || c.operator === "<=";
        if (gt) {
          if (needDomGTPre) {
            if (c.semver.prerelease && c.semver.prerelease.length && c.semver.major === needDomGTPre.major && c.semver.minor === needDomGTPre.minor && c.semver.patch === needDomGTPre.patch) {
              needDomGTPre = false;
            }
          }
          if (c.operator === ">" || c.operator === ">=") {
            higher = higherGT(gt, c, options);
            if (higher === c && higher !== gt) {
              return false;
            }
          } else if (gt.operator === ">=" && !satisfies(gt.semver, String(c), options)) {
            return false;
          }
        }
        if (lt) {
          if (needDomLTPre) {
            if (c.semver.prerelease && c.semver.prerelease.length && c.semver.major === needDomLTPre.major && c.semver.minor === needDomLTPre.minor && c.semver.patch === needDomLTPre.patch) {
              needDomLTPre = false;
            }
          }
          if (c.operator === "<" || c.operator === "<=") {
            lower = lowerLT(lt, c, options);
            if (lower === c && lower !== lt) {
              return false;
            }
          } else if (lt.operator === "<=" && !satisfies(lt.semver, String(c), options)) {
            return false;
          }
        }
        if (!c.operator && (lt || gt) && gtltComp !== 0) {
          return false;
        }
      }
      if (gt && hasDomLT && !lt && gtltComp !== 0) {
        return false;
      }
      if (lt && hasDomGT && !gt && gtltComp !== 0) {
        return false;
      }
      if (needDomGTPre || needDomLTPre) {
        return false;
      }
      return true;
    };
    var higherGT = (a, b, options) => {
      if (!a) {
        return b;
      }
      const comp = compare(a.semver, b.semver, options);
      return comp > 0 ? a : comp < 0 ? b : b.operator === ">" && a.operator === ">=" ? b : a;
    };
    var lowerLT = (a, b, options) => {
      if (!a) {
        return b;
      }
      const comp = compare(a.semver, b.semver, options);
      return comp < 0 ? a : comp > 0 ? b : b.operator === "<" && a.operator === "<=" ? b : a;
    };
    module2.exports = subset;
  }
});

// node_modules/semver/index.js
var require_semver2 = __commonJS({
  "node_modules/semver/index.js"(exports2, module2) {
    "use strict";
    var internalRe = require_re();
    var constants = require_constants();
    var SemVer = require_semver();
    var identifiers = require_identifiers();
    var parse = require_parse();
    var valid = require_valid();
    var clean = require_clean();
    var inc = require_inc();
    var diff = require_diff();
    var major = require_major();
    var minor = require_minor();
    var patch = require_patch();
    var prerelease = require_prerelease();
    var compare = require_compare();
    var rcompare = require_rcompare();
    var compareLoose = require_compare_loose();
    var compareBuild = require_compare_build();
    var sort = require_sort();
    var rsort = require_rsort();
    var gt = require_gt();
    var lt = require_lt();
    var eq = require_eq();
    var neq = require_neq();
    var gte = require_gte();
    var lte = require_lte();
    var cmp = require_cmp();
    var coerce = require_coerce();
    var Comparator = require_comparator();
    var Range = require_range();
    var satisfies = require_satisfies();
    var toComparators = require_to_comparators();
    var maxSatisfying = require_max_satisfying();
    var minSatisfying = require_min_satisfying();
    var minVersion = require_min_version();
    var validRange = require_valid2();
    var outside = require_outside();
    var gtr = require_gtr();
    var ltr = require_ltr();
    var intersects = require_intersects();
    var simplifyRange = require_simplify();
    var subset = require_subset();
    module2.exports = {
      parse,
      valid,
      clean,
      inc,
      diff,
      major,
      minor,
      patch,
      prerelease,
      compare,
      rcompare,
      compareLoose,
      compareBuild,
      sort,
      rsort,
      gt,
      lt,
      eq,
      neq,
      gte,
      lte,
      cmp,
      coerce,
      Comparator,
      Range,
      satisfies,
      toComparators,
      maxSatisfying,
      minSatisfying,
      minVersion,
      validRange,
      outside,
      gtr,
      ltr,
      intersects,
      simplifyRange,
      subset,
      SemVer,
      re: internalRe.re,
      src: internalRe.src,
      tokens: internalRe.t,
      SEMVER_SPEC_VERSION: constants.SEMVER_SPEC_VERSION,
      RELEASE_TYPES: constants.RELEASE_TYPES,
      compareIdentifiers: identifiers.compareIdentifiers,
      rcompareIdentifiers: identifiers.rcompareIdentifiers
    };
  }
});

// node_modules/vscode-languageclient/lib/common/utils/is.js
var require_is = __commonJS({
  "node_modules/vscode-languageclient/lib/common/utils/is.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.asPromise = exports2.thenable = exports2.typedArray = exports2.stringArray = exports2.array = exports2.func = exports2.error = exports2.number = exports2.string = exports2.boolean = void 0;
    function boolean(value) {
      return value === true || value === false;
    }
    exports2.boolean = boolean;
    function string(value) {
      return typeof value === "string" || value instanceof String;
    }
    exports2.string = string;
    function number(value) {
      return typeof value === "number" || value instanceof Number;
    }
    exports2.number = number;
    function error(value) {
      return value instanceof Error;
    }
    exports2.error = error;
    function func(value) {
      return typeof value === "function";
    }
    exports2.func = func;
    function array(value) {
      return Array.isArray(value);
    }
    exports2.array = array;
    function stringArray(value) {
      return array(value) && value.every((elem) => string(elem));
    }
    exports2.stringArray = stringArray;
    function typedArray(value, check) {
      return Array.isArray(value) && value.every(check);
    }
    exports2.typedArray = typedArray;
    function thenable(value) {
      return value && func(value.then);
    }
    exports2.thenable = thenable;
    function asPromise(value) {
      if (value instanceof Promise) {
        return value;
      } else if (thenable(value)) {
        return new Promise((resolve, reject) => {
          value.then((resolved) => resolve(resolved), (error2) => reject(error2));
        });
      } else {
        return Promise.resolve(value);
      }
    }
    exports2.asPromise = asPromise;
  }
});

// node_modules/vscode-jsonrpc/lib/common/ral.js
var require_ral = __commonJS({
  "node_modules/vscode-jsonrpc/lib/common/ral.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    var _ral;
    function RAL() {
      if (_ral === void 0) {
        throw new Error(`No runtime abstraction layer installed`);
      }
      return _ral;
    }
    (function(RAL2) {
      function install(ral) {
        if (ral === void 0) {
          throw new Error(`No runtime abstraction layer provided`);
        }
        _ral = ral;
      }
      RAL2.install = install;
    })(RAL || (RAL = {}));
    exports2.default = RAL;
  }
});

// node_modules/vscode-jsonrpc/lib/common/disposable.js
var require_disposable = __commonJS({
  "node_modules/vscode-jsonrpc/lib/common/disposable.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.Disposable = void 0;
    var Disposable2;
    (function(Disposable3) {
      function create(func) {
        return {
          dispose: func
        };
      }
      Disposable3.create = create;
    })(Disposable2 = exports2.Disposable || (exports2.Disposable = {}));
  }
});

// node_modules/vscode-jsonrpc/lib/common/messageBuffer.js
var require_messageBuffer = __commonJS({
  "node_modules/vscode-jsonrpc/lib/common/messageBuffer.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.AbstractMessageBuffer = void 0;
    var CR = 13;
    var LF = 10;
    var CRLF = "\r\n";
    var AbstractMessageBuffer = class {
      constructor(encoding = "utf-8") {
        this._encoding = encoding;
        this._chunks = [];
        this._totalLength = 0;
      }
      get encoding() {
        return this._encoding;
      }
      append(chunk) {
        const toAppend = typeof chunk === "string" ? this.fromString(chunk, this._encoding) : chunk;
        this._chunks.push(toAppend);
        this._totalLength += toAppend.byteLength;
      }
      tryReadHeaders() {
        if (this._chunks.length === 0) {
          return void 0;
        }
        let state = 0;
        let chunkIndex = 0;
        let offset = 0;
        let chunkBytesRead = 0;
        row: while (chunkIndex < this._chunks.length) {
          const chunk = this._chunks[chunkIndex];
          offset = 0;
          column: while (offset < chunk.length) {
            const value = chunk[offset];
            switch (value) {
              case CR:
                switch (state) {
                  case 0:
                    state = 1;
                    break;
                  case 2:
                    state = 3;
                    break;
                  default:
                    state = 0;
                }
                break;
              case LF:
                switch (state) {
                  case 1:
                    state = 2;
                    break;
                  case 3:
                    state = 4;
                    offset++;
                    break row;
                  default:
                    state = 0;
                }
                break;
              default:
                state = 0;
            }
            offset++;
          }
          chunkBytesRead += chunk.byteLength;
          chunkIndex++;
        }
        if (state !== 4) {
          return void 0;
        }
        const buffer = this._read(chunkBytesRead + offset);
        const result = /* @__PURE__ */ new Map();
        const headers = this.toString(buffer, "ascii").split(CRLF);
        if (headers.length < 2) {
          return result;
        }
        for (let i = 0; i < headers.length - 2; i++) {
          const header = headers[i];
          const index = header.indexOf(":");
          if (index === -1) {
            throw new Error("Message header must separate key and value using :");
          }
          const key = header.substr(0, index);
          const value = header.substr(index + 1).trim();
          result.set(key, value);
        }
        return result;
      }
      tryReadBody(length) {
        if (this._totalLength < length) {
          return void 0;
        }
        return this._read(length);
      }
      get numberOfBytes() {
        return this._totalLength;
      }
      _read(byteCount) {
        if (byteCount === 0) {
          return this.emptyBuffer();
        }
        if (byteCount > this._totalLength) {
          throw new Error(`Cannot read so many bytes!`);
        }
        if (this._chunks[0].byteLength === byteCount) {
          const chunk = this._chunks[0];
          this._chunks.shift();
          this._totalLength -= byteCount;
          return this.asNative(chunk);
        }
        if (this._chunks[0].byteLength > byteCount) {
          const chunk = this._chunks[0];
          const result2 = this.asNative(chunk, byteCount);
          this._chunks[0] = chunk.slice(byteCount);
          this._totalLength -= byteCount;
          return result2;
        }
        const result = this.allocNative(byteCount);
        let resultOffset = 0;
        let chunkIndex = 0;
        while (byteCount > 0) {
          const chunk = this._chunks[chunkIndex];
          if (chunk.byteLength > byteCount) {
            const chunkPart = chunk.slice(0, byteCount);
            result.set(chunkPart, resultOffset);
            resultOffset += byteCount;
            this._chunks[chunkIndex] = chunk.slice(byteCount);
            this._totalLength -= byteCount;
            byteCount -= byteCount;
          } else {
            result.set(chunk, resultOffset);
            resultOffset += chunk.byteLength;
            this._chunks.shift();
            this._totalLength -= chunk.byteLength;
            byteCount -= chunk.byteLength;
          }
        }
        return result;
      }
    };
    exports2.AbstractMessageBuffer = AbstractMessageBuffer;
  }
});

// node_modules/vscode-jsonrpc/lib/node/ril.js
var require_ril = __commonJS({
  "node_modules/vscode-jsonrpc/lib/node/ril.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    var ral_1 = require_ral();
    var util_1 = require("util");
    var disposable_1 = require_disposable();
    var messageBuffer_1 = require_messageBuffer();
    var MessageBuffer = class _MessageBuffer extends messageBuffer_1.AbstractMessageBuffer {
      constructor(encoding = "utf-8") {
        super(encoding);
      }
      emptyBuffer() {
        return _MessageBuffer.emptyBuffer;
      }
      fromString(value, encoding) {
        return Buffer.from(value, encoding);
      }
      toString(value, encoding) {
        if (value instanceof Buffer) {
          return value.toString(encoding);
        } else {
          return new util_1.TextDecoder(encoding).decode(value);
        }
      }
      asNative(buffer, length) {
        if (length === void 0) {
          return buffer instanceof Buffer ? buffer : Buffer.from(buffer);
        } else {
          return buffer instanceof Buffer ? buffer.slice(0, length) : Buffer.from(buffer, 0, length);
        }
      }
      allocNative(length) {
        return Buffer.allocUnsafe(length);
      }
    };
    MessageBuffer.emptyBuffer = Buffer.allocUnsafe(0);
    var ReadableStreamWrapper = class {
      constructor(stream) {
        this.stream = stream;
      }
      onClose(listener) {
        this.stream.on("close", listener);
        return disposable_1.Disposable.create(() => this.stream.off("close", listener));
      }
      onError(listener) {
        this.stream.on("error", listener);
        return disposable_1.Disposable.create(() => this.stream.off("error", listener));
      }
      onEnd(listener) {
        this.stream.on("end", listener);
        return disposable_1.Disposable.create(() => this.stream.off("end", listener));
      }
      onData(listener) {
        this.stream.on("data", listener);
        return disposable_1.Disposable.create(() => this.stream.off("data", listener));
      }
    };
    var WritableStreamWrapper = class {
      constructor(stream) {
        this.stream = stream;
      }
      onClose(listener) {
        this.stream.on("close", listener);
        return disposable_1.Disposable.create(() => this.stream.off("close", listener));
      }
      onError(listener) {
        this.stream.on("error", listener);
        return disposable_1.Disposable.create(() => this.stream.off("error", listener));
      }
      onEnd(listener) {
        this.stream.on("end", listener);
        return disposable_1.Disposable.create(() => this.stream.off("end", listener));
      }
      write(data, encoding) {
        return new Promise((resolve, reject) => {
          const callback = (error) => {
            if (error === void 0 || error === null) {
              resolve();
            } else {
              reject(error);
            }
          };
          if (typeof data === "string") {
            this.stream.write(data, encoding, callback);
          } else {
            this.stream.write(data, callback);
          }
        });
      }
      end() {
        this.stream.end();
      }
    };
    var _ril = Object.freeze({
      messageBuffer: Object.freeze({
        create: (encoding) => new MessageBuffer(encoding)
      }),
      applicationJson: Object.freeze({
        encoder: Object.freeze({
          name: "application/json",
          encode: (msg, options) => {
            try {
              return Promise.resolve(Buffer.from(JSON.stringify(msg, void 0, 0), options.charset));
            } catch (err) {
              return Promise.reject(err);
            }
          }
        }),
        decoder: Object.freeze({
          name: "application/json",
          decode: (buffer, options) => {
            try {
              if (buffer instanceof Buffer) {
                return Promise.resolve(JSON.parse(buffer.toString(options.charset)));
              } else {
                return Promise.resolve(JSON.parse(new util_1.TextDecoder(options.charset).decode(buffer)));
              }
            } catch (err) {
              return Promise.reject(err);
            }
          }
        })
      }),
      stream: Object.freeze({
        asReadableStream: (stream) => new ReadableStreamWrapper(stream),
        asWritableStream: (stream) => new WritableStreamWrapper(stream)
      }),
      console,
      timer: Object.freeze({
        setTimeout(callback, ms, ...args) {
          return setTimeout(callback, ms, ...args);
        },
        clearTimeout(handle) {
          clearTimeout(handle);
        },
        setImmediate(callback, ...args) {
          return setImmediate(callback, ...args);
        },
        clearImmediate(handle) {
          clearImmediate(handle);
        }
      })
    });
    function RIL() {
      return _ril;
    }
    (function(RIL2) {
      function install() {
        ral_1.default.install(_ril);
      }
      RIL2.install = install;
    })(RIL || (RIL = {}));
    exports2.default = RIL;
  }
});

// node_modules/vscode-jsonrpc/lib/common/is.js
var require_is2 = __commonJS({
  "node_modules/vscode-jsonrpc/lib/common/is.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.stringArray = exports2.array = exports2.func = exports2.error = exports2.number = exports2.string = exports2.boolean = void 0;
    function boolean(value) {
      return value === true || value === false;
    }
    exports2.boolean = boolean;
    function string(value) {
      return typeof value === "string" || value instanceof String;
    }
    exports2.string = string;
    function number(value) {
      return typeof value === "number" || value instanceof Number;
    }
    exports2.number = number;
    function error(value) {
      return value instanceof Error;
    }
    exports2.error = error;
    function func(value) {
      return typeof value === "function";
    }
    exports2.func = func;
    function array(value) {
      return Array.isArray(value);
    }
    exports2.array = array;
    function stringArray(value) {
      return array(value) && value.every((elem) => string(elem));
    }
    exports2.stringArray = stringArray;
  }
});

// node_modules/vscode-jsonrpc/lib/common/messages.js
var require_messages = __commonJS({
  "node_modules/vscode-jsonrpc/lib/common/messages.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.isResponseMessage = exports2.isNotificationMessage = exports2.isRequestMessage = exports2.NotificationType9 = exports2.NotificationType8 = exports2.NotificationType7 = exports2.NotificationType6 = exports2.NotificationType5 = exports2.NotificationType4 = exports2.NotificationType3 = exports2.NotificationType2 = exports2.NotificationType1 = exports2.NotificationType0 = exports2.NotificationType = exports2.RequestType9 = exports2.RequestType8 = exports2.RequestType7 = exports2.RequestType6 = exports2.RequestType5 = exports2.RequestType4 = exports2.RequestType3 = exports2.RequestType2 = exports2.RequestType1 = exports2.RequestType = exports2.RequestType0 = exports2.AbstractMessageSignature = exports2.ParameterStructures = exports2.ResponseError = exports2.ErrorCodes = void 0;
    var is = require_is2();
    var ErrorCodes;
    (function(ErrorCodes2) {
      ErrorCodes2.ParseError = -32700;
      ErrorCodes2.InvalidRequest = -32600;
      ErrorCodes2.MethodNotFound = -32601;
      ErrorCodes2.InvalidParams = -32602;
      ErrorCodes2.InternalError = -32603;
      ErrorCodes2.jsonrpcReservedErrorRangeStart = -32099;
      ErrorCodes2.serverErrorStart = ErrorCodes2.jsonrpcReservedErrorRangeStart;
      ErrorCodes2.MessageWriteError = -32099;
      ErrorCodes2.MessageReadError = -32098;
      ErrorCodes2.ServerNotInitialized = -32002;
      ErrorCodes2.UnknownErrorCode = -32001;
      ErrorCodes2.jsonrpcReservedErrorRangeEnd = -32e3;
      ErrorCodes2.serverErrorEnd = ErrorCodes2.jsonrpcReservedErrorRangeEnd;
    })(ErrorCodes = exports2.ErrorCodes || (exports2.ErrorCodes = {}));
    var ResponseError = class _ResponseError extends Error {
      constructor(code, message, data) {
        super(message);
        this.code = is.number(code) ? code : ErrorCodes.UnknownErrorCode;
        this.data = data;
        Object.setPrototypeOf(this, _ResponseError.prototype);
      }
      toJson() {
        return {
          code: this.code,
          message: this.message,
          data: this.data
        };
      }
    };
    exports2.ResponseError = ResponseError;
    var ParameterStructures = class _ParameterStructures {
      constructor(kind) {
        this.kind = kind;
      }
      static is(value) {
        return value === _ParameterStructures.auto || value === _ParameterStructures.byName || value === _ParameterStructures.byPosition;
      }
      toString() {
        return this.kind;
      }
    };
    exports2.ParameterStructures = ParameterStructures;
    ParameterStructures.auto = new ParameterStructures("auto");
    ParameterStructures.byPosition = new ParameterStructures("byPosition");
    ParameterStructures.byName = new ParameterStructures("byName");
    var AbstractMessageSignature = class {
      constructor(method, numberOfParams) {
        this.method = method;
        this.numberOfParams = numberOfParams;
      }
      get parameterStructures() {
        return ParameterStructures.auto;
      }
    };
    exports2.AbstractMessageSignature = AbstractMessageSignature;
    var RequestType0 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 0);
      }
    };
    exports2.RequestType0 = RequestType0;
    var RequestType = class extends AbstractMessageSignature {
      constructor(method, _parameterStructures = ParameterStructures.auto) {
        super(method, 1);
        this._parameterStructures = _parameterStructures;
      }
      get parameterStructures() {
        return this._parameterStructures;
      }
    };
    exports2.RequestType = RequestType;
    var RequestType1 = class extends AbstractMessageSignature {
      constructor(method, _parameterStructures = ParameterStructures.auto) {
        super(method, 1);
        this._parameterStructures = _parameterStructures;
      }
      get parameterStructures() {
        return this._parameterStructures;
      }
    };
    exports2.RequestType1 = RequestType1;
    var RequestType2 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 2);
      }
    };
    exports2.RequestType2 = RequestType2;
    var RequestType3 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 3);
      }
    };
    exports2.RequestType3 = RequestType3;
    var RequestType4 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 4);
      }
    };
    exports2.RequestType4 = RequestType4;
    var RequestType5 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 5);
      }
    };
    exports2.RequestType5 = RequestType5;
    var RequestType6 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 6);
      }
    };
    exports2.RequestType6 = RequestType6;
    var RequestType7 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 7);
      }
    };
    exports2.RequestType7 = RequestType7;
    var RequestType8 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 8);
      }
    };
    exports2.RequestType8 = RequestType8;
    var RequestType9 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 9);
      }
    };
    exports2.RequestType9 = RequestType9;
    var NotificationType = class extends AbstractMessageSignature {
      constructor(method, _parameterStructures = ParameterStructures.auto) {
        super(method, 1);
        this._parameterStructures = _parameterStructures;
      }
      get parameterStructures() {
        return this._parameterStructures;
      }
    };
    exports2.NotificationType = NotificationType;
    var NotificationType0 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 0);
      }
    };
    exports2.NotificationType0 = NotificationType0;
    var NotificationType1 = class extends AbstractMessageSignature {
      constructor(method, _parameterStructures = ParameterStructures.auto) {
        super(method, 1);
        this._parameterStructures = _parameterStructures;
      }
      get parameterStructures() {
        return this._parameterStructures;
      }
    };
    exports2.NotificationType1 = NotificationType1;
    var NotificationType2 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 2);
      }
    };
    exports2.NotificationType2 = NotificationType2;
    var NotificationType3 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 3);
      }
    };
    exports2.NotificationType3 = NotificationType3;
    var NotificationType4 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 4);
      }
    };
    exports2.NotificationType4 = NotificationType4;
    var NotificationType5 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 5);
      }
    };
    exports2.NotificationType5 = NotificationType5;
    var NotificationType6 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 6);
      }
    };
    exports2.NotificationType6 = NotificationType6;
    var NotificationType7 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 7);
      }
    };
    exports2.NotificationType7 = NotificationType7;
    var NotificationType8 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 8);
      }
    };
    exports2.NotificationType8 = NotificationType8;
    var NotificationType9 = class extends AbstractMessageSignature {
      constructor(method) {
        super(method, 9);
      }
    };
    exports2.NotificationType9 = NotificationType9;
    function isRequestMessage(message) {
      const candidate = message;
      return candidate && is.string(candidate.method) && (is.string(candidate.id) || is.number(candidate.id));
    }
    exports2.isRequestMessage = isRequestMessage;
    function isNotificationMessage(message) {
      const candidate = message;
      return candidate && is.string(candidate.method) && message.id === void 0;
    }
    exports2.isNotificationMessage = isNotificationMessage;
    function isResponseMessage(message) {
      const candidate = message;
      return candidate && (candidate.result !== void 0 || !!candidate.error) && (is.string(candidate.id) || is.number(candidate.id) || candidate.id === null);
    }
    exports2.isResponseMessage = isResponseMessage;
  }
});

// node_modules/vscode-jsonrpc/lib/common/events.js
var require_events = __commonJS({
  "node_modules/vscode-jsonrpc/lib/common/events.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.Emitter = exports2.Event = void 0;
    var ral_1 = require_ral();
    var Event;
    (function(Event2) {
      const _disposable = { dispose() {
      } };
      Event2.None = function() {
        return _disposable;
      };
    })(Event = exports2.Event || (exports2.Event = {}));
    var CallbackList = class {
      add(callback, context = null, bucket) {
        if (!this._callbacks) {
          this._callbacks = [];
          this._contexts = [];
        }
        this._callbacks.push(callback);
        this._contexts.push(context);
        if (Array.isArray(bucket)) {
          bucket.push({ dispose: () => this.remove(callback, context) });
        }
      }
      remove(callback, context = null) {
        if (!this._callbacks) {
          return;
        }
        let foundCallbackWithDifferentContext = false;
        for (let i = 0, len = this._callbacks.length; i < len; i++) {
          if (this._callbacks[i] === callback) {
            if (this._contexts[i] === context) {
              this._callbacks.splice(i, 1);
              this._contexts.splice(i, 1);
              return;
            } else {
              foundCallbackWithDifferentContext = true;
            }
          }
        }
        if (foundCallbackWithDifferentContext) {
          throw new Error("When adding a listener with a context, you should remove it with the same context");
        }
      }
      invoke(...args) {
        if (!this._callbacks) {
          return [];
        }
        const ret = [], callbacks = this._callbacks.slice(0), contexts = this._contexts.slice(0);
        for (let i = 0, len = callbacks.length; i < len; i++) {
          try {
            ret.push(callbacks[i].apply(contexts[i], args));
          } catch (e) {
            ral_1.default().console.error(e);
          }
        }
        return ret;
      }
      isEmpty() {
        return !this._callbacks || this._callbacks.length === 0;
      }
      dispose() {
        this._callbacks = void 0;
        this._contexts = void 0;
      }
    };
    var Emitter = class _Emitter {
      constructor(_options) {
        this._options = _options;
      }
      /**
       * For the public to allow to subscribe
       * to events from this Emitter
       */
      get event() {
        if (!this._event) {
          this._event = (listener, thisArgs, disposables) => {
            if (!this._callbacks) {
              this._callbacks = new CallbackList();
            }
            if (this._options && this._options.onFirstListenerAdd && this._callbacks.isEmpty()) {
              this._options.onFirstListenerAdd(this);
            }
            this._callbacks.add(listener, thisArgs);
            const result = {
              dispose: () => {
                if (!this._callbacks) {
                  return;
                }
                this._callbacks.remove(listener, thisArgs);
                result.dispose = _Emitter._noop;
                if (this._options && this._options.onLastListenerRemove && this._callbacks.isEmpty()) {
                  this._options.onLastListenerRemove(this);
                }
              }
            };
            if (Array.isArray(disposables)) {
              disposables.push(result);
            }
            return result;
          };
        }
        return this._event;
      }
      /**
       * To be kept private to fire an event to
       * subscribers
       */
      fire(event) {
        if (this._callbacks) {
          this._callbacks.invoke.call(this._callbacks, event);
        }
      }
      dispose() {
        if (this._callbacks) {
          this._callbacks.dispose();
          this._callbacks = void 0;
        }
      }
    };
    exports2.Emitter = Emitter;
    Emitter._noop = function() {
    };
  }
});

// node_modules/vscode-jsonrpc/lib/common/cancellation.js
var require_cancellation = __commonJS({
  "node_modules/vscode-jsonrpc/lib/common/cancellation.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.CancellationTokenSource = exports2.CancellationToken = void 0;
    var ral_1 = require_ral();
    var Is = require_is2();
    var events_1 = require_events();
    var CancellationToken;
    (function(CancellationToken2) {
      CancellationToken2.None = Object.freeze({
        isCancellationRequested: false,
        onCancellationRequested: events_1.Event.None
      });
      CancellationToken2.Cancelled = Object.freeze({
        isCancellationRequested: true,
        onCancellationRequested: events_1.Event.None
      });
      function is(value) {
        const candidate = value;
        return candidate && (candidate === CancellationToken2.None || candidate === CancellationToken2.Cancelled || Is.boolean(candidate.isCancellationRequested) && !!candidate.onCancellationRequested);
      }
      CancellationToken2.is = is;
    })(CancellationToken = exports2.CancellationToken || (exports2.CancellationToken = {}));
    var shortcutEvent = Object.freeze(function(callback, context) {
      const handle = ral_1.default().timer.setTimeout(callback.bind(context), 0);
      return { dispose() {
        ral_1.default().timer.clearTimeout(handle);
      } };
    });
    var MutableToken = class {
      constructor() {
        this._isCancelled = false;
      }
      cancel() {
        if (!this._isCancelled) {
          this._isCancelled = true;
          if (this._emitter) {
            this._emitter.fire(void 0);
            this.dispose();
          }
        }
      }
      get isCancellationRequested() {
        return this._isCancelled;
      }
      get onCancellationRequested() {
        if (this._isCancelled) {
          return shortcutEvent;
        }
        if (!this._emitter) {
          this._emitter = new events_1.Emitter();
        }
        return this._emitter.event;
      }
      dispose() {
        if (this._emitter) {
          this._emitter.dispose();
          this._emitter = void 0;
        }
      }
    };
    var CancellationTokenSource = class {
      get token() {
        if (!this._token) {
          this._token = new MutableToken();
        }
        return this._token;
      }
      cancel() {
        if (!this._token) {
          this._token = CancellationToken.Cancelled;
        } else {
          this._token.cancel();
        }
      }
      dispose() {
        if (!this._token) {
          this._token = CancellationToken.None;
        } else if (this._token instanceof MutableToken) {
          this._token.dispose();
        }
      }
    };
    exports2.CancellationTokenSource = CancellationTokenSource;
  }
});

// node_modules/vscode-jsonrpc/lib/common/messageReader.js
var require_messageReader = __commonJS({
  "node_modules/vscode-jsonrpc/lib/common/messageReader.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.ReadableStreamMessageReader = exports2.AbstractMessageReader = exports2.MessageReader = void 0;
    var ral_1 = require_ral();
    var Is = require_is2();
    var events_1 = require_events();
    var MessageReader;
    (function(MessageReader2) {
      function is(value) {
        let candidate = value;
        return candidate && Is.func(candidate.listen) && Is.func(candidate.dispose) && Is.func(candidate.onError) && Is.func(candidate.onClose) && Is.func(candidate.onPartialMessage);
      }
      MessageReader2.is = is;
    })(MessageReader = exports2.MessageReader || (exports2.MessageReader = {}));
    var AbstractMessageReader = class {
      constructor() {
        this.errorEmitter = new events_1.Emitter();
        this.closeEmitter = new events_1.Emitter();
        this.partialMessageEmitter = new events_1.Emitter();
      }
      dispose() {
        this.errorEmitter.dispose();
        this.closeEmitter.dispose();
      }
      get onError() {
        return this.errorEmitter.event;
      }
      fireError(error) {
        this.errorEmitter.fire(this.asError(error));
      }
      get onClose() {
        return this.closeEmitter.event;
      }
      fireClose() {
        this.closeEmitter.fire(void 0);
      }
      get onPartialMessage() {
        return this.partialMessageEmitter.event;
      }
      firePartialMessage(info) {
        this.partialMessageEmitter.fire(info);
      }
      asError(error) {
        if (error instanceof Error) {
          return error;
        } else {
          return new Error(`Reader received error. Reason: ${Is.string(error.message) ? error.message : "unknown"}`);
        }
      }
    };
    exports2.AbstractMessageReader = AbstractMessageReader;
    var ResolvedMessageReaderOptions;
    (function(ResolvedMessageReaderOptions2) {
      function fromOptions(options) {
        var _a;
        let charset;
        let result;
        let contentDecoder;
        const contentDecoders = /* @__PURE__ */ new Map();
        let contentTypeDecoder;
        const contentTypeDecoders = /* @__PURE__ */ new Map();
        if (options === void 0 || typeof options === "string") {
          charset = options !== null && options !== void 0 ? options : "utf-8";
        } else {
          charset = (_a = options.charset) !== null && _a !== void 0 ? _a : "utf-8";
          if (options.contentDecoder !== void 0) {
            contentDecoder = options.contentDecoder;
            contentDecoders.set(contentDecoder.name, contentDecoder);
          }
          if (options.contentDecoders !== void 0) {
            for (const decoder of options.contentDecoders) {
              contentDecoders.set(decoder.name, decoder);
            }
          }
          if (options.contentTypeDecoder !== void 0) {
            contentTypeDecoder = options.contentTypeDecoder;
            contentTypeDecoders.set(contentTypeDecoder.name, contentTypeDecoder);
          }
          if (options.contentTypeDecoders !== void 0) {
            for (const decoder of options.contentTypeDecoders) {
              contentTypeDecoders.set(decoder.name, decoder);
            }
          }
        }
        if (contentTypeDecoder === void 0) {
          contentTypeDecoder = ral_1.default().applicationJson.decoder;
          contentTypeDecoders.set(contentTypeDecoder.name, contentTypeDecoder);
        }
        return { charset, contentDecoder, contentDecoders, contentTypeDecoder, contentTypeDecoders };
      }
      ResolvedMessageReaderOptions2.fromOptions = fromOptions;
    })(ResolvedMessageReaderOptions || (ResolvedMessageReaderOptions = {}));
    var ReadableStreamMessageReader = class extends AbstractMessageReader {
      constructor(readable, options) {
        super();
        this.readable = readable;
        this.options = ResolvedMessageReaderOptions.fromOptions(options);
        this.buffer = ral_1.default().messageBuffer.create(this.options.charset);
        this._partialMessageTimeout = 1e4;
        this.nextMessageLength = -1;
        this.messageToken = 0;
      }
      set partialMessageTimeout(timeout) {
        this._partialMessageTimeout = timeout;
      }
      get partialMessageTimeout() {
        return this._partialMessageTimeout;
      }
      listen(callback) {
        this.nextMessageLength = -1;
        this.messageToken = 0;
        this.partialMessageTimer = void 0;
        this.callback = callback;
        const result = this.readable.onData((data) => {
          this.onData(data);
        });
        this.readable.onError((error) => this.fireError(error));
        this.readable.onClose(() => this.fireClose());
        return result;
      }
      onData(data) {
        this.buffer.append(data);
        while (true) {
          if (this.nextMessageLength === -1) {
            const headers = this.buffer.tryReadHeaders();
            if (!headers) {
              return;
            }
            const contentLength = headers.get("Content-Length");
            if (!contentLength) {
              throw new Error("Header must provide a Content-Length property.");
            }
            const length = parseInt(contentLength);
            if (isNaN(length)) {
              throw new Error("Content-Length value must be a number.");
            }
            this.nextMessageLength = length;
          }
          const body = this.buffer.tryReadBody(this.nextMessageLength);
          if (body === void 0) {
            this.setPartialMessageTimer();
            return;
          }
          this.clearPartialMessageTimer();
          this.nextMessageLength = -1;
          let p;
          if (this.options.contentDecoder !== void 0) {
            p = this.options.contentDecoder.decode(body);
          } else {
            p = Promise.resolve(body);
          }
          p.then((value) => {
            this.options.contentTypeDecoder.decode(value, this.options).then((msg) => {
              this.callback(msg);
            }, (error) => {
              this.fireError(error);
            });
          }, (error) => {
            this.fireError(error);
          });
        }
      }
      clearPartialMessageTimer() {
        if (this.partialMessageTimer) {
          ral_1.default().timer.clearTimeout(this.partialMessageTimer);
          this.partialMessageTimer = void 0;
        }
      }
      setPartialMessageTimer() {
        this.clearPartialMessageTimer();
        if (this._partialMessageTimeout <= 0) {
          return;
        }
        this.partialMessageTimer = ral_1.default().timer.setTimeout((token, timeout) => {
          this.partialMessageTimer = void 0;
          if (token === this.messageToken) {
            this.firePartialMessage({ messageToken: token, waitingTime: timeout });
            this.setPartialMessageTimer();
          }
        }, this._partialMessageTimeout, this.messageToken, this._partialMessageTimeout);
      }
    };
    exports2.ReadableStreamMessageReader = ReadableStreamMessageReader;
  }
});

// node_modules/vscode-jsonrpc/lib/common/semaphore.js
var require_semaphore = __commonJS({
  "node_modules/vscode-jsonrpc/lib/common/semaphore.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.Semaphore = void 0;
    var ral_1 = require_ral();
    var Semaphore = class {
      constructor(capacity = 1) {
        if (capacity <= 0) {
          throw new Error("Capacity must be greater than 0");
        }
        this._capacity = capacity;
        this._active = 0;
        this._waiting = [];
      }
      lock(thunk) {
        return new Promise((resolve, reject) => {
          this._waiting.push({ thunk, resolve, reject });
          this.runNext();
        });
      }
      get active() {
        return this._active;
      }
      runNext() {
        if (this._waiting.length === 0 || this._active === this._capacity) {
          return;
        }
        ral_1.default().timer.setImmediate(() => this.doRunNext());
      }
      doRunNext() {
        if (this._waiting.length === 0 || this._active === this._capacity) {
          return;
        }
        const next = this._waiting.shift();
        this._active++;
        if (this._active > this._capacity) {
          throw new Error(`To many thunks active`);
        }
        try {
          const result = next.thunk();
          if (result instanceof Promise) {
            result.then((value) => {
              this._active--;
              next.resolve(value);
              this.runNext();
            }, (err) => {
              this._active--;
              next.reject(err);
              this.runNext();
            });
          } else {
            this._active--;
            next.resolve(result);
            this.runNext();
          }
        } catch (err) {
          this._active--;
          next.reject(err);
          this.runNext();
        }
      }
    };
    exports2.Semaphore = Semaphore;
  }
});

// node_modules/vscode-jsonrpc/lib/common/messageWriter.js
var require_messageWriter = __commonJS({
  "node_modules/vscode-jsonrpc/lib/common/messageWriter.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.WriteableStreamMessageWriter = exports2.AbstractMessageWriter = exports2.MessageWriter = void 0;
    var ral_1 = require_ral();
    var Is = require_is2();
    var semaphore_1 = require_semaphore();
    var events_1 = require_events();
    var ContentLength = "Content-Length: ";
    var CRLF = "\r\n";
    var MessageWriter;
    (function(MessageWriter2) {
      function is(value) {
        let candidate = value;
        return candidate && Is.func(candidate.dispose) && Is.func(candidate.onClose) && Is.func(candidate.onError) && Is.func(candidate.write);
      }
      MessageWriter2.is = is;
    })(MessageWriter = exports2.MessageWriter || (exports2.MessageWriter = {}));
    var AbstractMessageWriter = class {
      constructor() {
        this.errorEmitter = new events_1.Emitter();
        this.closeEmitter = new events_1.Emitter();
      }
      dispose() {
        this.errorEmitter.dispose();
        this.closeEmitter.dispose();
      }
      get onError() {
        return this.errorEmitter.event;
      }
      fireError(error, message, count) {
        this.errorEmitter.fire([this.asError(error), message, count]);
      }
      get onClose() {
        return this.closeEmitter.event;
      }
      fireClose() {
        this.closeEmitter.fire(void 0);
      }
      asError(error) {
        if (error instanceof Error) {
          return error;
        } else {
          return new Error(`Writer received error. Reason: ${Is.string(error.message) ? error.message : "unknown"}`);
        }
      }
    };
    exports2.AbstractMessageWriter = AbstractMessageWriter;
    var ResolvedMessageWriterOptions;
    (function(ResolvedMessageWriterOptions2) {
      function fromOptions(options) {
        var _a, _b;
        if (options === void 0 || typeof options === "string") {
          return { charset: options !== null && options !== void 0 ? options : "utf-8", contentTypeEncoder: ral_1.default().applicationJson.encoder };
        } else {
          return { charset: (_a = options.charset) !== null && _a !== void 0 ? _a : "utf-8", contentEncoder: options.contentEncoder, contentTypeEncoder: (_b = options.contentTypeEncoder) !== null && _b !== void 0 ? _b : ral_1.default().applicationJson.encoder };
        }
      }
      ResolvedMessageWriterOptions2.fromOptions = fromOptions;
    })(ResolvedMessageWriterOptions || (ResolvedMessageWriterOptions = {}));
    var WriteableStreamMessageWriter = class extends AbstractMessageWriter {
      constructor(writable, options) {
        super();
        this.writable = writable;
        this.options = ResolvedMessageWriterOptions.fromOptions(options);
        this.errorCount = 0;
        this.writeSemaphore = new semaphore_1.Semaphore(1);
        this.writable.onError((error) => this.fireError(error));
        this.writable.onClose(() => this.fireClose());
      }
      async write(msg) {
        return this.writeSemaphore.lock(async () => {
          const payload = this.options.contentTypeEncoder.encode(msg, this.options).then((buffer) => {
            if (this.options.contentEncoder !== void 0) {
              return this.options.contentEncoder.encode(buffer);
            } else {
              return buffer;
            }
          });
          return payload.then((buffer) => {
            const headers = [];
            headers.push(ContentLength, buffer.byteLength.toString(), CRLF);
            headers.push(CRLF);
            return this.doWrite(msg, headers, buffer);
          }, (error) => {
            this.fireError(error);
            throw error;
          });
        });
      }
      async doWrite(msg, headers, data) {
        try {
          await this.writable.write(headers.join(""), "ascii");
          return this.writable.write(data);
        } catch (error) {
          this.handleError(error, msg);
          return Promise.reject(error);
        }
      }
      handleError(error, msg) {
        this.errorCount++;
        this.fireError(error, msg, this.errorCount);
      }
      end() {
        this.writable.end();
      }
    };
    exports2.WriteableStreamMessageWriter = WriteableStreamMessageWriter;
  }
});

// node_modules/vscode-jsonrpc/lib/common/linkedMap.js
var require_linkedMap = __commonJS({
  "node_modules/vscode-jsonrpc/lib/common/linkedMap.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.LRUCache = exports2.LinkedMap = exports2.Touch = void 0;
    var Touch;
    (function(Touch2) {
      Touch2.None = 0;
      Touch2.First = 1;
      Touch2.AsOld = Touch2.First;
      Touch2.Last = 2;
      Touch2.AsNew = Touch2.Last;
    })(Touch = exports2.Touch || (exports2.Touch = {}));
    var LinkedMap = class {
      constructor() {
        this[Symbol.toStringTag] = "LinkedMap";
        this._map = /* @__PURE__ */ new Map();
        this._head = void 0;
        this._tail = void 0;
        this._size = 0;
        this._state = 0;
      }
      clear() {
        this._map.clear();
        this._head = void 0;
        this._tail = void 0;
        this._size = 0;
        this._state++;
      }
      isEmpty() {
        return !this._head && !this._tail;
      }
      get size() {
        return this._size;
      }
      get first() {
        var _a;
        return (_a = this._head) === null || _a === void 0 ? void 0 : _a.value;
      }
      get last() {
        var _a;
        return (_a = this._tail) === null || _a === void 0 ? void 0 : _a.value;
      }
      has(key) {
        return this._map.has(key);
      }
      get(key, touch = Touch.None) {
        const item = this._map.get(key);
        if (!item) {
          return void 0;
        }
        if (touch !== Touch.None) {
          this.touch(item, touch);
        }
        return item.value;
      }
      set(key, value, touch = Touch.None) {
        let item = this._map.get(key);
        if (item) {
          item.value = value;
          if (touch !== Touch.None) {
            this.touch(item, touch);
          }
        } else {
          item = { key, value, next: void 0, previous: void 0 };
          switch (touch) {
            case Touch.None:
              this.addItemLast(item);
              break;
            case Touch.First:
              this.addItemFirst(item);
              break;
            case Touch.Last:
              this.addItemLast(item);
              break;
            default:
              this.addItemLast(item);
              break;
          }
          this._map.set(key, item);
          this._size++;
        }
        return this;
      }
      delete(key) {
        return !!this.remove(key);
      }
      remove(key) {
        const item = this._map.get(key);
        if (!item) {
          return void 0;
        }
        this._map.delete(key);
        this.removeItem(item);
        this._size--;
        return item.value;
      }
      shift() {
        if (!this._head && !this._tail) {
          return void 0;
        }
        if (!this._head || !this._tail) {
          throw new Error("Invalid list");
        }
        const item = this._head;
        this._map.delete(item.key);
        this.removeItem(item);
        this._size--;
        return item.value;
      }
      forEach(callbackfn, thisArg) {
        const state = this._state;
        let current = this._head;
        while (current) {
          if (thisArg) {
            callbackfn.bind(thisArg)(current.value, current.key, this);
          } else {
            callbackfn(current.value, current.key, this);
          }
          if (this._state !== state) {
            throw new Error(`LinkedMap got modified during iteration.`);
          }
          current = current.next;
        }
      }
      keys() {
        const map = this;
        const state = this._state;
        let current = this._head;
        const iterator = {
          [Symbol.iterator]() {
            return iterator;
          },
          next() {
            if (map._state !== state) {
              throw new Error(`LinkedMap got modified during iteration.`);
            }
            if (current) {
              const result = { value: current.key, done: false };
              current = current.next;
              return result;
            } else {
              return { value: void 0, done: true };
            }
          }
        };
        return iterator;
      }
      values() {
        const map = this;
        const state = this._state;
        let current = this._head;
        const iterator = {
          [Symbol.iterator]() {
            return iterator;
          },
          next() {
            if (map._state !== state) {
              throw new Error(`LinkedMap got modified during iteration.`);
            }
            if (current) {
              const result = { value: current.value, done: false };
              current = current.next;
              return result;
            } else {
              return { value: void 0, done: true };
            }
          }
        };
        return iterator;
      }
      entries() {
        const map = this;
        const state = this._state;
        let current = this._head;
        const iterator = {
          [Symbol.iterator]() {
            return iterator;
          },
          next() {
            if (map._state !== state) {
              throw new Error(`LinkedMap got modified during iteration.`);
            }
            if (current) {
              const result = { value: [current.key, current.value], done: false };
              current = current.next;
              return result;
            } else {
              return { value: void 0, done: true };
            }
          }
        };
        return iterator;
      }
      [Symbol.iterator]() {
        return this.entries();
      }
      trimOld(newSize) {
        if (newSize >= this.size) {
          return;
        }
        if (newSize === 0) {
          this.clear();
          return;
        }
        let current = this._head;
        let currentSize = this.size;
        while (current && currentSize > newSize) {
          this._map.delete(current.key);
          current = current.next;
          currentSize--;
        }
        this._head = current;
        this._size = currentSize;
        if (current) {
          current.previous = void 0;
        }
        this._state++;
      }
      addItemFirst(item) {
        if (!this._head && !this._tail) {
          this._tail = item;
        } else if (!this._head) {
          throw new Error("Invalid list");
        } else {
          item.next = this._head;
          this._head.previous = item;
        }
        this._head = item;
        this._state++;
      }
      addItemLast(item) {
        if (!this._head && !this._tail) {
          this._head = item;
        } else if (!this._tail) {
          throw new Error("Invalid list");
        } else {
          item.previous = this._tail;
          this._tail.next = item;
        }
        this._tail = item;
        this._state++;
      }
      removeItem(item) {
        if (item === this._head && item === this._tail) {
          this._head = void 0;
          this._tail = void 0;
        } else if (item === this._head) {
          if (!item.next) {
            throw new Error("Invalid list");
          }
          item.next.previous = void 0;
          this._head = item.next;
        } else if (item === this._tail) {
          if (!item.previous) {
            throw new Error("Invalid list");
          }
          item.previous.next = void 0;
          this._tail = item.previous;
        } else {
          const next = item.next;
          const previous = item.previous;
          if (!next || !previous) {
            throw new Error("Invalid list");
          }
          next.previous = previous;
          previous.next = next;
        }
        item.next = void 0;
        item.previous = void 0;
        this._state++;
      }
      touch(item, touch) {
        if (!this._head || !this._tail) {
          throw new Error("Invalid list");
        }
        if (touch !== Touch.First && touch !== Touch.Last) {
          return;
        }
        if (touch === Touch.First) {
          if (item === this._head) {
            return;
          }
          const next = item.next;
          const previous = item.previous;
          if (item === this._tail) {
            previous.next = void 0;
            this._tail = previous;
          } else {
            next.previous = previous;
            previous.next = next;
          }
          item.previous = void 0;
          item.next = this._head;
          this._head.previous = item;
          this._head = item;
          this._state++;
        } else if (touch === Touch.Last) {
          if (item === this._tail) {
            return;
          }
          const next = item.next;
          const previous = item.previous;
          if (item === this._head) {
            next.previous = void 0;
            this._head = next;
          } else {
            next.previous = previous;
            previous.next = next;
          }
          item.next = void 0;
          item.previous = this._tail;
          this._tail.next = item;
          this._tail = item;
          this._state++;
        }
      }
      toJSON() {
        const data = [];
        this.forEach((value, key) => {
          data.push([key, value]);
        });
        return data;
      }
      fromJSON(data) {
        this.clear();
        for (const [key, value] of data) {
          this.set(key, value);
        }
      }
    };
    exports2.LinkedMap = LinkedMap;
    var LRUCache = class extends LinkedMap {
      constructor(limit, ratio = 1) {
        super();
        this._limit = limit;
        this._ratio = Math.min(Math.max(0, ratio), 1);
      }
      get limit() {
        return this._limit;
      }
      set limit(limit) {
        this._limit = limit;
        this.checkTrim();
      }
      get ratio() {
        return this._ratio;
      }
      set ratio(ratio) {
        this._ratio = Math.min(Math.max(0, ratio), 1);
        this.checkTrim();
      }
      get(key, touch = Touch.AsNew) {
        return super.get(key, touch);
      }
      peek(key) {
        return super.get(key, Touch.None);
      }
      set(key, value) {
        super.set(key, value, Touch.Last);
        this.checkTrim();
        return this;
      }
      checkTrim() {
        if (this.size > this._limit) {
          this.trimOld(Math.round(this._limit * this._ratio));
        }
      }
    };
    exports2.LRUCache = LRUCache;
  }
});

// node_modules/vscode-jsonrpc/lib/common/connection.js
var require_connection = __commonJS({
  "node_modules/vscode-jsonrpc/lib/common/connection.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.createMessageConnection = exports2.ConnectionOptions = exports2.CancellationStrategy = exports2.CancellationSenderStrategy = exports2.CancellationReceiverStrategy = exports2.ConnectionStrategy = exports2.ConnectionError = exports2.ConnectionErrors = exports2.LogTraceNotification = exports2.SetTraceNotification = exports2.TraceFormat = exports2.Trace = exports2.NullLogger = exports2.ProgressType = void 0;
    var ral_1 = require_ral();
    var Is = require_is2();
    var messages_1 = require_messages();
    var linkedMap_1 = require_linkedMap();
    var events_1 = require_events();
    var cancellation_1 = require_cancellation();
    var CancelNotification;
    (function(CancelNotification2) {
      CancelNotification2.type = new messages_1.NotificationType("$/cancelRequest");
    })(CancelNotification || (CancelNotification = {}));
    var ProgressNotification;
    (function(ProgressNotification2) {
      ProgressNotification2.type = new messages_1.NotificationType("$/progress");
    })(ProgressNotification || (ProgressNotification = {}));
    var ProgressType = class {
      constructor() {
      }
    };
    exports2.ProgressType = ProgressType;
    var StarRequestHandler;
    (function(StarRequestHandler2) {
      function is(value) {
        return Is.func(value);
      }
      StarRequestHandler2.is = is;
    })(StarRequestHandler || (StarRequestHandler = {}));
    exports2.NullLogger = Object.freeze({
      error: () => {
      },
      warn: () => {
      },
      info: () => {
      },
      log: () => {
      }
    });
    var Trace;
    (function(Trace2) {
      Trace2[Trace2["Off"] = 0] = "Off";
      Trace2[Trace2["Messages"] = 1] = "Messages";
      Trace2[Trace2["Verbose"] = 2] = "Verbose";
    })(Trace = exports2.Trace || (exports2.Trace = {}));
    (function(Trace2) {
      function fromString(value) {
        if (!Is.string(value)) {
          return Trace2.Off;
        }
        value = value.toLowerCase();
        switch (value) {
          case "off":
            return Trace2.Off;
          case "messages":
            return Trace2.Messages;
          case "verbose":
            return Trace2.Verbose;
          default:
            return Trace2.Off;
        }
      }
      Trace2.fromString = fromString;
      function toString(value) {
        switch (value) {
          case Trace2.Off:
            return "off";
          case Trace2.Messages:
            return "messages";
          case Trace2.Verbose:
            return "verbose";
          default:
            return "off";
        }
      }
      Trace2.toString = toString;
    })(Trace = exports2.Trace || (exports2.Trace = {}));
    var TraceFormat;
    (function(TraceFormat2) {
      TraceFormat2["Text"] = "text";
      TraceFormat2["JSON"] = "json";
    })(TraceFormat = exports2.TraceFormat || (exports2.TraceFormat = {}));
    (function(TraceFormat2) {
      function fromString(value) {
        value = value.toLowerCase();
        if (value === "json") {
          return TraceFormat2.JSON;
        } else {
          return TraceFormat2.Text;
        }
      }
      TraceFormat2.fromString = fromString;
    })(TraceFormat = exports2.TraceFormat || (exports2.TraceFormat = {}));
    var SetTraceNotification;
    (function(SetTraceNotification2) {
      SetTraceNotification2.type = new messages_1.NotificationType("$/setTrace");
    })(SetTraceNotification = exports2.SetTraceNotification || (exports2.SetTraceNotification = {}));
    var LogTraceNotification;
    (function(LogTraceNotification2) {
      LogTraceNotification2.type = new messages_1.NotificationType("$/logTrace");
    })(LogTraceNotification = exports2.LogTraceNotification || (exports2.LogTraceNotification = {}));
    var ConnectionErrors;
    (function(ConnectionErrors2) {
      ConnectionErrors2[ConnectionErrors2["Closed"] = 1] = "Closed";
      ConnectionErrors2[ConnectionErrors2["Disposed"] = 2] = "Disposed";
      ConnectionErrors2[ConnectionErrors2["AlreadyListening"] = 3] = "AlreadyListening";
    })(ConnectionErrors = exports2.ConnectionErrors || (exports2.ConnectionErrors = {}));
    var ConnectionError = class _ConnectionError extends Error {
      constructor(code, message) {
        super(message);
        this.code = code;
        Object.setPrototypeOf(this, _ConnectionError.prototype);
      }
    };
    exports2.ConnectionError = ConnectionError;
    var ConnectionStrategy;
    (function(ConnectionStrategy2) {
      function is(value) {
        const candidate = value;
        return candidate && Is.func(candidate.cancelUndispatched);
      }
      ConnectionStrategy2.is = is;
    })(ConnectionStrategy = exports2.ConnectionStrategy || (exports2.ConnectionStrategy = {}));
    var CancellationReceiverStrategy;
    (function(CancellationReceiverStrategy2) {
      CancellationReceiverStrategy2.Message = Object.freeze({
        createCancellationTokenSource(_) {
          return new cancellation_1.CancellationTokenSource();
        }
      });
      function is(value) {
        const candidate = value;
        return candidate && Is.func(candidate.createCancellationTokenSource);
      }
      CancellationReceiverStrategy2.is = is;
    })(CancellationReceiverStrategy = exports2.CancellationReceiverStrategy || (exports2.CancellationReceiverStrategy = {}));
    var CancellationSenderStrategy;
    (function(CancellationSenderStrategy2) {
      CancellationSenderStrategy2.Message = Object.freeze({
        sendCancellation(conn, id) {
          conn.sendNotification(CancelNotification.type, { id });
        },
        cleanup(_) {
        }
      });
      function is(value) {
        const candidate = value;
        return candidate && Is.func(candidate.sendCancellation) && Is.func(candidate.cleanup);
      }
      CancellationSenderStrategy2.is = is;
    })(CancellationSenderStrategy = exports2.CancellationSenderStrategy || (exports2.CancellationSenderStrategy = {}));
    var CancellationStrategy;
    (function(CancellationStrategy2) {
      CancellationStrategy2.Message = Object.freeze({
        receiver: CancellationReceiverStrategy.Message,
        sender: CancellationSenderStrategy.Message
      });
      function is(value) {
        const candidate = value;
        return candidate && CancellationReceiverStrategy.is(candidate.receiver) && CancellationSenderStrategy.is(candidate.sender);
      }
      CancellationStrategy2.is = is;
    })(CancellationStrategy = exports2.CancellationStrategy || (exports2.CancellationStrategy = {}));
    var ConnectionOptions;
    (function(ConnectionOptions2) {
      function is(value) {
        const candidate = value;
        return candidate && (CancellationStrategy.is(candidate.cancellationStrategy) || ConnectionStrategy.is(candidate.connectionStrategy));
      }
      ConnectionOptions2.is = is;
    })(ConnectionOptions = exports2.ConnectionOptions || (exports2.ConnectionOptions = {}));
    var ConnectionState;
    (function(ConnectionState2) {
      ConnectionState2[ConnectionState2["New"] = 1] = "New";
      ConnectionState2[ConnectionState2["Listening"] = 2] = "Listening";
      ConnectionState2[ConnectionState2["Closed"] = 3] = "Closed";
      ConnectionState2[ConnectionState2["Disposed"] = 4] = "Disposed";
    })(ConnectionState || (ConnectionState = {}));
    function createMessageConnection(messageReader, messageWriter, _logger, options) {
      const logger = _logger !== void 0 ? _logger : exports2.NullLogger;
      let sequenceNumber = 0;
      let notificationSquenceNumber = 0;
      let unknownResponseSquenceNumber = 0;
      const version = "2.0";
      let starRequestHandler = void 0;
      const requestHandlers = /* @__PURE__ */ Object.create(null);
      let starNotificationHandler = void 0;
      const notificationHandlers = /* @__PURE__ */ Object.create(null);
      const progressHandlers = /* @__PURE__ */ new Map();
      let timer;
      let messageQueue = new linkedMap_1.LinkedMap();
      let responsePromises = /* @__PURE__ */ Object.create(null);
      let requestTokens = /* @__PURE__ */ Object.create(null);
      let trace = Trace.Off;
      let traceFormat = TraceFormat.Text;
      let tracer;
      let state = ConnectionState.New;
      const errorEmitter = new events_1.Emitter();
      const closeEmitter = new events_1.Emitter();
      const unhandledNotificationEmitter = new events_1.Emitter();
      const unhandledProgressEmitter = new events_1.Emitter();
      const disposeEmitter = new events_1.Emitter();
      const cancellationStrategy = options && options.cancellationStrategy ? options.cancellationStrategy : CancellationStrategy.Message;
      function createRequestQueueKey(id) {
        if (id === null) {
          throw new Error(`Can't send requests with id null since the response can't be correlated.`);
        }
        return "req-" + id.toString();
      }
      function createResponseQueueKey(id) {
        if (id === null) {
          return "res-unknown-" + (++unknownResponseSquenceNumber).toString();
        } else {
          return "res-" + id.toString();
        }
      }
      function createNotificationQueueKey() {
        return "not-" + (++notificationSquenceNumber).toString();
      }
      function addMessageToQueue(queue, message) {
        if (messages_1.isRequestMessage(message)) {
          queue.set(createRequestQueueKey(message.id), message);
        } else if (messages_1.isResponseMessage(message)) {
          queue.set(createResponseQueueKey(message.id), message);
        } else {
          queue.set(createNotificationQueueKey(), message);
        }
      }
      function cancelUndispatched(_message) {
        return void 0;
      }
      function isListening() {
        return state === ConnectionState.Listening;
      }
      function isClosed() {
        return state === ConnectionState.Closed;
      }
      function isDisposed() {
        return state === ConnectionState.Disposed;
      }
      function closeHandler() {
        if (state === ConnectionState.New || state === ConnectionState.Listening) {
          state = ConnectionState.Closed;
          closeEmitter.fire(void 0);
        }
      }
      function readErrorHandler(error) {
        errorEmitter.fire([error, void 0, void 0]);
      }
      function writeErrorHandler(data) {
        errorEmitter.fire(data);
      }
      messageReader.onClose(closeHandler);
      messageReader.onError(readErrorHandler);
      messageWriter.onClose(closeHandler);
      messageWriter.onError(writeErrorHandler);
      function triggerMessageQueue() {
        if (timer || messageQueue.size === 0) {
          return;
        }
        timer = ral_1.default().timer.setImmediate(() => {
          timer = void 0;
          processMessageQueue();
        });
      }
      function processMessageQueue() {
        if (messageQueue.size === 0) {
          return;
        }
        const message = messageQueue.shift();
        try {
          if (messages_1.isRequestMessage(message)) {
            handleRequest(message);
          } else if (messages_1.isNotificationMessage(message)) {
            handleNotification(message);
          } else if (messages_1.isResponseMessage(message)) {
            handleResponse(message);
          } else {
            handleInvalidMessage(message);
          }
        } finally {
          triggerMessageQueue();
        }
      }
      const callback = (message) => {
        try {
          if (messages_1.isNotificationMessage(message) && message.method === CancelNotification.type.method) {
            const key = createRequestQueueKey(message.params.id);
            const toCancel = messageQueue.get(key);
            if (messages_1.isRequestMessage(toCancel)) {
              const strategy = options === null || options === void 0 ? void 0 : options.connectionStrategy;
              const response = strategy && strategy.cancelUndispatched ? strategy.cancelUndispatched(toCancel, cancelUndispatched) : cancelUndispatched(toCancel);
              if (response && (response.error !== void 0 || response.result !== void 0)) {
                messageQueue.delete(key);
                response.id = toCancel.id;
                traceSendingResponse(response, message.method, Date.now());
                messageWriter.write(response);
                return;
              }
            }
          }
          addMessageToQueue(messageQueue, message);
        } finally {
          triggerMessageQueue();
        }
      };
      function handleRequest(requestMessage) {
        if (isDisposed()) {
          return;
        }
        function reply(resultOrError, method, startTime2) {
          const message = {
            jsonrpc: version,
            id: requestMessage.id
          };
          if (resultOrError instanceof messages_1.ResponseError) {
            message.error = resultOrError.toJson();
          } else {
            message.result = resultOrError === void 0 ? null : resultOrError;
          }
          traceSendingResponse(message, method, startTime2);
          messageWriter.write(message);
        }
        function replyError(error, method, startTime2) {
          const message = {
            jsonrpc: version,
            id: requestMessage.id,
            error: error.toJson()
          };
          traceSendingResponse(message, method, startTime2);
          messageWriter.write(message);
        }
        function replySuccess(result, method, startTime2) {
          if (result === void 0) {
            result = null;
          }
          const message = {
            jsonrpc: version,
            id: requestMessage.id,
            result
          };
          traceSendingResponse(message, method, startTime2);
          messageWriter.write(message);
        }
        traceReceivedRequest(requestMessage);
        const element = requestHandlers[requestMessage.method];
        let type;
        let requestHandler;
        if (element) {
          type = element.type;
          requestHandler = element.handler;
        }
        const startTime = Date.now();
        if (requestHandler || starRequestHandler) {
          const tokenKey = String(requestMessage.id);
          const cancellationSource = cancellationStrategy.receiver.createCancellationTokenSource(tokenKey);
          requestTokens[tokenKey] = cancellationSource;
          try {
            let handlerResult;
            if (requestHandler) {
              if (requestMessage.params === void 0) {
                if (type !== void 0 && type.numberOfParams !== 0) {
                  replyError(new messages_1.ResponseError(messages_1.ErrorCodes.InvalidParams, `Request ${requestMessage.method} defines ${type.numberOfParams} params but recevied none.`), requestMessage.method, startTime);
                  return;
                }
                handlerResult = requestHandler(cancellationSource.token);
              } else if (Array.isArray(requestMessage.params)) {
                if (type !== void 0 && type.parameterStructures === messages_1.ParameterStructures.byName) {
                  replyError(new messages_1.ResponseError(messages_1.ErrorCodes.InvalidParams, `Request ${requestMessage.method} defines parameters by name but received parameters by position`), requestMessage.method, startTime);
                  return;
                }
                handlerResult = requestHandler(...requestMessage.params, cancellationSource.token);
              } else {
                if (type !== void 0 && type.parameterStructures === messages_1.ParameterStructures.byPosition) {
                  replyError(new messages_1.ResponseError(messages_1.ErrorCodes.InvalidParams, `Request ${requestMessage.method} defines parameters by position but received parameters by name`), requestMessage.method, startTime);
                  return;
                }
                handlerResult = requestHandler(requestMessage.params, cancellationSource.token);
              }
            } else if (starRequestHandler) {
              handlerResult = starRequestHandler(requestMessage.method, requestMessage.params, cancellationSource.token);
            }
            const promise = handlerResult;
            if (!handlerResult) {
              delete requestTokens[tokenKey];
              replySuccess(handlerResult, requestMessage.method, startTime);
            } else if (promise.then) {
              promise.then((resultOrError) => {
                delete requestTokens[tokenKey];
                reply(resultOrError, requestMessage.method, startTime);
              }, (error) => {
                delete requestTokens[tokenKey];
                if (error instanceof messages_1.ResponseError) {
                  replyError(error, requestMessage.method, startTime);
                } else if (error && Is.string(error.message)) {
                  replyError(new messages_1.ResponseError(messages_1.ErrorCodes.InternalError, `Request ${requestMessage.method} failed with message: ${error.message}`), requestMessage.method, startTime);
                } else {
                  replyError(new messages_1.ResponseError(messages_1.ErrorCodes.InternalError, `Request ${requestMessage.method} failed unexpectedly without providing any details.`), requestMessage.method, startTime);
                }
              });
            } else {
              delete requestTokens[tokenKey];
              reply(handlerResult, requestMessage.method, startTime);
            }
          } catch (error) {
            delete requestTokens[tokenKey];
            if (error instanceof messages_1.ResponseError) {
              reply(error, requestMessage.method, startTime);
            } else if (error && Is.string(error.message)) {
              replyError(new messages_1.ResponseError(messages_1.ErrorCodes.InternalError, `Request ${requestMessage.method} failed with message: ${error.message}`), requestMessage.method, startTime);
            } else {
              replyError(new messages_1.ResponseError(messages_1.ErrorCodes.InternalError, `Request ${requestMessage.method} failed unexpectedly without providing any details.`), requestMessage.method, startTime);
            }
          }
        } else {
          replyError(new messages_1.ResponseError(messages_1.ErrorCodes.MethodNotFound, `Unhandled method ${requestMessage.method}`), requestMessage.method, startTime);
        }
      }
      function handleResponse(responseMessage) {
        if (isDisposed()) {
          return;
        }
        if (responseMessage.id === null) {
          if (responseMessage.error) {
            logger.error(`Received response message without id: Error is: 
${JSON.stringify(responseMessage.error, void 0, 4)}`);
          } else {
            logger.error(`Received response message without id. No further error information provided.`);
          }
        } else {
          const key = String(responseMessage.id);
          const responsePromise = responsePromises[key];
          traceReceivedResponse(responseMessage, responsePromise);
          if (responsePromise) {
            delete responsePromises[key];
            try {
              if (responseMessage.error) {
                const error = responseMessage.error;
                responsePromise.reject(new messages_1.ResponseError(error.code, error.message, error.data));
              } else if (responseMessage.result !== void 0) {
                responsePromise.resolve(responseMessage.result);
              } else {
                throw new Error("Should never happen.");
              }
            } catch (error) {
              if (error.message) {
                logger.error(`Response handler '${responsePromise.method}' failed with message: ${error.message}`);
              } else {
                logger.error(`Response handler '${responsePromise.method}' failed unexpectedly.`);
              }
            }
          }
        }
      }
      function handleNotification(message) {
        if (isDisposed()) {
          return;
        }
        let type = void 0;
        let notificationHandler;
        if (message.method === CancelNotification.type.method) {
          notificationHandler = (params) => {
            const id = params.id;
            const source = requestTokens[String(id)];
            if (source) {
              source.cancel();
            }
          };
        } else {
          const element = notificationHandlers[message.method];
          if (element) {
            notificationHandler = element.handler;
            type = element.type;
          }
        }
        if (notificationHandler || starNotificationHandler) {
          try {
            traceReceivedNotification(message);
            if (notificationHandler) {
              if (message.params === void 0) {
                if (type !== void 0) {
                  if (type.numberOfParams !== 0 && type.parameterStructures !== messages_1.ParameterStructures.byName) {
                    logger.error(`Notification ${message.method} defines ${type.numberOfParams} params but recevied none.`);
                  }
                }
                notificationHandler();
              } else if (Array.isArray(message.params)) {
                if (type !== void 0) {
                  if (type.parameterStructures === messages_1.ParameterStructures.byName) {
                    logger.error(`Notification ${message.method} defines parameters by name but received parameters by position`);
                  }
                  if (type.numberOfParams !== message.params.length) {
                    logger.error(`Notification ${message.method} defines ${type.numberOfParams} params but received ${message.params.length} argumennts`);
                  }
                }
                notificationHandler(...message.params);
              } else {
                if (type !== void 0 && type.parameterStructures === messages_1.ParameterStructures.byPosition) {
                  logger.error(`Notification ${message.method} defines parameters by position but received parameters by name`);
                }
                notificationHandler(message.params);
              }
            } else if (starNotificationHandler) {
              starNotificationHandler(message.method, message.params);
            }
          } catch (error) {
            if (error.message) {
              logger.error(`Notification handler '${message.method}' failed with message: ${error.message}`);
            } else {
              logger.error(`Notification handler '${message.method}' failed unexpectedly.`);
            }
          }
        } else {
          unhandledNotificationEmitter.fire(message);
        }
      }
      function handleInvalidMessage(message) {
        if (!message) {
          logger.error("Received empty message.");
          return;
        }
        logger.error(`Received message which is neither a response nor a notification message:
${JSON.stringify(message, null, 4)}`);
        const responseMessage = message;
        if (Is.string(responseMessage.id) || Is.number(responseMessage.id)) {
          const key = String(responseMessage.id);
          const responseHandler = responsePromises[key];
          if (responseHandler) {
            responseHandler.reject(new Error("The received response has neither a result nor an error property."));
          }
        }
      }
      function traceSendingRequest(message) {
        if (trace === Trace.Off || !tracer) {
          return;
        }
        if (traceFormat === TraceFormat.Text) {
          let data = void 0;
          if (trace === Trace.Verbose && message.params) {
            data = `Params: ${JSON.stringify(message.params, null, 4)}

`;
          }
          tracer.log(`Sending request '${message.method} - (${message.id})'.`, data);
        } else {
          logLSPMessage("send-request", message);
        }
      }
      function traceSendingNotification(message) {
        if (trace === Trace.Off || !tracer) {
          return;
        }
        if (traceFormat === TraceFormat.Text) {
          let data = void 0;
          if (trace === Trace.Verbose) {
            if (message.params) {
              data = `Params: ${JSON.stringify(message.params, null, 4)}

`;
            } else {
              data = "No parameters provided.\n\n";
            }
          }
          tracer.log(`Sending notification '${message.method}'.`, data);
        } else {
          logLSPMessage("send-notification", message);
        }
      }
      function traceSendingResponse(message, method, startTime) {
        if (trace === Trace.Off || !tracer) {
          return;
        }
        if (traceFormat === TraceFormat.Text) {
          let data = void 0;
          if (trace === Trace.Verbose) {
            if (message.error && message.error.data) {
              data = `Error data: ${JSON.stringify(message.error.data, null, 4)}

`;
            } else {
              if (message.result) {
                data = `Result: ${JSON.stringify(message.result, null, 4)}

`;
              } else if (message.error === void 0) {
                data = "No result returned.\n\n";
              }
            }
          }
          tracer.log(`Sending response '${method} - (${message.id})'. Processing request took ${Date.now() - startTime}ms`, data);
        } else {
          logLSPMessage("send-response", message);
        }
      }
      function traceReceivedRequest(message) {
        if (trace === Trace.Off || !tracer) {
          return;
        }
        if (traceFormat === TraceFormat.Text) {
          let data = void 0;
          if (trace === Trace.Verbose && message.params) {
            data = `Params: ${JSON.stringify(message.params, null, 4)}

`;
          }
          tracer.log(`Received request '${message.method} - (${message.id})'.`, data);
        } else {
          logLSPMessage("receive-request", message);
        }
      }
      function traceReceivedNotification(message) {
        if (trace === Trace.Off || !tracer || message.method === LogTraceNotification.type.method) {
          return;
        }
        if (traceFormat === TraceFormat.Text) {
          let data = void 0;
          if (trace === Trace.Verbose) {
            if (message.params) {
              data = `Params: ${JSON.stringify(message.params, null, 4)}

`;
            } else {
              data = "No parameters provided.\n\n";
            }
          }
          tracer.log(`Received notification '${message.method}'.`, data);
        } else {
          logLSPMessage("receive-notification", message);
        }
      }
      function traceReceivedResponse(message, responsePromise) {
        if (trace === Trace.Off || !tracer) {
          return;
        }
        if (traceFormat === TraceFormat.Text) {
          let data = void 0;
          if (trace === Trace.Verbose) {
            if (message.error && message.error.data) {
              data = `Error data: ${JSON.stringify(message.error.data, null, 4)}

`;
            } else {
              if (message.result) {
                data = `Result: ${JSON.stringify(message.result, null, 4)}

`;
              } else if (message.error === void 0) {
                data = "No result returned.\n\n";
              }
            }
          }
          if (responsePromise) {
            const error = message.error ? ` Request failed: ${message.error.message} (${message.error.code}).` : "";
            tracer.log(`Received response '${responsePromise.method} - (${message.id})' in ${Date.now() - responsePromise.timerStart}ms.${error}`, data);
          } else {
            tracer.log(`Received response ${message.id} without active response promise.`, data);
          }
        } else {
          logLSPMessage("receive-response", message);
        }
      }
      function logLSPMessage(type, message) {
        if (!tracer || trace === Trace.Off) {
          return;
        }
        const lspMessage = {
          isLSPMessage: true,
          type,
          message,
          timestamp: Date.now()
        };
        tracer.log(lspMessage);
      }
      function throwIfClosedOrDisposed() {
        if (isClosed()) {
          throw new ConnectionError(ConnectionErrors.Closed, "Connection is closed.");
        }
        if (isDisposed()) {
          throw new ConnectionError(ConnectionErrors.Disposed, "Connection is disposed.");
        }
      }
      function throwIfListening() {
        if (isListening()) {
          throw new ConnectionError(ConnectionErrors.AlreadyListening, "Connection is already listening");
        }
      }
      function throwIfNotListening() {
        if (!isListening()) {
          throw new Error("Call listen() first.");
        }
      }
      function undefinedToNull(param) {
        if (param === void 0) {
          return null;
        } else {
          return param;
        }
      }
      function nullToUndefined(param) {
        if (param === null) {
          return void 0;
        } else {
          return param;
        }
      }
      function isNamedParam(param) {
        return param !== void 0 && param !== null && !Array.isArray(param) && typeof param === "object";
      }
      function computeSingleParam(parameterStructures, param) {
        switch (parameterStructures) {
          case messages_1.ParameterStructures.auto:
            if (isNamedParam(param)) {
              return nullToUndefined(param);
            } else {
              return [undefinedToNull(param)];
            }
            break;
          case messages_1.ParameterStructures.byName:
            if (!isNamedParam(param)) {
              throw new Error(`Recevied parameters by name but param is not an object literal.`);
            }
            return nullToUndefined(param);
          case messages_1.ParameterStructures.byPosition:
            return [undefinedToNull(param)];
          default:
            throw new Error(`Unknown parameter structure ${parameterStructures.toString()}`);
        }
      }
      function computeMessageParams(type, params) {
        let result;
        const numberOfParams = type.numberOfParams;
        switch (numberOfParams) {
          case 0:
            result = void 0;
            break;
          case 1:
            result = computeSingleParam(type.parameterStructures, params[0]);
            break;
          default:
            result = [];
            for (let i = 0; i < params.length && i < numberOfParams; i++) {
              result.push(undefinedToNull(params[i]));
            }
            if (params.length < numberOfParams) {
              for (let i = params.length; i < numberOfParams; i++) {
                result.push(null);
              }
            }
            break;
        }
        return result;
      }
      const connection = {
        sendNotification: (type, ...args) => {
          throwIfClosedOrDisposed();
          let method;
          let messageParams;
          if (Is.string(type)) {
            method = type;
            const first = args[0];
            let paramStart = 0;
            let parameterStructures = messages_1.ParameterStructures.auto;
            if (messages_1.ParameterStructures.is(first)) {
              paramStart = 1;
              parameterStructures = first;
            }
            let paramEnd = args.length;
            const numberOfParams = paramEnd - paramStart;
            switch (numberOfParams) {
              case 0:
                messageParams = void 0;
                break;
              case 1:
                messageParams = computeSingleParam(parameterStructures, args[paramStart]);
                break;
              default:
                if (parameterStructures === messages_1.ParameterStructures.byName) {
                  throw new Error(`Recevied ${numberOfParams} parameters for 'by Name' notification parameter structure.`);
                }
                messageParams = args.slice(paramStart, paramEnd).map((value) => undefinedToNull(value));
                break;
            }
          } else {
            const params = args;
            method = type.method;
            messageParams = computeMessageParams(type, params);
          }
          const notificationMessage = {
            jsonrpc: version,
            method,
            params: messageParams
          };
          traceSendingNotification(notificationMessage);
          messageWriter.write(notificationMessage);
        },
        onNotification: (type, handler) => {
          throwIfClosedOrDisposed();
          let method;
          if (Is.func(type)) {
            starNotificationHandler = type;
          } else if (handler) {
            if (Is.string(type)) {
              method = type;
              notificationHandlers[type] = { type: void 0, handler };
            } else {
              method = type.method;
              notificationHandlers[type.method] = { type, handler };
            }
          }
          return {
            dispose: () => {
              if (method !== void 0) {
                delete notificationHandlers[method];
              } else {
                starNotificationHandler = void 0;
              }
            }
          };
        },
        onProgress: (_type, token, handler) => {
          if (progressHandlers.has(token)) {
            throw new Error(`Progress handler for token ${token} already registered`);
          }
          progressHandlers.set(token, handler);
          return {
            dispose: () => {
              progressHandlers.delete(token);
            }
          };
        },
        sendProgress: (_type, token, value) => {
          connection.sendNotification(ProgressNotification.type, { token, value });
        },
        onUnhandledProgress: unhandledProgressEmitter.event,
        sendRequest: (type, ...args) => {
          throwIfClosedOrDisposed();
          throwIfNotListening();
          let method;
          let messageParams;
          let token = void 0;
          if (Is.string(type)) {
            method = type;
            const first = args[0];
            const last = args[args.length - 1];
            let paramStart = 0;
            let parameterStructures = messages_1.ParameterStructures.auto;
            if (messages_1.ParameterStructures.is(first)) {
              paramStart = 1;
              parameterStructures = first;
            }
            let paramEnd = args.length;
            if (cancellation_1.CancellationToken.is(last)) {
              paramEnd = paramEnd - 1;
              token = last;
            }
            const numberOfParams = paramEnd - paramStart;
            switch (numberOfParams) {
              case 0:
                messageParams = void 0;
                break;
              case 1:
                messageParams = computeSingleParam(parameterStructures, args[paramStart]);
                break;
              default:
                if (parameterStructures === messages_1.ParameterStructures.byName) {
                  throw new Error(`Recevied ${numberOfParams} parameters for 'by Name' request parameter structure.`);
                }
                messageParams = args.slice(paramStart, paramEnd).map((value) => undefinedToNull(value));
                break;
            }
          } else {
            const params = args;
            method = type.method;
            messageParams = computeMessageParams(type, params);
            const numberOfParams = type.numberOfParams;
            token = cancellation_1.CancellationToken.is(params[numberOfParams]) ? params[numberOfParams] : void 0;
          }
          const id = sequenceNumber++;
          let disposable;
          if (token) {
            disposable = token.onCancellationRequested(() => {
              cancellationStrategy.sender.sendCancellation(connection, id);
            });
          }
          const result = new Promise((resolve, reject) => {
            const requestMessage = {
              jsonrpc: version,
              id,
              method,
              params: messageParams
            };
            const resolveWithCleanup = (r) => {
              resolve(r);
              cancellationStrategy.sender.cleanup(id);
              disposable === null || disposable === void 0 ? void 0 : disposable.dispose();
            };
            const rejectWithCleanup = (r) => {
              reject(r);
              cancellationStrategy.sender.cleanup(id);
              disposable === null || disposable === void 0 ? void 0 : disposable.dispose();
            };
            let responsePromise = { method, timerStart: Date.now(), resolve: resolveWithCleanup, reject: rejectWithCleanup };
            traceSendingRequest(requestMessage);
            try {
              messageWriter.write(requestMessage);
            } catch (e) {
              responsePromise.reject(new messages_1.ResponseError(messages_1.ErrorCodes.MessageWriteError, e.message ? e.message : "Unknown reason"));
              responsePromise = null;
            }
            if (responsePromise) {
              responsePromises[String(id)] = responsePromise;
            }
          });
          return result;
        },
        onRequest: (type, handler) => {
          throwIfClosedOrDisposed();
          let method = null;
          if (StarRequestHandler.is(type)) {
            method = void 0;
            starRequestHandler = type;
          } else if (Is.string(type)) {
            method = null;
            if (handler !== void 0) {
              method = type;
              requestHandlers[type] = { handler, type: void 0 };
            }
          } else {
            if (handler !== void 0) {
              method = type.method;
              requestHandlers[type.method] = { type, handler };
            }
          }
          return {
            dispose: () => {
              if (method === null) {
                return;
              }
              if (method !== void 0) {
                delete requestHandlers[method];
              } else {
                starRequestHandler = void 0;
              }
            }
          };
        },
        trace: (_value, _tracer, sendNotificationOrTraceOptions) => {
          let _sendNotification = false;
          let _traceFormat = TraceFormat.Text;
          if (sendNotificationOrTraceOptions !== void 0) {
            if (Is.boolean(sendNotificationOrTraceOptions)) {
              _sendNotification = sendNotificationOrTraceOptions;
            } else {
              _sendNotification = sendNotificationOrTraceOptions.sendNotification || false;
              _traceFormat = sendNotificationOrTraceOptions.traceFormat || TraceFormat.Text;
            }
          }
          trace = _value;
          traceFormat = _traceFormat;
          if (trace === Trace.Off) {
            tracer = void 0;
          } else {
            tracer = _tracer;
          }
          if (_sendNotification && !isClosed() && !isDisposed()) {
            connection.sendNotification(SetTraceNotification.type, { value: Trace.toString(_value) });
          }
        },
        onError: errorEmitter.event,
        onClose: closeEmitter.event,
        onUnhandledNotification: unhandledNotificationEmitter.event,
        onDispose: disposeEmitter.event,
        end: () => {
          messageWriter.end();
        },
        dispose: () => {
          if (isDisposed()) {
            return;
          }
          state = ConnectionState.Disposed;
          disposeEmitter.fire(void 0);
          const error = new Error("Connection got disposed.");
          Object.keys(responsePromises).forEach((key) => {
            responsePromises[key].reject(error);
          });
          responsePromises = /* @__PURE__ */ Object.create(null);
          requestTokens = /* @__PURE__ */ Object.create(null);
          messageQueue = new linkedMap_1.LinkedMap();
          if (Is.func(messageWriter.dispose)) {
            messageWriter.dispose();
          }
          if (Is.func(messageReader.dispose)) {
            messageReader.dispose();
          }
        },
        listen: () => {
          throwIfClosedOrDisposed();
          throwIfListening();
          state = ConnectionState.Listening;
          messageReader.listen(callback);
        },
        inspect: () => {
          ral_1.default().console.log("inspect");
        }
      };
      connection.onNotification(LogTraceNotification.type, (params) => {
        if (trace === Trace.Off || !tracer) {
          return;
        }
        tracer.log(params.message, trace === Trace.Verbose ? params.verbose : void 0);
      });
      connection.onNotification(ProgressNotification.type, (params) => {
        const handler = progressHandlers.get(params.token);
        if (handler) {
          handler(params.value);
        } else {
          unhandledProgressEmitter.fire(params);
        }
      });
      return connection;
    }
    exports2.createMessageConnection = createMessageConnection;
  }
});

// node_modules/vscode-jsonrpc/lib/common/api.js
var require_api = __commonJS({
  "node_modules/vscode-jsonrpc/lib/common/api.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.CancellationSenderStrategy = exports2.CancellationReceiverStrategy = exports2.ConnectionError = exports2.ConnectionErrors = exports2.LogTraceNotification = exports2.SetTraceNotification = exports2.TraceFormat = exports2.Trace = exports2.ProgressType = exports2.createMessageConnection = exports2.NullLogger = exports2.ConnectionOptions = exports2.ConnectionStrategy = exports2.WriteableStreamMessageWriter = exports2.AbstractMessageWriter = exports2.MessageWriter = exports2.ReadableStreamMessageReader = exports2.AbstractMessageReader = exports2.MessageReader = exports2.CancellationToken = exports2.CancellationTokenSource = exports2.Emitter = exports2.Event = exports2.Disposable = exports2.ParameterStructures = exports2.NotificationType9 = exports2.NotificationType8 = exports2.NotificationType7 = exports2.NotificationType6 = exports2.NotificationType5 = exports2.NotificationType4 = exports2.NotificationType3 = exports2.NotificationType2 = exports2.NotificationType1 = exports2.NotificationType0 = exports2.NotificationType = exports2.ErrorCodes = exports2.ResponseError = exports2.RequestType9 = exports2.RequestType8 = exports2.RequestType7 = exports2.RequestType6 = exports2.RequestType5 = exports2.RequestType4 = exports2.RequestType3 = exports2.RequestType2 = exports2.RequestType1 = exports2.RequestType0 = exports2.RequestType = exports2.RAL = void 0;
    exports2.CancellationStrategy = void 0;
    var messages_1 = require_messages();
    Object.defineProperty(exports2, "RequestType", { enumerable: true, get: function() {
      return messages_1.RequestType;
    } });
    Object.defineProperty(exports2, "RequestType0", { enumerable: true, get: function() {
      return messages_1.RequestType0;
    } });
    Object.defineProperty(exports2, "RequestType1", { enumerable: true, get: function() {
      return messages_1.RequestType1;
    } });
    Object.defineProperty(exports2, "RequestType2", { enumerable: true, get: function() {
      return messages_1.RequestType2;
    } });
    Object.defineProperty(exports2, "RequestType3", { enumerable: true, get: function() {
      return messages_1.RequestType3;
    } });
    Object.defineProperty(exports2, "RequestType4", { enumerable: true, get: function() {
      return messages_1.RequestType4;
    } });
    Object.defineProperty(exports2, "RequestType5", { enumerable: true, get: function() {
      return messages_1.RequestType5;
    } });
    Object.defineProperty(exports2, "RequestType6", { enumerable: true, get: function() {
      return messages_1.RequestType6;
    } });
    Object.defineProperty(exports2, "RequestType7", { enumerable: true, get: function() {
      return messages_1.RequestType7;
    } });
    Object.defineProperty(exports2, "RequestType8", { enumerable: true, get: function() {
      return messages_1.RequestType8;
    } });
    Object.defineProperty(exports2, "RequestType9", { enumerable: true, get: function() {
      return messages_1.RequestType9;
    } });
    Object.defineProperty(exports2, "ResponseError", { enumerable: true, get: function() {
      return messages_1.ResponseError;
    } });
    Object.defineProperty(exports2, "ErrorCodes", { enumerable: true, get: function() {
      return messages_1.ErrorCodes;
    } });
    Object.defineProperty(exports2, "NotificationType", { enumerable: true, get: function() {
      return messages_1.NotificationType;
    } });
    Object.defineProperty(exports2, "NotificationType0", { enumerable: true, get: function() {
      return messages_1.NotificationType0;
    } });
    Object.defineProperty(exports2, "NotificationType1", { enumerable: true, get: function() {
      return messages_1.NotificationType1;
    } });
    Object.defineProperty(exports2, "NotificationType2", { enumerable: true, get: function() {
      return messages_1.NotificationType2;
    } });
    Object.defineProperty(exports2, "NotificationType3", { enumerable: true, get: function() {
      return messages_1.NotificationType3;
    } });
    Object.defineProperty(exports2, "NotificationType4", { enumerable: true, get: function() {
      return messages_1.NotificationType4;
    } });
    Object.defineProperty(exports2, "NotificationType5", { enumerable: true, get: function() {
      return messages_1.NotificationType5;
    } });
    Object.defineProperty(exports2, "NotificationType6", { enumerable: true, get: function() {
      return messages_1.NotificationType6;
    } });
    Object.defineProperty(exports2, "NotificationType7", { enumerable: true, get: function() {
      return messages_1.NotificationType7;
    } });
    Object.defineProperty(exports2, "NotificationType8", { enumerable: true, get: function() {
      return messages_1.NotificationType8;
    } });
    Object.defineProperty(exports2, "NotificationType9", { enumerable: true, get: function() {
      return messages_1.NotificationType9;
    } });
    Object.defineProperty(exports2, "ParameterStructures", { enumerable: true, get: function() {
      return messages_1.ParameterStructures;
    } });
    var disposable_1 = require_disposable();
    Object.defineProperty(exports2, "Disposable", { enumerable: true, get: function() {
      return disposable_1.Disposable;
    } });
    var events_1 = require_events();
    Object.defineProperty(exports2, "Event", { enumerable: true, get: function() {
      return events_1.Event;
    } });
    Object.defineProperty(exports2, "Emitter", { enumerable: true, get: function() {
      return events_1.Emitter;
    } });
    var cancellation_1 = require_cancellation();
    Object.defineProperty(exports2, "CancellationTokenSource", { enumerable: true, get: function() {
      return cancellation_1.CancellationTokenSource;
    } });
    Object.defineProperty(exports2, "CancellationToken", { enumerable: true, get: function() {
      return cancellation_1.CancellationToken;
    } });
    var messageReader_1 = require_messageReader();
    Object.defineProperty(exports2, "MessageReader", { enumerable: true, get: function() {
      return messageReader_1.MessageReader;
    } });
    Object.defineProperty(exports2, "AbstractMessageReader", { enumerable: true, get: function() {
      return messageReader_1.AbstractMessageReader;
    } });
    Object.defineProperty(exports2, "ReadableStreamMessageReader", { enumerable: true, get: function() {
      return messageReader_1.ReadableStreamMessageReader;
    } });
    var messageWriter_1 = require_messageWriter();
    Object.defineProperty(exports2, "MessageWriter", { enumerable: true, get: function() {
      return messageWriter_1.MessageWriter;
    } });
    Object.defineProperty(exports2, "AbstractMessageWriter", { enumerable: true, get: function() {
      return messageWriter_1.AbstractMessageWriter;
    } });
    Object.defineProperty(exports2, "WriteableStreamMessageWriter", { enumerable: true, get: function() {
      return messageWriter_1.WriteableStreamMessageWriter;
    } });
    var connection_1 = require_connection();
    Object.defineProperty(exports2, "ConnectionStrategy", { enumerable: true, get: function() {
      return connection_1.ConnectionStrategy;
    } });
    Object.defineProperty(exports2, "ConnectionOptions", { enumerable: true, get: function() {
      return connection_1.ConnectionOptions;
    } });
    Object.defineProperty(exports2, "NullLogger", { enumerable: true, get: function() {
      return connection_1.NullLogger;
    } });
    Object.defineProperty(exports2, "createMessageConnection", { enumerable: true, get: function() {
      return connection_1.createMessageConnection;
    } });
    Object.defineProperty(exports2, "ProgressType", { enumerable: true, get: function() {
      return connection_1.ProgressType;
    } });
    Object.defineProperty(exports2, "Trace", { enumerable: true, get: function() {
      return connection_1.Trace;
    } });
    Object.defineProperty(exports2, "TraceFormat", { enumerable: true, get: function() {
      return connection_1.TraceFormat;
    } });
    Object.defineProperty(exports2, "SetTraceNotification", { enumerable: true, get: function() {
      return connection_1.SetTraceNotification;
    } });
    Object.defineProperty(exports2, "LogTraceNotification", { enumerable: true, get: function() {
      return connection_1.LogTraceNotification;
    } });
    Object.defineProperty(exports2, "ConnectionErrors", { enumerable: true, get: function() {
      return connection_1.ConnectionErrors;
    } });
    Object.defineProperty(exports2, "ConnectionError", { enumerable: true, get: function() {
      return connection_1.ConnectionError;
    } });
    Object.defineProperty(exports2, "CancellationReceiverStrategy", { enumerable: true, get: function() {
      return connection_1.CancellationReceiverStrategy;
    } });
    Object.defineProperty(exports2, "CancellationSenderStrategy", { enumerable: true, get: function() {
      return connection_1.CancellationSenderStrategy;
    } });
    Object.defineProperty(exports2, "CancellationStrategy", { enumerable: true, get: function() {
      return connection_1.CancellationStrategy;
    } });
    var ral_1 = require_ral();
    exports2.RAL = ral_1.default;
  }
});

// node_modules/vscode-jsonrpc/lib/node/main.js
var require_main = __commonJS({
  "node_modules/vscode-jsonrpc/lib/node/main.js"(exports2) {
    "use strict";
    var __createBinding = exports2 && exports2.__createBinding || (Object.create ? function(o, m, k, k2) {
      if (k2 === void 0) k2 = k;
      Object.defineProperty(o, k2, { enumerable: true, get: function() {
        return m[k];
      } });
    } : function(o, m, k, k2) {
      if (k2 === void 0) k2 = k;
      o[k2] = m[k];
    });
    var __exportStar = exports2 && exports2.__exportStar || function(m, exports3) {
      for (var p in m) if (p !== "default" && !Object.prototype.hasOwnProperty.call(exports3, p)) __createBinding(exports3, m, p);
    };
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.createMessageConnection = exports2.createServerSocketTransport = exports2.createClientSocketTransport = exports2.createServerPipeTransport = exports2.createClientPipeTransport = exports2.generateRandomPipeName = exports2.StreamMessageWriter = exports2.StreamMessageReader = exports2.SocketMessageWriter = exports2.SocketMessageReader = exports2.IPCMessageWriter = exports2.IPCMessageReader = void 0;
    var ril_1 = require_ril();
    ril_1.default.install();
    var api_1 = require_api();
    var path2 = require("path");
    var os = require("os");
    var crypto_1 = require("crypto");
    var net_1 = require("net");
    __exportStar(require_api(), exports2);
    var IPCMessageReader = class extends api_1.AbstractMessageReader {
      constructor(process2) {
        super();
        this.process = process2;
        let eventEmitter = this.process;
        eventEmitter.on("error", (error) => this.fireError(error));
        eventEmitter.on("close", () => this.fireClose());
      }
      listen(callback) {
        this.process.on("message", callback);
        return api_1.Disposable.create(() => this.process.off("message", callback));
      }
    };
    exports2.IPCMessageReader = IPCMessageReader;
    var IPCMessageWriter = class extends api_1.AbstractMessageWriter {
      constructor(process2) {
        super();
        this.process = process2;
        this.errorCount = 0;
        let eventEmitter = this.process;
        eventEmitter.on("error", (error) => this.fireError(error));
        eventEmitter.on("close", () => this.fireClose);
      }
      write(msg) {
        try {
          if (typeof this.process.send === "function") {
            this.process.send(msg, void 0, void 0, (error) => {
              if (error) {
                this.errorCount++;
                this.handleError(error, msg);
              } else {
                this.errorCount = 0;
              }
            });
          }
          return Promise.resolve();
        } catch (error) {
          this.handleError(error, msg);
          return Promise.reject(error);
        }
      }
      handleError(error, msg) {
        this.errorCount++;
        this.fireError(error, msg, this.errorCount);
      }
      end() {
      }
    };
    exports2.IPCMessageWriter = IPCMessageWriter;
    var SocketMessageReader = class extends api_1.ReadableStreamMessageReader {
      constructor(socket, encoding = "utf-8") {
        super(ril_1.default().stream.asReadableStream(socket), encoding);
      }
    };
    exports2.SocketMessageReader = SocketMessageReader;
    var SocketMessageWriter = class extends api_1.WriteableStreamMessageWriter {
      constructor(socket, options) {
        super(ril_1.default().stream.asWritableStream(socket), options);
        this.socket = socket;
      }
      dispose() {
        super.dispose();
        this.socket.destroy();
      }
    };
    exports2.SocketMessageWriter = SocketMessageWriter;
    var StreamMessageReader = class extends api_1.ReadableStreamMessageReader {
      constructor(readble, encoding) {
        super(ril_1.default().stream.asReadableStream(readble), encoding);
      }
    };
    exports2.StreamMessageReader = StreamMessageReader;
    var StreamMessageWriter = class extends api_1.WriteableStreamMessageWriter {
      constructor(writable, options) {
        super(ril_1.default().stream.asWritableStream(writable), options);
      }
    };
    exports2.StreamMessageWriter = StreamMessageWriter;
    var XDG_RUNTIME_DIR = process.env["XDG_RUNTIME_DIR"];
    var safeIpcPathLengths = /* @__PURE__ */ new Map([
      ["linux", 107],
      ["darwin", 103]
    ]);
    function generateRandomPipeName() {
      const randomSuffix = crypto_1.randomBytes(21).toString("hex");
      if (process.platform === "win32") {
        return `\\\\.\\pipe\\vscode-jsonrpc-${randomSuffix}-sock`;
      }
      let result;
      if (XDG_RUNTIME_DIR) {
        result = path2.join(XDG_RUNTIME_DIR, `vscode-ipc-${randomSuffix}.sock`);
      } else {
        result = path2.join(os.tmpdir(), `vscode-${randomSuffix}.sock`);
      }
      const limit = safeIpcPathLengths.get(process.platform);
      if (limit !== void 0 && result.length >= limit) {
        ril_1.default().console.warn(`WARNING: IPC handle "${result}" is longer than ${limit} characters.`);
      }
      return result;
    }
    exports2.generateRandomPipeName = generateRandomPipeName;
    function createClientPipeTransport(pipeName, encoding = "utf-8") {
      let connectResolve;
      const connected = new Promise((resolve, _reject) => {
        connectResolve = resolve;
      });
      return new Promise((resolve, reject) => {
        let server = net_1.createServer((socket) => {
          server.close();
          connectResolve([
            new SocketMessageReader(socket, encoding),
            new SocketMessageWriter(socket, encoding)
          ]);
        });
        server.on("error", reject);
        server.listen(pipeName, () => {
          server.removeListener("error", reject);
          resolve({
            onConnected: () => {
              return connected;
            }
          });
        });
      });
    }
    exports2.createClientPipeTransport = createClientPipeTransport;
    function createServerPipeTransport(pipeName, encoding = "utf-8") {
      const socket = net_1.createConnection(pipeName);
      return [
        new SocketMessageReader(socket, encoding),
        new SocketMessageWriter(socket, encoding)
      ];
    }
    exports2.createServerPipeTransport = createServerPipeTransport;
    function createClientSocketTransport(port, encoding = "utf-8") {
      let connectResolve;
      const connected = new Promise((resolve, _reject) => {
        connectResolve = resolve;
      });
      return new Promise((resolve, reject) => {
        const server = net_1.createServer((socket) => {
          server.close();
          connectResolve([
            new SocketMessageReader(socket, encoding),
            new SocketMessageWriter(socket, encoding)
          ]);
        });
        server.on("error", reject);
        server.listen(port, "127.0.0.1", () => {
          server.removeListener("error", reject);
          resolve({
            onConnected: () => {
              return connected;
            }
          });
        });
      });
    }
    exports2.createClientSocketTransport = createClientSocketTransport;
    function createServerSocketTransport(port, encoding = "utf-8") {
      const socket = net_1.createConnection(port, "127.0.0.1");
      return [
        new SocketMessageReader(socket, encoding),
        new SocketMessageWriter(socket, encoding)
      ];
    }
    exports2.createServerSocketTransport = createServerSocketTransport;
    function isReadableStream(value) {
      const candidate = value;
      return candidate.read !== void 0 && candidate.addListener !== void 0;
    }
    function isWritableStream(value) {
      const candidate = value;
      return candidate.write !== void 0 && candidate.addListener !== void 0;
    }
    function createMessageConnection(input, output, logger, options) {
      if (!logger) {
        logger = api_1.NullLogger;
      }
      const reader = isReadableStream(input) ? new StreamMessageReader(input) : input;
      const writer = isWritableStream(output) ? new StreamMessageWriter(output) : output;
      if (api_1.ConnectionStrategy.is(options)) {
        options = { connectionStrategy: options };
      }
      return api_1.createMessageConnection(reader, writer, logger, options);
    }
    exports2.createMessageConnection = createMessageConnection;
  }
});

// node_modules/vscode-jsonrpc/node.js
var require_node = __commonJS({
  "node_modules/vscode-jsonrpc/node.js"(exports2, module2) {
    "use strict";
    module2.exports = require_main();
  }
});

// node_modules/vscode-languageserver-types/lib/umd/main.js
var require_main2 = __commonJS({
  "node_modules/vscode-languageserver-types/lib/umd/main.js"(exports2, module2) {
    (function(factory) {
      if (typeof module2 === "object" && typeof module2.exports === "object") {
        var v = factory(require, exports2);
        if (v !== void 0) module2.exports = v;
      } else if (typeof define === "function" && define.amd) {
        define(["require", "exports"], factory);
      }
    })(function(require2, exports3) {
      "use strict";
      Object.defineProperty(exports3, "__esModule", { value: true });
      exports3.TextDocument = exports3.EOL = exports3.SelectionRange = exports3.DocumentLink = exports3.FormattingOptions = exports3.CodeLens = exports3.CodeAction = exports3.CodeActionContext = exports3.CodeActionKind = exports3.DocumentSymbol = exports3.SymbolInformation = exports3.SymbolTag = exports3.SymbolKind = exports3.DocumentHighlight = exports3.DocumentHighlightKind = exports3.SignatureInformation = exports3.ParameterInformation = exports3.Hover = exports3.MarkedString = exports3.CompletionList = exports3.CompletionItem = exports3.InsertTextMode = exports3.InsertReplaceEdit = exports3.CompletionItemTag = exports3.InsertTextFormat = exports3.CompletionItemKind = exports3.MarkupContent = exports3.MarkupKind = exports3.TextDocumentItem = exports3.OptionalVersionedTextDocumentIdentifier = exports3.VersionedTextDocumentIdentifier = exports3.TextDocumentIdentifier = exports3.WorkspaceChange = exports3.WorkspaceEdit = exports3.DeleteFile = exports3.RenameFile = exports3.CreateFile = exports3.TextDocumentEdit = exports3.AnnotatedTextEdit = exports3.ChangeAnnotationIdentifier = exports3.ChangeAnnotation = exports3.TextEdit = exports3.Command = exports3.Diagnostic = exports3.CodeDescription = exports3.DiagnosticTag = exports3.DiagnosticSeverity = exports3.DiagnosticRelatedInformation = exports3.FoldingRange = exports3.FoldingRangeKind = exports3.ColorPresentation = exports3.ColorInformation = exports3.Color = exports3.LocationLink = exports3.Location = exports3.Range = exports3.Position = exports3.uinteger = exports3.integer = void 0;
      var integer;
      (function(integer2) {
        integer2.MIN_VALUE = -2147483648;
        integer2.MAX_VALUE = 2147483647;
      })(integer = exports3.integer || (exports3.integer = {}));
      var uinteger;
      (function(uinteger2) {
        uinteger2.MIN_VALUE = 0;
        uinteger2.MAX_VALUE = 2147483647;
      })(uinteger = exports3.uinteger || (exports3.uinteger = {}));
      var Position;
      (function(Position2) {
        function create(line, character) {
          if (line === Number.MAX_VALUE) {
            line = uinteger.MAX_VALUE;
          }
          if (character === Number.MAX_VALUE) {
            character = uinteger.MAX_VALUE;
          }
          return { line, character };
        }
        Position2.create = create;
        function is(value) {
          var candidate = value;
          return Is.objectLiteral(candidate) && Is.uinteger(candidate.line) && Is.uinteger(candidate.character);
        }
        Position2.is = is;
      })(Position = exports3.Position || (exports3.Position = {}));
      var Range;
      (function(Range2) {
        function create(one, two, three, four) {
          if (Is.uinteger(one) && Is.uinteger(two) && Is.uinteger(three) && Is.uinteger(four)) {
            return { start: Position.create(one, two), end: Position.create(three, four) };
          } else if (Position.is(one) && Position.is(two)) {
            return { start: one, end: two };
          } else {
            throw new Error("Range#create called with invalid arguments[" + one + ", " + two + ", " + three + ", " + four + "]");
          }
        }
        Range2.create = create;
        function is(value) {
          var candidate = value;
          return Is.objectLiteral(candidate) && Position.is(candidate.start) && Position.is(candidate.end);
        }
        Range2.is = is;
      })(Range = exports3.Range || (exports3.Range = {}));
      var Location;
      (function(Location2) {
        function create(uri, range) {
          return { uri, range };
        }
        Location2.create = create;
        function is(value) {
          var candidate = value;
          return Is.defined(candidate) && Range.is(candidate.range) && (Is.string(candidate.uri) || Is.undefined(candidate.uri));
        }
        Location2.is = is;
      })(Location = exports3.Location || (exports3.Location = {}));
      var LocationLink;
      (function(LocationLink2) {
        function create(targetUri, targetRange, targetSelectionRange, originSelectionRange) {
          return { targetUri, targetRange, targetSelectionRange, originSelectionRange };
        }
        LocationLink2.create = create;
        function is(value) {
          var candidate = value;
          return Is.defined(candidate) && Range.is(candidate.targetRange) && Is.string(candidate.targetUri) && (Range.is(candidate.targetSelectionRange) || Is.undefined(candidate.targetSelectionRange)) && (Range.is(candidate.originSelectionRange) || Is.undefined(candidate.originSelectionRange));
        }
        LocationLink2.is = is;
      })(LocationLink = exports3.LocationLink || (exports3.LocationLink = {}));
      var Color;
      (function(Color2) {
        function create(red, green, blue, alpha) {
          return {
            red,
            green,
            blue,
            alpha
          };
        }
        Color2.create = create;
        function is(value) {
          var candidate = value;
          return Is.numberRange(candidate.red, 0, 1) && Is.numberRange(candidate.green, 0, 1) && Is.numberRange(candidate.blue, 0, 1) && Is.numberRange(candidate.alpha, 0, 1);
        }
        Color2.is = is;
      })(Color = exports3.Color || (exports3.Color = {}));
      var ColorInformation;
      (function(ColorInformation2) {
        function create(range, color) {
          return {
            range,
            color
          };
        }
        ColorInformation2.create = create;
        function is(value) {
          var candidate = value;
          return Range.is(candidate.range) && Color.is(candidate.color);
        }
        ColorInformation2.is = is;
      })(ColorInformation = exports3.ColorInformation || (exports3.ColorInformation = {}));
      var ColorPresentation;
      (function(ColorPresentation2) {
        function create(label, textEdit, additionalTextEdits) {
          return {
            label,
            textEdit,
            additionalTextEdits
          };
        }
        ColorPresentation2.create = create;
        function is(value) {
          var candidate = value;
          return Is.string(candidate.label) && (Is.undefined(candidate.textEdit) || TextEdit.is(candidate)) && (Is.undefined(candidate.additionalTextEdits) || Is.typedArray(candidate.additionalTextEdits, TextEdit.is));
        }
        ColorPresentation2.is = is;
      })(ColorPresentation = exports3.ColorPresentation || (exports3.ColorPresentation = {}));
      var FoldingRangeKind;
      (function(FoldingRangeKind2) {
        FoldingRangeKind2["Comment"] = "comment";
        FoldingRangeKind2["Imports"] = "imports";
        FoldingRangeKind2["Region"] = "region";
      })(FoldingRangeKind = exports3.FoldingRangeKind || (exports3.FoldingRangeKind = {}));
      var FoldingRange;
      (function(FoldingRange2) {
        function create(startLine, endLine, startCharacter, endCharacter, kind) {
          var result = {
            startLine,
            endLine
          };
          if (Is.defined(startCharacter)) {
            result.startCharacter = startCharacter;
          }
          if (Is.defined(endCharacter)) {
            result.endCharacter = endCharacter;
          }
          if (Is.defined(kind)) {
            result.kind = kind;
          }
          return result;
        }
        FoldingRange2.create = create;
        function is(value) {
          var candidate = value;
          return Is.uinteger(candidate.startLine) && Is.uinteger(candidate.startLine) && (Is.undefined(candidate.startCharacter) || Is.uinteger(candidate.startCharacter)) && (Is.undefined(candidate.endCharacter) || Is.uinteger(candidate.endCharacter)) && (Is.undefined(candidate.kind) || Is.string(candidate.kind));
        }
        FoldingRange2.is = is;
      })(FoldingRange = exports3.FoldingRange || (exports3.FoldingRange = {}));
      var DiagnosticRelatedInformation;
      (function(DiagnosticRelatedInformation2) {
        function create(location, message) {
          return {
            location,
            message
          };
        }
        DiagnosticRelatedInformation2.create = create;
        function is(value) {
          var candidate = value;
          return Is.defined(candidate) && Location.is(candidate.location) && Is.string(candidate.message);
        }
        DiagnosticRelatedInformation2.is = is;
      })(DiagnosticRelatedInformation = exports3.DiagnosticRelatedInformation || (exports3.DiagnosticRelatedInformation = {}));
      var DiagnosticSeverity;
      (function(DiagnosticSeverity2) {
        DiagnosticSeverity2.Error = 1;
        DiagnosticSeverity2.Warning = 2;
        DiagnosticSeverity2.Information = 3;
        DiagnosticSeverity2.Hint = 4;
      })(DiagnosticSeverity = exports3.DiagnosticSeverity || (exports3.DiagnosticSeverity = {}));
      var DiagnosticTag;
      (function(DiagnosticTag2) {
        DiagnosticTag2.Unnecessary = 1;
        DiagnosticTag2.Deprecated = 2;
      })(DiagnosticTag = exports3.DiagnosticTag || (exports3.DiagnosticTag = {}));
      var CodeDescription;
      (function(CodeDescription2) {
        function is(value) {
          var candidate = value;
          return candidate !== void 0 && candidate !== null && Is.string(candidate.href);
        }
        CodeDescription2.is = is;
      })(CodeDescription = exports3.CodeDescription || (exports3.CodeDescription = {}));
      var Diagnostic;
      (function(Diagnostic2) {
        function create(range, message, severity, code, source, relatedInformation) {
          var result = { range, message };
          if (Is.defined(severity)) {
            result.severity = severity;
          }
          if (Is.defined(code)) {
            result.code = code;
          }
          if (Is.defined(source)) {
            result.source = source;
          }
          if (Is.defined(relatedInformation)) {
            result.relatedInformation = relatedInformation;
          }
          return result;
        }
        Diagnostic2.create = create;
        function is(value) {
          var _a;
          var candidate = value;
          return Is.defined(candidate) && Range.is(candidate.range) && Is.string(candidate.message) && (Is.number(candidate.severity) || Is.undefined(candidate.severity)) && (Is.integer(candidate.code) || Is.string(candidate.code) || Is.undefined(candidate.code)) && (Is.undefined(candidate.codeDescription) || Is.string((_a = candidate.codeDescription) === null || _a === void 0 ? void 0 : _a.href)) && (Is.string(candidate.source) || Is.undefined(candidate.source)) && (Is.undefined(candidate.relatedInformation) || Is.typedArray(candidate.relatedInformation, DiagnosticRelatedInformation.is));
        }
        Diagnostic2.is = is;
      })(Diagnostic = exports3.Diagnostic || (exports3.Diagnostic = {}));
      var Command;
      (function(Command2) {
        function create(title, command) {
          var args = [];
          for (var _i = 2; _i < arguments.length; _i++) {
            args[_i - 2] = arguments[_i];
          }
          var result = { title, command };
          if (Is.defined(args) && args.length > 0) {
            result.arguments = args;
          }
          return result;
        }
        Command2.create = create;
        function is(value) {
          var candidate = value;
          return Is.defined(candidate) && Is.string(candidate.title) && Is.string(candidate.command);
        }
        Command2.is = is;
      })(Command = exports3.Command || (exports3.Command = {}));
      var TextEdit;
      (function(TextEdit2) {
        function replace(range, newText) {
          return { range, newText };
        }
        TextEdit2.replace = replace;
        function insert(position, newText) {
          return { range: { start: position, end: position }, newText };
        }
        TextEdit2.insert = insert;
        function del(range) {
          return { range, newText: "" };
        }
        TextEdit2.del = del;
        function is(value) {
          var candidate = value;
          return Is.objectLiteral(candidate) && Is.string(candidate.newText) && Range.is(candidate.range);
        }
        TextEdit2.is = is;
      })(TextEdit = exports3.TextEdit || (exports3.TextEdit = {}));
      var ChangeAnnotation;
      (function(ChangeAnnotation2) {
        function create(label, needsConfirmation, description) {
          var result = { label };
          if (needsConfirmation !== void 0) {
            result.needsConfirmation = needsConfirmation;
          }
          if (description !== void 0) {
            result.description = description;
          }
          return result;
        }
        ChangeAnnotation2.create = create;
        function is(value) {
          var candidate = value;
          return candidate !== void 0 && Is.objectLiteral(candidate) && Is.string(candidate.label) && (Is.boolean(candidate.needsConfirmation) || candidate.needsConfirmation === void 0) && (Is.string(candidate.description) || candidate.description === void 0);
        }
        ChangeAnnotation2.is = is;
      })(ChangeAnnotation = exports3.ChangeAnnotation || (exports3.ChangeAnnotation = {}));
      var ChangeAnnotationIdentifier;
      (function(ChangeAnnotationIdentifier2) {
        function is(value) {
          var candidate = value;
          return typeof candidate === "string";
        }
        ChangeAnnotationIdentifier2.is = is;
      })(ChangeAnnotationIdentifier = exports3.ChangeAnnotationIdentifier || (exports3.ChangeAnnotationIdentifier = {}));
      var AnnotatedTextEdit;
      (function(AnnotatedTextEdit2) {
        function replace(range, newText, annotation) {
          return { range, newText, annotationId: annotation };
        }
        AnnotatedTextEdit2.replace = replace;
        function insert(position, newText, annotation) {
          return { range: { start: position, end: position }, newText, annotationId: annotation };
        }
        AnnotatedTextEdit2.insert = insert;
        function del(range, annotation) {
          return { range, newText: "", annotationId: annotation };
        }
        AnnotatedTextEdit2.del = del;
        function is(value) {
          var candidate = value;
          return TextEdit.is(candidate) && (ChangeAnnotation.is(candidate.annotationId) || ChangeAnnotationIdentifier.is(candidate.annotationId));
        }
        AnnotatedTextEdit2.is = is;
      })(AnnotatedTextEdit = exports3.AnnotatedTextEdit || (exports3.AnnotatedTextEdit = {}));
      var TextDocumentEdit;
      (function(TextDocumentEdit2) {
        function create(textDocument, edits) {
          return { textDocument, edits };
        }
        TextDocumentEdit2.create = create;
        function is(value) {
          var candidate = value;
          return Is.defined(candidate) && OptionalVersionedTextDocumentIdentifier.is(candidate.textDocument) && Array.isArray(candidate.edits);
        }
        TextDocumentEdit2.is = is;
      })(TextDocumentEdit = exports3.TextDocumentEdit || (exports3.TextDocumentEdit = {}));
      var CreateFile;
      (function(CreateFile2) {
        function create(uri, options, annotation) {
          var result = {
            kind: "create",
            uri
          };
          if (options !== void 0 && (options.overwrite !== void 0 || options.ignoreIfExists !== void 0)) {
            result.options = options;
          }
          if (annotation !== void 0) {
            result.annotationId = annotation;
          }
          return result;
        }
        CreateFile2.create = create;
        function is(value) {
          var candidate = value;
          return candidate && candidate.kind === "create" && Is.string(candidate.uri) && (candidate.options === void 0 || (candidate.options.overwrite === void 0 || Is.boolean(candidate.options.overwrite)) && (candidate.options.ignoreIfExists === void 0 || Is.boolean(candidate.options.ignoreIfExists))) && (candidate.annotationId === void 0 || ChangeAnnotationIdentifier.is(candidate.annotationId));
        }
        CreateFile2.is = is;
      })(CreateFile = exports3.CreateFile || (exports3.CreateFile = {}));
      var RenameFile;
      (function(RenameFile2) {
        function create(oldUri, newUri, options, annotation) {
          var result = {
            kind: "rename",
            oldUri,
            newUri
          };
          if (options !== void 0 && (options.overwrite !== void 0 || options.ignoreIfExists !== void 0)) {
            result.options = options;
          }
          if (annotation !== void 0) {
            result.annotationId = annotation;
          }
          return result;
        }
        RenameFile2.create = create;
        function is(value) {
          var candidate = value;
          return candidate && candidate.kind === "rename" && Is.string(candidate.oldUri) && Is.string(candidate.newUri) && (candidate.options === void 0 || (candidate.options.overwrite === void 0 || Is.boolean(candidate.options.overwrite)) && (candidate.options.ignoreIfExists === void 0 || Is.boolean(candidate.options.ignoreIfExists))) && (candidate.annotationId === void 0 || ChangeAnnotationIdentifier.is(candidate.annotationId));
        }
        RenameFile2.is = is;
      })(RenameFile = exports3.RenameFile || (exports3.RenameFile = {}));
      var DeleteFile;
      (function(DeleteFile2) {
        function create(uri, options, annotation) {
          var result = {
            kind: "delete",
            uri
          };
          if (options !== void 0 && (options.recursive !== void 0 || options.ignoreIfNotExists !== void 0)) {
            result.options = options;
          }
          if (annotation !== void 0) {
            result.annotationId = annotation;
          }
          return result;
        }
        DeleteFile2.create = create;
        function is(value) {
          var candidate = value;
          return candidate && candidate.kind === "delete" && Is.string(candidate.uri) && (candidate.options === void 0 || (candidate.options.recursive === void 0 || Is.boolean(candidate.options.recursive)) && (candidate.options.ignoreIfNotExists === void 0 || Is.boolean(candidate.options.ignoreIfNotExists))) && (candidate.annotationId === void 0 || ChangeAnnotationIdentifier.is(candidate.annotationId));
        }
        DeleteFile2.is = is;
      })(DeleteFile = exports3.DeleteFile || (exports3.DeleteFile = {}));
      var WorkspaceEdit;
      (function(WorkspaceEdit2) {
        function is(value) {
          var candidate = value;
          return candidate && (candidate.changes !== void 0 || candidate.documentChanges !== void 0) && (candidate.documentChanges === void 0 || candidate.documentChanges.every(function(change) {
            if (Is.string(change.kind)) {
              return CreateFile.is(change) || RenameFile.is(change) || DeleteFile.is(change);
            } else {
              return TextDocumentEdit.is(change);
            }
          }));
        }
        WorkspaceEdit2.is = is;
      })(WorkspaceEdit = exports3.WorkspaceEdit || (exports3.WorkspaceEdit = {}));
      var TextEditChangeImpl = (
        /** @class */
        function() {
          function TextEditChangeImpl2(edits, changeAnnotations) {
            this.edits = edits;
            this.changeAnnotations = changeAnnotations;
          }
          TextEditChangeImpl2.prototype.insert = function(position, newText, annotation) {
            var edit;
            var id;
            if (annotation === void 0) {
              edit = TextEdit.insert(position, newText);
            } else if (ChangeAnnotationIdentifier.is(annotation)) {
              id = annotation;
              edit = AnnotatedTextEdit.insert(position, newText, annotation);
            } else {
              this.assertChangeAnnotations(this.changeAnnotations);
              id = this.changeAnnotations.manage(annotation);
              edit = AnnotatedTextEdit.insert(position, newText, id);
            }
            this.edits.push(edit);
            if (id !== void 0) {
              return id;
            }
          };
          TextEditChangeImpl2.prototype.replace = function(range, newText, annotation) {
            var edit;
            var id;
            if (annotation === void 0) {
              edit = TextEdit.replace(range, newText);
            } else if (ChangeAnnotationIdentifier.is(annotation)) {
              id = annotation;
              edit = AnnotatedTextEdit.replace(range, newText, annotation);
            } else {
              this.assertChangeAnnotations(this.changeAnnotations);
              id = this.changeAnnotations.manage(annotation);
              edit = AnnotatedTextEdit.replace(range, newText, id);
            }
            this.edits.push(edit);
            if (id !== void 0) {
              return id;
            }
          };
          TextEditChangeImpl2.prototype.delete = function(range, annotation) {
            var edit;
            var id;
            if (annotation === void 0) {
              edit = TextEdit.del(range);
            } else if (ChangeAnnotationIdentifier.is(annotation)) {
              id = annotation;
              edit = AnnotatedTextEdit.del(range, annotation);
            } else {
              this.assertChangeAnnotations(this.changeAnnotations);
              id = this.changeAnnotations.manage(annotation);
              edit = AnnotatedTextEdit.del(range, id);
            }
            this.edits.push(edit);
            if (id !== void 0) {
              return id;
            }
          };
          TextEditChangeImpl2.prototype.add = function(edit) {
            this.edits.push(edit);
          };
          TextEditChangeImpl2.prototype.all = function() {
            return this.edits;
          };
          TextEditChangeImpl2.prototype.clear = function() {
            this.edits.splice(0, this.edits.length);
          };
          TextEditChangeImpl2.prototype.assertChangeAnnotations = function(value) {
            if (value === void 0) {
              throw new Error("Text edit change is not configured to manage change annotations.");
            }
          };
          return TextEditChangeImpl2;
        }()
      );
      var ChangeAnnotations = (
        /** @class */
        function() {
          function ChangeAnnotations2(annotations) {
            this._annotations = annotations === void 0 ? /* @__PURE__ */ Object.create(null) : annotations;
            this._counter = 0;
            this._size = 0;
          }
          ChangeAnnotations2.prototype.all = function() {
            return this._annotations;
          };
          Object.defineProperty(ChangeAnnotations2.prototype, "size", {
            get: function() {
              return this._size;
            },
            enumerable: false,
            configurable: true
          });
          ChangeAnnotations2.prototype.manage = function(idOrAnnotation, annotation) {
            var id;
            if (ChangeAnnotationIdentifier.is(idOrAnnotation)) {
              id = idOrAnnotation;
            } else {
              id = this.nextId();
              annotation = idOrAnnotation;
            }
            if (this._annotations[id] !== void 0) {
              throw new Error("Id " + id + " is already in use.");
            }
            if (annotation === void 0) {
              throw new Error("No annotation provided for id " + id);
            }
            this._annotations[id] = annotation;
            this._size++;
            return id;
          };
          ChangeAnnotations2.prototype.nextId = function() {
            this._counter++;
            return this._counter.toString();
          };
          return ChangeAnnotations2;
        }()
      );
      var WorkspaceChange = (
        /** @class */
        function() {
          function WorkspaceChange2(workspaceEdit) {
            var _this = this;
            this._textEditChanges = /* @__PURE__ */ Object.create(null);
            if (workspaceEdit !== void 0) {
              this._workspaceEdit = workspaceEdit;
              if (workspaceEdit.documentChanges) {
                this._changeAnnotations = new ChangeAnnotations(workspaceEdit.changeAnnotations);
                workspaceEdit.changeAnnotations = this._changeAnnotations.all();
                workspaceEdit.documentChanges.forEach(function(change) {
                  if (TextDocumentEdit.is(change)) {
                    var textEditChange = new TextEditChangeImpl(change.edits, _this._changeAnnotations);
                    _this._textEditChanges[change.textDocument.uri] = textEditChange;
                  }
                });
              } else if (workspaceEdit.changes) {
                Object.keys(workspaceEdit.changes).forEach(function(key) {
                  var textEditChange = new TextEditChangeImpl(workspaceEdit.changes[key]);
                  _this._textEditChanges[key] = textEditChange;
                });
              }
            } else {
              this._workspaceEdit = {};
            }
          }
          Object.defineProperty(WorkspaceChange2.prototype, "edit", {
            /**
             * Returns the underlying [WorkspaceEdit](#WorkspaceEdit) literal
             * use to be returned from a workspace edit operation like rename.
             */
            get: function() {
              this.initDocumentChanges();
              if (this._changeAnnotations !== void 0) {
                if (this._changeAnnotations.size === 0) {
                  this._workspaceEdit.changeAnnotations = void 0;
                } else {
                  this._workspaceEdit.changeAnnotations = this._changeAnnotations.all();
                }
              }
              return this._workspaceEdit;
            },
            enumerable: false,
            configurable: true
          });
          WorkspaceChange2.prototype.getTextEditChange = function(key) {
            if (OptionalVersionedTextDocumentIdentifier.is(key)) {
              this.initDocumentChanges();
              if (this._workspaceEdit.documentChanges === void 0) {
                throw new Error("Workspace edit is not configured for document changes.");
              }
              var textDocument = { uri: key.uri, version: key.version };
              var result = this._textEditChanges[textDocument.uri];
              if (!result) {
                var edits = [];
                var textDocumentEdit = {
                  textDocument,
                  edits
                };
                this._workspaceEdit.documentChanges.push(textDocumentEdit);
                result = new TextEditChangeImpl(edits, this._changeAnnotations);
                this._textEditChanges[textDocument.uri] = result;
              }
              return result;
            } else {
              this.initChanges();
              if (this._workspaceEdit.changes === void 0) {
                throw new Error("Workspace edit is not configured for normal text edit changes.");
              }
              var result = this._textEditChanges[key];
              if (!result) {
                var edits = [];
                this._workspaceEdit.changes[key] = edits;
                result = new TextEditChangeImpl(edits);
                this._textEditChanges[key] = result;
              }
              return result;
            }
          };
          WorkspaceChange2.prototype.initDocumentChanges = function() {
            if (this._workspaceEdit.documentChanges === void 0 && this._workspaceEdit.changes === void 0) {
              this._changeAnnotations = new ChangeAnnotations();
              this._workspaceEdit.documentChanges = [];
              this._workspaceEdit.changeAnnotations = this._changeAnnotations.all();
            }
          };
          WorkspaceChange2.prototype.initChanges = function() {
            if (this._workspaceEdit.documentChanges === void 0 && this._workspaceEdit.changes === void 0) {
              this._workspaceEdit.changes = /* @__PURE__ */ Object.create(null);
            }
          };
          WorkspaceChange2.prototype.createFile = function(uri, optionsOrAnnotation, options) {
            this.initDocumentChanges();
            if (this._workspaceEdit.documentChanges === void 0) {
              throw new Error("Workspace edit is not configured for document changes.");
            }
            var annotation;
            if (ChangeAnnotation.is(optionsOrAnnotation) || ChangeAnnotationIdentifier.is(optionsOrAnnotation)) {
              annotation = optionsOrAnnotation;
            } else {
              options = optionsOrAnnotation;
            }
            var operation;
            var id;
            if (annotation === void 0) {
              operation = CreateFile.create(uri, options);
            } else {
              id = ChangeAnnotationIdentifier.is(annotation) ? annotation : this._changeAnnotations.manage(annotation);
              operation = CreateFile.create(uri, options, id);
            }
            this._workspaceEdit.documentChanges.push(operation);
            if (id !== void 0) {
              return id;
            }
          };
          WorkspaceChange2.prototype.renameFile = function(oldUri, newUri, optionsOrAnnotation, options) {
            this.initDocumentChanges();
            if (this._workspaceEdit.documentChanges === void 0) {
              throw new Error("Workspace edit is not configured for document changes.");
            }
            var annotation;
            if (ChangeAnnotation.is(optionsOrAnnotation) || ChangeAnnotationIdentifier.is(optionsOrAnnotation)) {
              annotation = optionsOrAnnotation;
            } else {
              options = optionsOrAnnotation;
            }
            var operation;
            var id;
            if (annotation === void 0) {
              operation = RenameFile.create(oldUri, newUri, options);
            } else {
              id = ChangeAnnotationIdentifier.is(annotation) ? annotation : this._changeAnnotations.manage(annotation);
              operation = RenameFile.create(oldUri, newUri, options, id);
            }
            this._workspaceEdit.documentChanges.push(operation);
            if (id !== void 0) {
              return id;
            }
          };
          WorkspaceChange2.prototype.deleteFile = function(uri, optionsOrAnnotation, options) {
            this.initDocumentChanges();
            if (this._workspaceEdit.documentChanges === void 0) {
              throw new Error("Workspace edit is not configured for document changes.");
            }
            var annotation;
            if (ChangeAnnotation.is(optionsOrAnnotation) || ChangeAnnotationIdentifier.is(optionsOrAnnotation)) {
              annotation = optionsOrAnnotation;
            } else {
              options = optionsOrAnnotation;
            }
            var operation;
            var id;
            if (annotation === void 0) {
              operation = DeleteFile.create(uri, options);
            } else {
              id = ChangeAnnotationIdentifier.is(annotation) ? annotation : this._changeAnnotations.manage(annotation);
              operation = DeleteFile.create(uri, options, id);
            }
            this._workspaceEdit.documentChanges.push(operation);
            if (id !== void 0) {
              return id;
            }
          };
          return WorkspaceChange2;
        }()
      );
      exports3.WorkspaceChange = WorkspaceChange;
      var TextDocumentIdentifier;
      (function(TextDocumentIdentifier2) {
        function create(uri) {
          return { uri };
        }
        TextDocumentIdentifier2.create = create;
        function is(value) {
          var candidate = value;
          return Is.defined(candidate) && Is.string(candidate.uri);
        }
        TextDocumentIdentifier2.is = is;
      })(TextDocumentIdentifier = exports3.TextDocumentIdentifier || (exports3.TextDocumentIdentifier = {}));
      var VersionedTextDocumentIdentifier;
      (function(VersionedTextDocumentIdentifier2) {
        function create(uri, version) {
          return { uri, version };
        }
        VersionedTextDocumentIdentifier2.create = create;
        function is(value) {
          var candidate = value;
          return Is.defined(candidate) && Is.string(candidate.uri) && Is.integer(candidate.version);
        }
        VersionedTextDocumentIdentifier2.is = is;
      })(VersionedTextDocumentIdentifier = exports3.VersionedTextDocumentIdentifier || (exports3.VersionedTextDocumentIdentifier = {}));
      var OptionalVersionedTextDocumentIdentifier;
      (function(OptionalVersionedTextDocumentIdentifier2) {
        function create(uri, version) {
          return { uri, version };
        }
        OptionalVersionedTextDocumentIdentifier2.create = create;
        function is(value) {
          var candidate = value;
          return Is.defined(candidate) && Is.string(candidate.uri) && (candidate.version === null || Is.integer(candidate.version));
        }
        OptionalVersionedTextDocumentIdentifier2.is = is;
      })(OptionalVersionedTextDocumentIdentifier = exports3.OptionalVersionedTextDocumentIdentifier || (exports3.OptionalVersionedTextDocumentIdentifier = {}));
      var TextDocumentItem;
      (function(TextDocumentItem2) {
        function create(uri, languageId, version, text) {
          return { uri, languageId, version, text };
        }
        TextDocumentItem2.create = create;
        function is(value) {
          var candidate = value;
          return Is.defined(candidate) && Is.string(candidate.uri) && Is.string(candidate.languageId) && Is.integer(candidate.version) && Is.string(candidate.text);
        }
        TextDocumentItem2.is = is;
      })(TextDocumentItem = exports3.TextDocumentItem || (exports3.TextDocumentItem = {}));
      var MarkupKind;
      (function(MarkupKind2) {
        MarkupKind2.PlainText = "plaintext";
        MarkupKind2.Markdown = "markdown";
      })(MarkupKind = exports3.MarkupKind || (exports3.MarkupKind = {}));
      (function(MarkupKind2) {
        function is(value) {
          var candidate = value;
          return candidate === MarkupKind2.PlainText || candidate === MarkupKind2.Markdown;
        }
        MarkupKind2.is = is;
      })(MarkupKind = exports3.MarkupKind || (exports3.MarkupKind = {}));
      var MarkupContent;
      (function(MarkupContent2) {
        function is(value) {
          var candidate = value;
          return Is.objectLiteral(value) && MarkupKind.is(candidate.kind) && Is.string(candidate.value);
        }
        MarkupContent2.is = is;
      })(MarkupContent = exports3.MarkupContent || (exports3.MarkupContent = {}));
      var CompletionItemKind;
      (function(CompletionItemKind2) {
        CompletionItemKind2.Text = 1;
        CompletionItemKind2.Method = 2;
        CompletionItemKind2.Function = 3;
        CompletionItemKind2.Constructor = 4;
        CompletionItemKind2.Field = 5;
        CompletionItemKind2.Variable = 6;
        CompletionItemKind2.Class = 7;
        CompletionItemKind2.Interface = 8;
        CompletionItemKind2.Module = 9;
        CompletionItemKind2.Property = 10;
        CompletionItemKind2.Unit = 11;
        CompletionItemKind2.Value = 12;
        CompletionItemKind2.Enum = 13;
        CompletionItemKind2.Keyword = 14;
        CompletionItemKind2.Snippet = 15;
        CompletionItemKind2.Color = 16;
        CompletionItemKind2.File = 17;
        CompletionItemKind2.Reference = 18;
        CompletionItemKind2.Folder = 19;
        CompletionItemKind2.EnumMember = 20;
        CompletionItemKind2.Constant = 21;
        CompletionItemKind2.Struct = 22;
        CompletionItemKind2.Event = 23;
        CompletionItemKind2.Operator = 24;
        CompletionItemKind2.TypeParameter = 25;
      })(CompletionItemKind = exports3.CompletionItemKind || (exports3.CompletionItemKind = {}));
      var InsertTextFormat;
      (function(InsertTextFormat2) {
        InsertTextFormat2.PlainText = 1;
        InsertTextFormat2.Snippet = 2;
      })(InsertTextFormat = exports3.InsertTextFormat || (exports3.InsertTextFormat = {}));
      var CompletionItemTag;
      (function(CompletionItemTag2) {
        CompletionItemTag2.Deprecated = 1;
      })(CompletionItemTag = exports3.CompletionItemTag || (exports3.CompletionItemTag = {}));
      var InsertReplaceEdit;
      (function(InsertReplaceEdit2) {
        function create(newText, insert, replace) {
          return { newText, insert, replace };
        }
        InsertReplaceEdit2.create = create;
        function is(value) {
          var candidate = value;
          return candidate && Is.string(candidate.newText) && Range.is(candidate.insert) && Range.is(candidate.replace);
        }
        InsertReplaceEdit2.is = is;
      })(InsertReplaceEdit = exports3.InsertReplaceEdit || (exports3.InsertReplaceEdit = {}));
      var InsertTextMode;
      (function(InsertTextMode2) {
        InsertTextMode2.asIs = 1;
        InsertTextMode2.adjustIndentation = 2;
      })(InsertTextMode = exports3.InsertTextMode || (exports3.InsertTextMode = {}));
      var CompletionItem;
      (function(CompletionItem2) {
        function create(label) {
          return { label };
        }
        CompletionItem2.create = create;
      })(CompletionItem = exports3.CompletionItem || (exports3.CompletionItem = {}));
      var CompletionList;
      (function(CompletionList2) {
        function create(items, isIncomplete) {
          return { items: items ? items : [], isIncomplete: !!isIncomplete };
        }
        CompletionList2.create = create;
      })(CompletionList = exports3.CompletionList || (exports3.CompletionList = {}));
      var MarkedString;
      (function(MarkedString2) {
        function fromPlainText(plainText) {
          return plainText.replace(/[\\`*_{}[\]()#+\-.!]/g, "\\$&");
        }
        MarkedString2.fromPlainText = fromPlainText;
        function is(value) {
          var candidate = value;
          return Is.string(candidate) || Is.objectLiteral(candidate) && Is.string(candidate.language) && Is.string(candidate.value);
        }
        MarkedString2.is = is;
      })(MarkedString = exports3.MarkedString || (exports3.MarkedString = {}));
      var Hover;
      (function(Hover2) {
        function is(value) {
          var candidate = value;
          return !!candidate && Is.objectLiteral(candidate) && (MarkupContent.is(candidate.contents) || MarkedString.is(candidate.contents) || Is.typedArray(candidate.contents, MarkedString.is)) && (value.range === void 0 || Range.is(value.range));
        }
        Hover2.is = is;
      })(Hover = exports3.Hover || (exports3.Hover = {}));
      var ParameterInformation;
      (function(ParameterInformation2) {
        function create(label, documentation) {
          return documentation ? { label, documentation } : { label };
        }
        ParameterInformation2.create = create;
      })(ParameterInformation = exports3.ParameterInformation || (exports3.ParameterInformation = {}));
      var SignatureInformation;
      (function(SignatureInformation2) {
        function create(label, documentation) {
          var parameters = [];
          for (var _i = 2; _i < arguments.length; _i++) {
            parameters[_i - 2] = arguments[_i];
          }
          var result = { label };
          if (Is.defined(documentation)) {
            result.documentation = documentation;
          }
          if (Is.defined(parameters)) {
            result.parameters = parameters;
          } else {
            result.parameters = [];
          }
          return result;
        }
        SignatureInformation2.create = create;
      })(SignatureInformation = exports3.SignatureInformation || (exports3.SignatureInformation = {}));
      var DocumentHighlightKind;
      (function(DocumentHighlightKind2) {
        DocumentHighlightKind2.Text = 1;
        DocumentHighlightKind2.Read = 2;
        DocumentHighlightKind2.Write = 3;
      })(DocumentHighlightKind = exports3.DocumentHighlightKind || (exports3.DocumentHighlightKind = {}));
      var DocumentHighlight;
      (function(DocumentHighlight2) {
        function create(range, kind) {
          var result = { range };
          if (Is.number(kind)) {
            result.kind = kind;
          }
          return result;
        }
        DocumentHighlight2.create = create;
      })(DocumentHighlight = exports3.DocumentHighlight || (exports3.DocumentHighlight = {}));
      var SymbolKind;
      (function(SymbolKind2) {
        SymbolKind2.File = 1;
        SymbolKind2.Module = 2;
        SymbolKind2.Namespace = 3;
        SymbolKind2.Package = 4;
        SymbolKind2.Class = 5;
        SymbolKind2.Method = 6;
        SymbolKind2.Property = 7;
        SymbolKind2.Field = 8;
        SymbolKind2.Constructor = 9;
        SymbolKind2.Enum = 10;
        SymbolKind2.Interface = 11;
        SymbolKind2.Function = 12;
        SymbolKind2.Variable = 13;
        SymbolKind2.Constant = 14;
        SymbolKind2.String = 15;
        SymbolKind2.Number = 16;
        SymbolKind2.Boolean = 17;
        SymbolKind2.Array = 18;
        SymbolKind2.Object = 19;
        SymbolKind2.Key = 20;
        SymbolKind2.Null = 21;
        SymbolKind2.EnumMember = 22;
        SymbolKind2.Struct = 23;
        SymbolKind2.Event = 24;
        SymbolKind2.Operator = 25;
        SymbolKind2.TypeParameter = 26;
      })(SymbolKind = exports3.SymbolKind || (exports3.SymbolKind = {}));
      var SymbolTag;
      (function(SymbolTag2) {
        SymbolTag2.Deprecated = 1;
      })(SymbolTag = exports3.SymbolTag || (exports3.SymbolTag = {}));
      var SymbolInformation;
      (function(SymbolInformation2) {
        function create(name, kind, range, uri, containerName) {
          var result = {
            name,
            kind,
            location: { uri, range }
          };
          if (containerName) {
            result.containerName = containerName;
          }
          return result;
        }
        SymbolInformation2.create = create;
      })(SymbolInformation = exports3.SymbolInformation || (exports3.SymbolInformation = {}));
      var DocumentSymbol;
      (function(DocumentSymbol2) {
        function create(name, detail, kind, range, selectionRange, children) {
          var result = {
            name,
            detail,
            kind,
            range,
            selectionRange
          };
          if (children !== void 0) {
            result.children = children;
          }
          return result;
        }
        DocumentSymbol2.create = create;
        function is(value) {
          var candidate = value;
          return candidate && Is.string(candidate.name) && Is.number(candidate.kind) && Range.is(candidate.range) && Range.is(candidate.selectionRange) && (candidate.detail === void 0 || Is.string(candidate.detail)) && (candidate.deprecated === void 0 || Is.boolean(candidate.deprecated)) && (candidate.children === void 0 || Array.isArray(candidate.children)) && (candidate.tags === void 0 || Array.isArray(candidate.tags));
        }
        DocumentSymbol2.is = is;
      })(DocumentSymbol = exports3.DocumentSymbol || (exports3.DocumentSymbol = {}));
      var CodeActionKind;
      (function(CodeActionKind2) {
        CodeActionKind2.Empty = "";
        CodeActionKind2.QuickFix = "quickfix";
        CodeActionKind2.Refactor = "refactor";
        CodeActionKind2.RefactorExtract = "refactor.extract";
        CodeActionKind2.RefactorInline = "refactor.inline";
        CodeActionKind2.RefactorRewrite = "refactor.rewrite";
        CodeActionKind2.Source = "source";
        CodeActionKind2.SourceOrganizeImports = "source.organizeImports";
        CodeActionKind2.SourceFixAll = "source.fixAll";
      })(CodeActionKind = exports3.CodeActionKind || (exports3.CodeActionKind = {}));
      var CodeActionContext;
      (function(CodeActionContext2) {
        function create(diagnostics, only) {
          var result = { diagnostics };
          if (only !== void 0 && only !== null) {
            result.only = only;
          }
          return result;
        }
        CodeActionContext2.create = create;
        function is(value) {
          var candidate = value;
          return Is.defined(candidate) && Is.typedArray(candidate.diagnostics, Diagnostic.is) && (candidate.only === void 0 || Is.typedArray(candidate.only, Is.string));
        }
        CodeActionContext2.is = is;
      })(CodeActionContext = exports3.CodeActionContext || (exports3.CodeActionContext = {}));
      var CodeAction;
      (function(CodeAction2) {
        function create(title, kindOrCommandOrEdit, kind) {
          var result = { title };
          var checkKind = true;
          if (typeof kindOrCommandOrEdit === "string") {
            checkKind = false;
            result.kind = kindOrCommandOrEdit;
          } else if (Command.is(kindOrCommandOrEdit)) {
            result.command = kindOrCommandOrEdit;
          } else {
            result.edit = kindOrCommandOrEdit;
          }
          if (checkKind && kind !== void 0) {
            result.kind = kind;
          }
          return result;
        }
        CodeAction2.create = create;
        function is(value) {
          var candidate = value;
          return candidate && Is.string(candidate.title) && (candidate.diagnostics === void 0 || Is.typedArray(candidate.diagnostics, Diagnostic.is)) && (candidate.kind === void 0 || Is.string(candidate.kind)) && (candidate.edit !== void 0 || candidate.command !== void 0) && (candidate.command === void 0 || Command.is(candidate.command)) && (candidate.isPreferred === void 0 || Is.boolean(candidate.isPreferred)) && (candidate.edit === void 0 || WorkspaceEdit.is(candidate.edit));
        }
        CodeAction2.is = is;
      })(CodeAction = exports3.CodeAction || (exports3.CodeAction = {}));
      var CodeLens;
      (function(CodeLens2) {
        function create(range, data) {
          var result = { range };
          if (Is.defined(data)) {
            result.data = data;
          }
          return result;
        }
        CodeLens2.create = create;
        function is(value) {
          var candidate = value;
          return Is.defined(candidate) && Range.is(candidate.range) && (Is.undefined(candidate.command) || Command.is(candidate.command));
        }
        CodeLens2.is = is;
      })(CodeLens = exports3.CodeLens || (exports3.CodeLens = {}));
      var FormattingOptions;
      (function(FormattingOptions2) {
        function create(tabSize, insertSpaces) {
          return { tabSize, insertSpaces };
        }
        FormattingOptions2.create = create;
        function is(value) {
          var candidate = value;
          return Is.defined(candidate) && Is.uinteger(candidate.tabSize) && Is.boolean(candidate.insertSpaces);
        }
        FormattingOptions2.is = is;
      })(FormattingOptions = exports3.FormattingOptions || (exports3.FormattingOptions = {}));
      var DocumentLink;
      (function(DocumentLink2) {
        function create(range, target, data) {
          return { range, target, data };
        }
        DocumentLink2.create = create;
        function is(value) {
          var candidate = value;
          return Is.defined(candidate) && Range.is(candidate.range) && (Is.undefined(candidate.target) || Is.string(candidate.target));
        }
        DocumentLink2.is = is;
      })(DocumentLink = exports3.DocumentLink || (exports3.DocumentLink = {}));
      var SelectionRange;
      (function(SelectionRange2) {
        function create(range, parent) {
          return { range, parent };
        }
        SelectionRange2.create = create;
        function is(value) {
          var candidate = value;
          return candidate !== void 0 && Range.is(candidate.range) && (candidate.parent === void 0 || SelectionRange2.is(candidate.parent));
        }
        SelectionRange2.is = is;
      })(SelectionRange = exports3.SelectionRange || (exports3.SelectionRange = {}));
      exports3.EOL = ["\n", "\r\n", "\r"];
      var TextDocument;
      (function(TextDocument2) {
        function create(uri, languageId, version, content) {
          return new FullTextDocument(uri, languageId, version, content);
        }
        TextDocument2.create = create;
        function is(value) {
          var candidate = value;
          return Is.defined(candidate) && Is.string(candidate.uri) && (Is.undefined(candidate.languageId) || Is.string(candidate.languageId)) && Is.uinteger(candidate.lineCount) && Is.func(candidate.getText) && Is.func(candidate.positionAt) && Is.func(candidate.offsetAt) ? true : false;
        }
        TextDocument2.is = is;
        function applyEdits(document, edits) {
          var text = document.getText();
          var sortedEdits = mergeSort(edits, function(a, b) {
            var diff = a.range.start.line - b.range.start.line;
            if (diff === 0) {
              return a.range.start.character - b.range.start.character;
            }
            return diff;
          });
          var lastModifiedOffset = text.length;
          for (var i = sortedEdits.length - 1; i >= 0; i--) {
            var e = sortedEdits[i];
            var startOffset = document.offsetAt(e.range.start);
            var endOffset = document.offsetAt(e.range.end);
            if (endOffset <= lastModifiedOffset) {
              text = text.substring(0, startOffset) + e.newText + text.substring(endOffset, text.length);
            } else {
              throw new Error("Overlapping edit");
            }
            lastModifiedOffset = startOffset;
          }
          return text;
        }
        TextDocument2.applyEdits = applyEdits;
        function mergeSort(data, compare) {
          if (data.length <= 1) {
            return data;
          }
          var p = data.length / 2 | 0;
          var left = data.slice(0, p);
          var right = data.slice(p);
          mergeSort(left, compare);
          mergeSort(right, compare);
          var leftIdx = 0;
          var rightIdx = 0;
          var i = 0;
          while (leftIdx < left.length && rightIdx < right.length) {
            var ret = compare(left[leftIdx], right[rightIdx]);
            if (ret <= 0) {
              data[i++] = left[leftIdx++];
            } else {
              data[i++] = right[rightIdx++];
            }
          }
          while (leftIdx < left.length) {
            data[i++] = left[leftIdx++];
          }
          while (rightIdx < right.length) {
            data[i++] = right[rightIdx++];
          }
          return data;
        }
      })(TextDocument = exports3.TextDocument || (exports3.TextDocument = {}));
      var FullTextDocument = (
        /** @class */
        function() {
          function FullTextDocument2(uri, languageId, version, content) {
            this._uri = uri;
            this._languageId = languageId;
            this._version = version;
            this._content = content;
            this._lineOffsets = void 0;
          }
          Object.defineProperty(FullTextDocument2.prototype, "uri", {
            get: function() {
              return this._uri;
            },
            enumerable: false,
            configurable: true
          });
          Object.defineProperty(FullTextDocument2.prototype, "languageId", {
            get: function() {
              return this._languageId;
            },
            enumerable: false,
            configurable: true
          });
          Object.defineProperty(FullTextDocument2.prototype, "version", {
            get: function() {
              return this._version;
            },
            enumerable: false,
            configurable: true
          });
          FullTextDocument2.prototype.getText = function(range) {
            if (range) {
              var start = this.offsetAt(range.start);
              var end = this.offsetAt(range.end);
              return this._content.substring(start, end);
            }
            return this._content;
          };
          FullTextDocument2.prototype.update = function(event, version) {
            this._content = event.text;
            this._version = version;
            this._lineOffsets = void 0;
          };
          FullTextDocument2.prototype.getLineOffsets = function() {
            if (this._lineOffsets === void 0) {
              var lineOffsets = [];
              var text = this._content;
              var isLineStart = true;
              for (var i = 0; i < text.length; i++) {
                if (isLineStart) {
                  lineOffsets.push(i);
                  isLineStart = false;
                }
                var ch = text.charAt(i);
                isLineStart = ch === "\r" || ch === "\n";
                if (ch === "\r" && i + 1 < text.length && text.charAt(i + 1) === "\n") {
                  i++;
                }
              }
              if (isLineStart && text.length > 0) {
                lineOffsets.push(text.length);
              }
              this._lineOffsets = lineOffsets;
            }
            return this._lineOffsets;
          };
          FullTextDocument2.prototype.positionAt = function(offset) {
            offset = Math.max(Math.min(offset, this._content.length), 0);
            var lineOffsets = this.getLineOffsets();
            var low = 0, high = lineOffsets.length;
            if (high === 0) {
              return Position.create(0, offset);
            }
            while (low < high) {
              var mid = Math.floor((low + high) / 2);
              if (lineOffsets[mid] > offset) {
                high = mid;
              } else {
                low = mid + 1;
              }
            }
            var line = low - 1;
            return Position.create(line, offset - lineOffsets[line]);
          };
          FullTextDocument2.prototype.offsetAt = function(position) {
            var lineOffsets = this.getLineOffsets();
            if (position.line >= lineOffsets.length) {
              return this._content.length;
            } else if (position.line < 0) {
              return 0;
            }
            var lineOffset = lineOffsets[position.line];
            var nextLineOffset = position.line + 1 < lineOffsets.length ? lineOffsets[position.line + 1] : this._content.length;
            return Math.max(Math.min(lineOffset + position.character, nextLineOffset), lineOffset);
          };
          Object.defineProperty(FullTextDocument2.prototype, "lineCount", {
            get: function() {
              return this.getLineOffsets().length;
            },
            enumerable: false,
            configurable: true
          });
          return FullTextDocument2;
        }()
      );
      var Is;
      (function(Is2) {
        var toString = Object.prototype.toString;
        function defined(value) {
          return typeof value !== "undefined";
        }
        Is2.defined = defined;
        function undefined2(value) {
          return typeof value === "undefined";
        }
        Is2.undefined = undefined2;
        function boolean(value) {
          return value === true || value === false;
        }
        Is2.boolean = boolean;
        function string(value) {
          return toString.call(value) === "[object String]";
        }
        Is2.string = string;
        function number(value) {
          return toString.call(value) === "[object Number]";
        }
        Is2.number = number;
        function numberRange(value, min, max) {
          return toString.call(value) === "[object Number]" && min <= value && value <= max;
        }
        Is2.numberRange = numberRange;
        function integer2(value) {
          return toString.call(value) === "[object Number]" && -2147483648 <= value && value <= 2147483647;
        }
        Is2.integer = integer2;
        function uinteger2(value) {
          return toString.call(value) === "[object Number]" && 0 <= value && value <= 2147483647;
        }
        Is2.uinteger = uinteger2;
        function func(value) {
          return toString.call(value) === "[object Function]";
        }
        Is2.func = func;
        function objectLiteral(value) {
          return value !== null && typeof value === "object";
        }
        Is2.objectLiteral = objectLiteral;
        function typedArray(value, check) {
          return Array.isArray(value) && value.every(check);
        }
        Is2.typedArray = typedArray;
      })(Is || (Is = {}));
    });
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/messages.js
var require_messages2 = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/messages.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.ProtocolNotificationType = exports2.ProtocolNotificationType0 = exports2.ProtocolRequestType = exports2.ProtocolRequestType0 = exports2.RegistrationType = void 0;
    var vscode_jsonrpc_1 = require_main();
    var RegistrationType = class {
      constructor(method) {
        this.method = method;
      }
    };
    exports2.RegistrationType = RegistrationType;
    var ProtocolRequestType0 = class extends vscode_jsonrpc_1.RequestType0 {
      constructor(method) {
        super(method);
      }
    };
    exports2.ProtocolRequestType0 = ProtocolRequestType0;
    var ProtocolRequestType = class extends vscode_jsonrpc_1.RequestType {
      constructor(method) {
        super(method, vscode_jsonrpc_1.ParameterStructures.byName);
      }
    };
    exports2.ProtocolRequestType = ProtocolRequestType;
    var ProtocolNotificationType0 = class extends vscode_jsonrpc_1.NotificationType0 {
      constructor(method) {
        super(method);
      }
    };
    exports2.ProtocolNotificationType0 = ProtocolNotificationType0;
    var ProtocolNotificationType = class extends vscode_jsonrpc_1.NotificationType {
      constructor(method) {
        super(method, vscode_jsonrpc_1.ParameterStructures.byName);
      }
    };
    exports2.ProtocolNotificationType = ProtocolNotificationType;
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/utils/is.js
var require_is3 = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/utils/is.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.objectLiteral = exports2.typedArray = exports2.stringArray = exports2.array = exports2.func = exports2.error = exports2.number = exports2.string = exports2.boolean = void 0;
    function boolean(value) {
      return value === true || value === false;
    }
    exports2.boolean = boolean;
    function string(value) {
      return typeof value === "string" || value instanceof String;
    }
    exports2.string = string;
    function number(value) {
      return typeof value === "number" || value instanceof Number;
    }
    exports2.number = number;
    function error(value) {
      return value instanceof Error;
    }
    exports2.error = error;
    function func(value) {
      return typeof value === "function";
    }
    exports2.func = func;
    function array(value) {
      return Array.isArray(value);
    }
    exports2.array = array;
    function stringArray(value) {
      return array(value) && value.every((elem) => string(elem));
    }
    exports2.stringArray = stringArray;
    function typedArray(value, check) {
      return Array.isArray(value) && value.every(check);
    }
    exports2.typedArray = typedArray;
    function objectLiteral(value) {
      return value !== null && typeof value === "object";
    }
    exports2.objectLiteral = objectLiteral;
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.implementation.js
var require_protocol_implementation = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.implementation.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.ImplementationRequest = void 0;
    var messages_1 = require_messages2();
    var ImplementationRequest;
    (function(ImplementationRequest2) {
      ImplementationRequest2.method = "textDocument/implementation";
      ImplementationRequest2.type = new messages_1.ProtocolRequestType(ImplementationRequest2.method);
    })(ImplementationRequest = exports2.ImplementationRequest || (exports2.ImplementationRequest = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.typeDefinition.js
var require_protocol_typeDefinition = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.typeDefinition.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.TypeDefinitionRequest = void 0;
    var messages_1 = require_messages2();
    var TypeDefinitionRequest;
    (function(TypeDefinitionRequest2) {
      TypeDefinitionRequest2.method = "textDocument/typeDefinition";
      TypeDefinitionRequest2.type = new messages_1.ProtocolRequestType(TypeDefinitionRequest2.method);
    })(TypeDefinitionRequest = exports2.TypeDefinitionRequest || (exports2.TypeDefinitionRequest = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.workspaceFolders.js
var require_protocol_workspaceFolders = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.workspaceFolders.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.DidChangeWorkspaceFoldersNotification = exports2.WorkspaceFoldersRequest = void 0;
    var messages_1 = require_messages2();
    var WorkspaceFoldersRequest;
    (function(WorkspaceFoldersRequest2) {
      WorkspaceFoldersRequest2.type = new messages_1.ProtocolRequestType0("workspace/workspaceFolders");
    })(WorkspaceFoldersRequest = exports2.WorkspaceFoldersRequest || (exports2.WorkspaceFoldersRequest = {}));
    var DidChangeWorkspaceFoldersNotification;
    (function(DidChangeWorkspaceFoldersNotification2) {
      DidChangeWorkspaceFoldersNotification2.type = new messages_1.ProtocolNotificationType("workspace/didChangeWorkspaceFolders");
    })(DidChangeWorkspaceFoldersNotification = exports2.DidChangeWorkspaceFoldersNotification || (exports2.DidChangeWorkspaceFoldersNotification = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.configuration.js
var require_protocol_configuration = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.configuration.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.ConfigurationRequest = void 0;
    var messages_1 = require_messages2();
    var ConfigurationRequest;
    (function(ConfigurationRequest2) {
      ConfigurationRequest2.type = new messages_1.ProtocolRequestType("workspace/configuration");
    })(ConfigurationRequest = exports2.ConfigurationRequest || (exports2.ConfigurationRequest = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.colorProvider.js
var require_protocol_colorProvider = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.colorProvider.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.ColorPresentationRequest = exports2.DocumentColorRequest = void 0;
    var messages_1 = require_messages2();
    var DocumentColorRequest;
    (function(DocumentColorRequest2) {
      DocumentColorRequest2.method = "textDocument/documentColor";
      DocumentColorRequest2.type = new messages_1.ProtocolRequestType(DocumentColorRequest2.method);
    })(DocumentColorRequest = exports2.DocumentColorRequest || (exports2.DocumentColorRequest = {}));
    var ColorPresentationRequest;
    (function(ColorPresentationRequest2) {
      ColorPresentationRequest2.type = new messages_1.ProtocolRequestType("textDocument/colorPresentation");
    })(ColorPresentationRequest = exports2.ColorPresentationRequest || (exports2.ColorPresentationRequest = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.foldingRange.js
var require_protocol_foldingRange = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.foldingRange.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.FoldingRangeRequest = exports2.FoldingRangeKind = void 0;
    var messages_1 = require_messages2();
    var FoldingRangeKind;
    (function(FoldingRangeKind2) {
      FoldingRangeKind2["Comment"] = "comment";
      FoldingRangeKind2["Imports"] = "imports";
      FoldingRangeKind2["Region"] = "region";
    })(FoldingRangeKind = exports2.FoldingRangeKind || (exports2.FoldingRangeKind = {}));
    var FoldingRangeRequest;
    (function(FoldingRangeRequest2) {
      FoldingRangeRequest2.method = "textDocument/foldingRange";
      FoldingRangeRequest2.type = new messages_1.ProtocolRequestType(FoldingRangeRequest2.method);
    })(FoldingRangeRequest = exports2.FoldingRangeRequest || (exports2.FoldingRangeRequest = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.declaration.js
var require_protocol_declaration = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.declaration.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.DeclarationRequest = void 0;
    var messages_1 = require_messages2();
    var DeclarationRequest;
    (function(DeclarationRequest2) {
      DeclarationRequest2.method = "textDocument/declaration";
      DeclarationRequest2.type = new messages_1.ProtocolRequestType(DeclarationRequest2.method);
    })(DeclarationRequest = exports2.DeclarationRequest || (exports2.DeclarationRequest = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.selectionRange.js
var require_protocol_selectionRange = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.selectionRange.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.SelectionRangeRequest = void 0;
    var messages_1 = require_messages2();
    var SelectionRangeRequest;
    (function(SelectionRangeRequest2) {
      SelectionRangeRequest2.method = "textDocument/selectionRange";
      SelectionRangeRequest2.type = new messages_1.ProtocolRequestType(SelectionRangeRequest2.method);
    })(SelectionRangeRequest = exports2.SelectionRangeRequest || (exports2.SelectionRangeRequest = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.progress.js
var require_protocol_progress = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.progress.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.WorkDoneProgressCancelNotification = exports2.WorkDoneProgressCreateRequest = exports2.WorkDoneProgress = void 0;
    var vscode_jsonrpc_1 = require_main();
    var messages_1 = require_messages2();
    var WorkDoneProgress;
    (function(WorkDoneProgress2) {
      WorkDoneProgress2.type = new vscode_jsonrpc_1.ProgressType();
      function is(value) {
        return value === WorkDoneProgress2.type;
      }
      WorkDoneProgress2.is = is;
    })(WorkDoneProgress = exports2.WorkDoneProgress || (exports2.WorkDoneProgress = {}));
    var WorkDoneProgressCreateRequest;
    (function(WorkDoneProgressCreateRequest2) {
      WorkDoneProgressCreateRequest2.type = new messages_1.ProtocolRequestType("window/workDoneProgress/create");
    })(WorkDoneProgressCreateRequest = exports2.WorkDoneProgressCreateRequest || (exports2.WorkDoneProgressCreateRequest = {}));
    var WorkDoneProgressCancelNotification;
    (function(WorkDoneProgressCancelNotification2) {
      WorkDoneProgressCancelNotification2.type = new messages_1.ProtocolNotificationType("window/workDoneProgress/cancel");
    })(WorkDoneProgressCancelNotification = exports2.WorkDoneProgressCancelNotification || (exports2.WorkDoneProgressCancelNotification = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.callHierarchy.js
var require_protocol_callHierarchy = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.callHierarchy.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.CallHierarchyOutgoingCallsRequest = exports2.CallHierarchyIncomingCallsRequest = exports2.CallHierarchyPrepareRequest = void 0;
    var messages_1 = require_messages2();
    var CallHierarchyPrepareRequest;
    (function(CallHierarchyPrepareRequest2) {
      CallHierarchyPrepareRequest2.method = "textDocument/prepareCallHierarchy";
      CallHierarchyPrepareRequest2.type = new messages_1.ProtocolRequestType(CallHierarchyPrepareRequest2.method);
    })(CallHierarchyPrepareRequest = exports2.CallHierarchyPrepareRequest || (exports2.CallHierarchyPrepareRequest = {}));
    var CallHierarchyIncomingCallsRequest;
    (function(CallHierarchyIncomingCallsRequest2) {
      CallHierarchyIncomingCallsRequest2.method = "callHierarchy/incomingCalls";
      CallHierarchyIncomingCallsRequest2.type = new messages_1.ProtocolRequestType(CallHierarchyIncomingCallsRequest2.method);
    })(CallHierarchyIncomingCallsRequest = exports2.CallHierarchyIncomingCallsRequest || (exports2.CallHierarchyIncomingCallsRequest = {}));
    var CallHierarchyOutgoingCallsRequest;
    (function(CallHierarchyOutgoingCallsRequest2) {
      CallHierarchyOutgoingCallsRequest2.method = "callHierarchy/outgoingCalls";
      CallHierarchyOutgoingCallsRequest2.type = new messages_1.ProtocolRequestType(CallHierarchyOutgoingCallsRequest2.method);
    })(CallHierarchyOutgoingCallsRequest = exports2.CallHierarchyOutgoingCallsRequest || (exports2.CallHierarchyOutgoingCallsRequest = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.semanticTokens.js
var require_protocol_semanticTokens = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.semanticTokens.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.SemanticTokensRefreshRequest = exports2.SemanticTokensRangeRequest = exports2.SemanticTokensDeltaRequest = exports2.SemanticTokensRequest = exports2.SemanticTokensRegistrationType = exports2.TokenFormat = exports2.SemanticTokens = exports2.SemanticTokenModifiers = exports2.SemanticTokenTypes = void 0;
    var messages_1 = require_messages2();
    var SemanticTokenTypes;
    (function(SemanticTokenTypes2) {
      SemanticTokenTypes2["namespace"] = "namespace";
      SemanticTokenTypes2["type"] = "type";
      SemanticTokenTypes2["class"] = "class";
      SemanticTokenTypes2["enum"] = "enum";
      SemanticTokenTypes2["interface"] = "interface";
      SemanticTokenTypes2["struct"] = "struct";
      SemanticTokenTypes2["typeParameter"] = "typeParameter";
      SemanticTokenTypes2["parameter"] = "parameter";
      SemanticTokenTypes2["variable"] = "variable";
      SemanticTokenTypes2["property"] = "property";
      SemanticTokenTypes2["enumMember"] = "enumMember";
      SemanticTokenTypes2["event"] = "event";
      SemanticTokenTypes2["function"] = "function";
      SemanticTokenTypes2["method"] = "method";
      SemanticTokenTypes2["macro"] = "macro";
      SemanticTokenTypes2["keyword"] = "keyword";
      SemanticTokenTypes2["modifier"] = "modifier";
      SemanticTokenTypes2["comment"] = "comment";
      SemanticTokenTypes2["string"] = "string";
      SemanticTokenTypes2["number"] = "number";
      SemanticTokenTypes2["regexp"] = "regexp";
      SemanticTokenTypes2["operator"] = "operator";
    })(SemanticTokenTypes = exports2.SemanticTokenTypes || (exports2.SemanticTokenTypes = {}));
    var SemanticTokenModifiers;
    (function(SemanticTokenModifiers2) {
      SemanticTokenModifiers2["declaration"] = "declaration";
      SemanticTokenModifiers2["definition"] = "definition";
      SemanticTokenModifiers2["readonly"] = "readonly";
      SemanticTokenModifiers2["static"] = "static";
      SemanticTokenModifiers2["deprecated"] = "deprecated";
      SemanticTokenModifiers2["abstract"] = "abstract";
      SemanticTokenModifiers2["async"] = "async";
      SemanticTokenModifiers2["modification"] = "modification";
      SemanticTokenModifiers2["documentation"] = "documentation";
      SemanticTokenModifiers2["defaultLibrary"] = "defaultLibrary";
    })(SemanticTokenModifiers = exports2.SemanticTokenModifiers || (exports2.SemanticTokenModifiers = {}));
    var SemanticTokens;
    (function(SemanticTokens2) {
      function is(value) {
        const candidate = value;
        return candidate !== void 0 && (candidate.resultId === void 0 || typeof candidate.resultId === "string") && Array.isArray(candidate.data) && (candidate.data.length === 0 || typeof candidate.data[0] === "number");
      }
      SemanticTokens2.is = is;
    })(SemanticTokens = exports2.SemanticTokens || (exports2.SemanticTokens = {}));
    var TokenFormat;
    (function(TokenFormat2) {
      TokenFormat2.Relative = "relative";
    })(TokenFormat = exports2.TokenFormat || (exports2.TokenFormat = {}));
    var SemanticTokensRegistrationType;
    (function(SemanticTokensRegistrationType2) {
      SemanticTokensRegistrationType2.method = "textDocument/semanticTokens";
      SemanticTokensRegistrationType2.type = new messages_1.RegistrationType(SemanticTokensRegistrationType2.method);
    })(SemanticTokensRegistrationType = exports2.SemanticTokensRegistrationType || (exports2.SemanticTokensRegistrationType = {}));
    var SemanticTokensRequest;
    (function(SemanticTokensRequest2) {
      SemanticTokensRequest2.method = "textDocument/semanticTokens/full";
      SemanticTokensRequest2.type = new messages_1.ProtocolRequestType(SemanticTokensRequest2.method);
    })(SemanticTokensRequest = exports2.SemanticTokensRequest || (exports2.SemanticTokensRequest = {}));
    var SemanticTokensDeltaRequest;
    (function(SemanticTokensDeltaRequest2) {
      SemanticTokensDeltaRequest2.method = "textDocument/semanticTokens/full/delta";
      SemanticTokensDeltaRequest2.type = new messages_1.ProtocolRequestType(SemanticTokensDeltaRequest2.method);
    })(SemanticTokensDeltaRequest = exports2.SemanticTokensDeltaRequest || (exports2.SemanticTokensDeltaRequest = {}));
    var SemanticTokensRangeRequest;
    (function(SemanticTokensRangeRequest2) {
      SemanticTokensRangeRequest2.method = "textDocument/semanticTokens/range";
      SemanticTokensRangeRequest2.type = new messages_1.ProtocolRequestType(SemanticTokensRangeRequest2.method);
    })(SemanticTokensRangeRequest = exports2.SemanticTokensRangeRequest || (exports2.SemanticTokensRangeRequest = {}));
    var SemanticTokensRefreshRequest;
    (function(SemanticTokensRefreshRequest2) {
      SemanticTokensRefreshRequest2.method = `workspace/semanticTokens/refresh`;
      SemanticTokensRefreshRequest2.type = new messages_1.ProtocolRequestType0(SemanticTokensRefreshRequest2.method);
    })(SemanticTokensRefreshRequest = exports2.SemanticTokensRefreshRequest || (exports2.SemanticTokensRefreshRequest = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.showDocument.js
var require_protocol_showDocument = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.showDocument.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.ShowDocumentRequest = void 0;
    var messages_1 = require_messages2();
    var ShowDocumentRequest;
    (function(ShowDocumentRequest2) {
      ShowDocumentRequest2.method = "window/showDocument";
      ShowDocumentRequest2.type = new messages_1.ProtocolRequestType(ShowDocumentRequest2.method);
    })(ShowDocumentRequest = exports2.ShowDocumentRequest || (exports2.ShowDocumentRequest = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.linkedEditingRange.js
var require_protocol_linkedEditingRange = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.linkedEditingRange.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.LinkedEditingRangeRequest = void 0;
    var messages_1 = require_messages2();
    var LinkedEditingRangeRequest;
    (function(LinkedEditingRangeRequest2) {
      LinkedEditingRangeRequest2.method = "textDocument/linkedEditingRange";
      LinkedEditingRangeRequest2.type = new messages_1.ProtocolRequestType(LinkedEditingRangeRequest2.method);
    })(LinkedEditingRangeRequest = exports2.LinkedEditingRangeRequest || (exports2.LinkedEditingRangeRequest = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.fileOperations.js
var require_protocol_fileOperations = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.fileOperations.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.WillDeleteFilesRequest = exports2.DidDeleteFilesNotification = exports2.DidRenameFilesNotification = exports2.WillRenameFilesRequest = exports2.DidCreateFilesNotification = exports2.WillCreateFilesRequest = exports2.FileOperationPatternKind = void 0;
    var messages_1 = require_messages2();
    var FileOperationPatternKind;
    (function(FileOperationPatternKind2) {
      FileOperationPatternKind2.file = "file";
      FileOperationPatternKind2.folder = "folder";
    })(FileOperationPatternKind = exports2.FileOperationPatternKind || (exports2.FileOperationPatternKind = {}));
    var WillCreateFilesRequest;
    (function(WillCreateFilesRequest2) {
      WillCreateFilesRequest2.method = "workspace/willCreateFiles";
      WillCreateFilesRequest2.type = new messages_1.ProtocolRequestType(WillCreateFilesRequest2.method);
    })(WillCreateFilesRequest = exports2.WillCreateFilesRequest || (exports2.WillCreateFilesRequest = {}));
    var DidCreateFilesNotification;
    (function(DidCreateFilesNotification2) {
      DidCreateFilesNotification2.method = "workspace/didCreateFiles";
      DidCreateFilesNotification2.type = new messages_1.ProtocolNotificationType(DidCreateFilesNotification2.method);
    })(DidCreateFilesNotification = exports2.DidCreateFilesNotification || (exports2.DidCreateFilesNotification = {}));
    var WillRenameFilesRequest;
    (function(WillRenameFilesRequest2) {
      WillRenameFilesRequest2.method = "workspace/willRenameFiles";
      WillRenameFilesRequest2.type = new messages_1.ProtocolRequestType(WillRenameFilesRequest2.method);
    })(WillRenameFilesRequest = exports2.WillRenameFilesRequest || (exports2.WillRenameFilesRequest = {}));
    var DidRenameFilesNotification;
    (function(DidRenameFilesNotification2) {
      DidRenameFilesNotification2.method = "workspace/didRenameFiles";
      DidRenameFilesNotification2.type = new messages_1.ProtocolNotificationType(DidRenameFilesNotification2.method);
    })(DidRenameFilesNotification = exports2.DidRenameFilesNotification || (exports2.DidRenameFilesNotification = {}));
    var DidDeleteFilesNotification;
    (function(DidDeleteFilesNotification2) {
      DidDeleteFilesNotification2.method = "workspace/didDeleteFiles";
      DidDeleteFilesNotification2.type = new messages_1.ProtocolNotificationType(DidDeleteFilesNotification2.method);
    })(DidDeleteFilesNotification = exports2.DidDeleteFilesNotification || (exports2.DidDeleteFilesNotification = {}));
    var WillDeleteFilesRequest;
    (function(WillDeleteFilesRequest2) {
      WillDeleteFilesRequest2.method = "workspace/willDeleteFiles";
      WillDeleteFilesRequest2.type = new messages_1.ProtocolRequestType(WillDeleteFilesRequest2.method);
    })(WillDeleteFilesRequest = exports2.WillDeleteFilesRequest || (exports2.WillDeleteFilesRequest = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.moniker.js
var require_protocol_moniker = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.moniker.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.MonikerRequest = exports2.MonikerKind = exports2.UniquenessLevel = void 0;
    var messages_1 = require_messages2();
    var UniquenessLevel;
    (function(UniquenessLevel2) {
      UniquenessLevel2["document"] = "document";
      UniquenessLevel2["project"] = "project";
      UniquenessLevel2["group"] = "group";
      UniquenessLevel2["scheme"] = "scheme";
      UniquenessLevel2["global"] = "global";
    })(UniquenessLevel = exports2.UniquenessLevel || (exports2.UniquenessLevel = {}));
    var MonikerKind;
    (function(MonikerKind2) {
      MonikerKind2["import"] = "import";
      MonikerKind2["export"] = "export";
      MonikerKind2["local"] = "local";
    })(MonikerKind = exports2.MonikerKind || (exports2.MonikerKind = {}));
    var MonikerRequest;
    (function(MonikerRequest2) {
      MonikerRequest2.method = "textDocument/moniker";
      MonikerRequest2.type = new messages_1.ProtocolRequestType(MonikerRequest2.method);
    })(MonikerRequest = exports2.MonikerRequest || (exports2.MonikerRequest = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/protocol.js
var require_protocol = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/protocol.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.DocumentLinkRequest = exports2.CodeLensRefreshRequest = exports2.CodeLensResolveRequest = exports2.CodeLensRequest = exports2.WorkspaceSymbolRequest = exports2.CodeActionResolveRequest = exports2.CodeActionRequest = exports2.DocumentSymbolRequest = exports2.DocumentHighlightRequest = exports2.ReferencesRequest = exports2.DefinitionRequest = exports2.SignatureHelpRequest = exports2.SignatureHelpTriggerKind = exports2.HoverRequest = exports2.CompletionResolveRequest = exports2.CompletionRequest = exports2.CompletionTriggerKind = exports2.PublishDiagnosticsNotification = exports2.WatchKind = exports2.FileChangeType = exports2.DidChangeWatchedFilesNotification = exports2.WillSaveTextDocumentWaitUntilRequest = exports2.WillSaveTextDocumentNotification = exports2.TextDocumentSaveReason = exports2.DidSaveTextDocumentNotification = exports2.DidCloseTextDocumentNotification = exports2.DidChangeTextDocumentNotification = exports2.TextDocumentContentChangeEvent = exports2.DidOpenTextDocumentNotification = exports2.TextDocumentSyncKind = exports2.TelemetryEventNotification = exports2.LogMessageNotification = exports2.ShowMessageRequest = exports2.ShowMessageNotification = exports2.MessageType = exports2.DidChangeConfigurationNotification = exports2.ExitNotification = exports2.ShutdownRequest = exports2.InitializedNotification = exports2.InitializeError = exports2.InitializeRequest = exports2.WorkDoneProgressOptions = exports2.TextDocumentRegistrationOptions = exports2.StaticRegistrationOptions = exports2.FailureHandlingKind = exports2.ResourceOperationKind = exports2.UnregistrationRequest = exports2.RegistrationRequest = exports2.DocumentSelector = exports2.DocumentFilter = void 0;
    exports2.MonikerRequest = exports2.MonikerKind = exports2.UniquenessLevel = exports2.WillDeleteFilesRequest = exports2.DidDeleteFilesNotification = exports2.WillRenameFilesRequest = exports2.DidRenameFilesNotification = exports2.WillCreateFilesRequest = exports2.DidCreateFilesNotification = exports2.FileOperationPatternKind = exports2.LinkedEditingRangeRequest = exports2.ShowDocumentRequest = exports2.SemanticTokensRegistrationType = exports2.SemanticTokensRefreshRequest = exports2.SemanticTokensRangeRequest = exports2.SemanticTokensDeltaRequest = exports2.SemanticTokensRequest = exports2.TokenFormat = exports2.SemanticTokens = exports2.SemanticTokenModifiers = exports2.SemanticTokenTypes = exports2.CallHierarchyPrepareRequest = exports2.CallHierarchyOutgoingCallsRequest = exports2.CallHierarchyIncomingCallsRequest = exports2.WorkDoneProgressCancelNotification = exports2.WorkDoneProgressCreateRequest = exports2.WorkDoneProgress = exports2.SelectionRangeRequest = exports2.DeclarationRequest = exports2.FoldingRangeRequest = exports2.ColorPresentationRequest = exports2.DocumentColorRequest = exports2.ConfigurationRequest = exports2.DidChangeWorkspaceFoldersNotification = exports2.WorkspaceFoldersRequest = exports2.TypeDefinitionRequest = exports2.ImplementationRequest = exports2.ApplyWorkspaceEditRequest = exports2.ExecuteCommandRequest = exports2.PrepareRenameRequest = exports2.RenameRequest = exports2.PrepareSupportDefaultBehavior = exports2.DocumentOnTypeFormattingRequest = exports2.DocumentRangeFormattingRequest = exports2.DocumentFormattingRequest = exports2.DocumentLinkResolveRequest = void 0;
    var Is = require_is3();
    var messages_1 = require_messages2();
    var protocol_implementation_1 = require_protocol_implementation();
    Object.defineProperty(exports2, "ImplementationRequest", { enumerable: true, get: function() {
      return protocol_implementation_1.ImplementationRequest;
    } });
    var protocol_typeDefinition_1 = require_protocol_typeDefinition();
    Object.defineProperty(exports2, "TypeDefinitionRequest", { enumerable: true, get: function() {
      return protocol_typeDefinition_1.TypeDefinitionRequest;
    } });
    var protocol_workspaceFolders_1 = require_protocol_workspaceFolders();
    Object.defineProperty(exports2, "WorkspaceFoldersRequest", { enumerable: true, get: function() {
      return protocol_workspaceFolders_1.WorkspaceFoldersRequest;
    } });
    Object.defineProperty(exports2, "DidChangeWorkspaceFoldersNotification", { enumerable: true, get: function() {
      return protocol_workspaceFolders_1.DidChangeWorkspaceFoldersNotification;
    } });
    var protocol_configuration_1 = require_protocol_configuration();
    Object.defineProperty(exports2, "ConfigurationRequest", { enumerable: true, get: function() {
      return protocol_configuration_1.ConfigurationRequest;
    } });
    var protocol_colorProvider_1 = require_protocol_colorProvider();
    Object.defineProperty(exports2, "DocumentColorRequest", { enumerable: true, get: function() {
      return protocol_colorProvider_1.DocumentColorRequest;
    } });
    Object.defineProperty(exports2, "ColorPresentationRequest", { enumerable: true, get: function() {
      return protocol_colorProvider_1.ColorPresentationRequest;
    } });
    var protocol_foldingRange_1 = require_protocol_foldingRange();
    Object.defineProperty(exports2, "FoldingRangeRequest", { enumerable: true, get: function() {
      return protocol_foldingRange_1.FoldingRangeRequest;
    } });
    var protocol_declaration_1 = require_protocol_declaration();
    Object.defineProperty(exports2, "DeclarationRequest", { enumerable: true, get: function() {
      return protocol_declaration_1.DeclarationRequest;
    } });
    var protocol_selectionRange_1 = require_protocol_selectionRange();
    Object.defineProperty(exports2, "SelectionRangeRequest", { enumerable: true, get: function() {
      return protocol_selectionRange_1.SelectionRangeRequest;
    } });
    var protocol_progress_1 = require_protocol_progress();
    Object.defineProperty(exports2, "WorkDoneProgress", { enumerable: true, get: function() {
      return protocol_progress_1.WorkDoneProgress;
    } });
    Object.defineProperty(exports2, "WorkDoneProgressCreateRequest", { enumerable: true, get: function() {
      return protocol_progress_1.WorkDoneProgressCreateRequest;
    } });
    Object.defineProperty(exports2, "WorkDoneProgressCancelNotification", { enumerable: true, get: function() {
      return protocol_progress_1.WorkDoneProgressCancelNotification;
    } });
    var protocol_callHierarchy_1 = require_protocol_callHierarchy();
    Object.defineProperty(exports2, "CallHierarchyIncomingCallsRequest", { enumerable: true, get: function() {
      return protocol_callHierarchy_1.CallHierarchyIncomingCallsRequest;
    } });
    Object.defineProperty(exports2, "CallHierarchyOutgoingCallsRequest", { enumerable: true, get: function() {
      return protocol_callHierarchy_1.CallHierarchyOutgoingCallsRequest;
    } });
    Object.defineProperty(exports2, "CallHierarchyPrepareRequest", { enumerable: true, get: function() {
      return protocol_callHierarchy_1.CallHierarchyPrepareRequest;
    } });
    var protocol_semanticTokens_1 = require_protocol_semanticTokens();
    Object.defineProperty(exports2, "SemanticTokenTypes", { enumerable: true, get: function() {
      return protocol_semanticTokens_1.SemanticTokenTypes;
    } });
    Object.defineProperty(exports2, "SemanticTokenModifiers", { enumerable: true, get: function() {
      return protocol_semanticTokens_1.SemanticTokenModifiers;
    } });
    Object.defineProperty(exports2, "SemanticTokens", { enumerable: true, get: function() {
      return protocol_semanticTokens_1.SemanticTokens;
    } });
    Object.defineProperty(exports2, "TokenFormat", { enumerable: true, get: function() {
      return protocol_semanticTokens_1.TokenFormat;
    } });
    Object.defineProperty(exports2, "SemanticTokensRequest", { enumerable: true, get: function() {
      return protocol_semanticTokens_1.SemanticTokensRequest;
    } });
    Object.defineProperty(exports2, "SemanticTokensDeltaRequest", { enumerable: true, get: function() {
      return protocol_semanticTokens_1.SemanticTokensDeltaRequest;
    } });
    Object.defineProperty(exports2, "SemanticTokensRangeRequest", { enumerable: true, get: function() {
      return protocol_semanticTokens_1.SemanticTokensRangeRequest;
    } });
    Object.defineProperty(exports2, "SemanticTokensRefreshRequest", { enumerable: true, get: function() {
      return protocol_semanticTokens_1.SemanticTokensRefreshRequest;
    } });
    Object.defineProperty(exports2, "SemanticTokensRegistrationType", { enumerable: true, get: function() {
      return protocol_semanticTokens_1.SemanticTokensRegistrationType;
    } });
    var protocol_showDocument_1 = require_protocol_showDocument();
    Object.defineProperty(exports2, "ShowDocumentRequest", { enumerable: true, get: function() {
      return protocol_showDocument_1.ShowDocumentRequest;
    } });
    var protocol_linkedEditingRange_1 = require_protocol_linkedEditingRange();
    Object.defineProperty(exports2, "LinkedEditingRangeRequest", { enumerable: true, get: function() {
      return protocol_linkedEditingRange_1.LinkedEditingRangeRequest;
    } });
    var protocol_fileOperations_1 = require_protocol_fileOperations();
    Object.defineProperty(exports2, "FileOperationPatternKind", { enumerable: true, get: function() {
      return protocol_fileOperations_1.FileOperationPatternKind;
    } });
    Object.defineProperty(exports2, "DidCreateFilesNotification", { enumerable: true, get: function() {
      return protocol_fileOperations_1.DidCreateFilesNotification;
    } });
    Object.defineProperty(exports2, "WillCreateFilesRequest", { enumerable: true, get: function() {
      return protocol_fileOperations_1.WillCreateFilesRequest;
    } });
    Object.defineProperty(exports2, "DidRenameFilesNotification", { enumerable: true, get: function() {
      return protocol_fileOperations_1.DidRenameFilesNotification;
    } });
    Object.defineProperty(exports2, "WillRenameFilesRequest", { enumerable: true, get: function() {
      return protocol_fileOperations_1.WillRenameFilesRequest;
    } });
    Object.defineProperty(exports2, "DidDeleteFilesNotification", { enumerable: true, get: function() {
      return protocol_fileOperations_1.DidDeleteFilesNotification;
    } });
    Object.defineProperty(exports2, "WillDeleteFilesRequest", { enumerable: true, get: function() {
      return protocol_fileOperations_1.WillDeleteFilesRequest;
    } });
    var protocol_moniker_1 = require_protocol_moniker();
    Object.defineProperty(exports2, "UniquenessLevel", { enumerable: true, get: function() {
      return protocol_moniker_1.UniquenessLevel;
    } });
    Object.defineProperty(exports2, "MonikerKind", { enumerable: true, get: function() {
      return protocol_moniker_1.MonikerKind;
    } });
    Object.defineProperty(exports2, "MonikerRequest", { enumerable: true, get: function() {
      return protocol_moniker_1.MonikerRequest;
    } });
    var DocumentFilter;
    (function(DocumentFilter2) {
      function is(value) {
        const candidate = value;
        return Is.string(candidate.language) || Is.string(candidate.scheme) || Is.string(candidate.pattern);
      }
      DocumentFilter2.is = is;
    })(DocumentFilter = exports2.DocumentFilter || (exports2.DocumentFilter = {}));
    var DocumentSelector;
    (function(DocumentSelector2) {
      function is(value) {
        if (!Array.isArray(value)) {
          return false;
        }
        for (let elem of value) {
          if (!Is.string(elem) && !DocumentFilter.is(elem)) {
            return false;
          }
        }
        return true;
      }
      DocumentSelector2.is = is;
    })(DocumentSelector = exports2.DocumentSelector || (exports2.DocumentSelector = {}));
    var RegistrationRequest;
    (function(RegistrationRequest2) {
      RegistrationRequest2.type = new messages_1.ProtocolRequestType("client/registerCapability");
    })(RegistrationRequest = exports2.RegistrationRequest || (exports2.RegistrationRequest = {}));
    var UnregistrationRequest;
    (function(UnregistrationRequest2) {
      UnregistrationRequest2.type = new messages_1.ProtocolRequestType("client/unregisterCapability");
    })(UnregistrationRequest = exports2.UnregistrationRequest || (exports2.UnregistrationRequest = {}));
    var ResourceOperationKind;
    (function(ResourceOperationKind2) {
      ResourceOperationKind2.Create = "create";
      ResourceOperationKind2.Rename = "rename";
      ResourceOperationKind2.Delete = "delete";
    })(ResourceOperationKind = exports2.ResourceOperationKind || (exports2.ResourceOperationKind = {}));
    var FailureHandlingKind;
    (function(FailureHandlingKind2) {
      FailureHandlingKind2.Abort = "abort";
      FailureHandlingKind2.Transactional = "transactional";
      FailureHandlingKind2.TextOnlyTransactional = "textOnlyTransactional";
      FailureHandlingKind2.Undo = "undo";
    })(FailureHandlingKind = exports2.FailureHandlingKind || (exports2.FailureHandlingKind = {}));
    var StaticRegistrationOptions;
    (function(StaticRegistrationOptions2) {
      function hasId(value) {
        const candidate = value;
        return candidate && Is.string(candidate.id) && candidate.id.length > 0;
      }
      StaticRegistrationOptions2.hasId = hasId;
    })(StaticRegistrationOptions = exports2.StaticRegistrationOptions || (exports2.StaticRegistrationOptions = {}));
    var TextDocumentRegistrationOptions;
    (function(TextDocumentRegistrationOptions2) {
      function is(value) {
        const candidate = value;
        return candidate && (candidate.documentSelector === null || DocumentSelector.is(candidate.documentSelector));
      }
      TextDocumentRegistrationOptions2.is = is;
    })(TextDocumentRegistrationOptions = exports2.TextDocumentRegistrationOptions || (exports2.TextDocumentRegistrationOptions = {}));
    var WorkDoneProgressOptions;
    (function(WorkDoneProgressOptions2) {
      function is(value) {
        const candidate = value;
        return Is.objectLiteral(candidate) && (candidate.workDoneProgress === void 0 || Is.boolean(candidate.workDoneProgress));
      }
      WorkDoneProgressOptions2.is = is;
      function hasWorkDoneProgress(value) {
        const candidate = value;
        return candidate && Is.boolean(candidate.workDoneProgress);
      }
      WorkDoneProgressOptions2.hasWorkDoneProgress = hasWorkDoneProgress;
    })(WorkDoneProgressOptions = exports2.WorkDoneProgressOptions || (exports2.WorkDoneProgressOptions = {}));
    var InitializeRequest;
    (function(InitializeRequest2) {
      InitializeRequest2.type = new messages_1.ProtocolRequestType("initialize");
    })(InitializeRequest = exports2.InitializeRequest || (exports2.InitializeRequest = {}));
    var InitializeError;
    (function(InitializeError2) {
      InitializeError2.unknownProtocolVersion = 1;
    })(InitializeError = exports2.InitializeError || (exports2.InitializeError = {}));
    var InitializedNotification;
    (function(InitializedNotification2) {
      InitializedNotification2.type = new messages_1.ProtocolNotificationType("initialized");
    })(InitializedNotification = exports2.InitializedNotification || (exports2.InitializedNotification = {}));
    var ShutdownRequest;
    (function(ShutdownRequest2) {
      ShutdownRequest2.type = new messages_1.ProtocolRequestType0("shutdown");
    })(ShutdownRequest = exports2.ShutdownRequest || (exports2.ShutdownRequest = {}));
    var ExitNotification;
    (function(ExitNotification2) {
      ExitNotification2.type = new messages_1.ProtocolNotificationType0("exit");
    })(ExitNotification = exports2.ExitNotification || (exports2.ExitNotification = {}));
    var DidChangeConfigurationNotification;
    (function(DidChangeConfigurationNotification2) {
      DidChangeConfigurationNotification2.type = new messages_1.ProtocolNotificationType("workspace/didChangeConfiguration");
    })(DidChangeConfigurationNotification = exports2.DidChangeConfigurationNotification || (exports2.DidChangeConfigurationNotification = {}));
    var MessageType;
    (function(MessageType2) {
      MessageType2.Error = 1;
      MessageType2.Warning = 2;
      MessageType2.Info = 3;
      MessageType2.Log = 4;
    })(MessageType = exports2.MessageType || (exports2.MessageType = {}));
    var ShowMessageNotification;
    (function(ShowMessageNotification2) {
      ShowMessageNotification2.type = new messages_1.ProtocolNotificationType("window/showMessage");
    })(ShowMessageNotification = exports2.ShowMessageNotification || (exports2.ShowMessageNotification = {}));
    var ShowMessageRequest;
    (function(ShowMessageRequest2) {
      ShowMessageRequest2.type = new messages_1.ProtocolRequestType("window/showMessageRequest");
    })(ShowMessageRequest = exports2.ShowMessageRequest || (exports2.ShowMessageRequest = {}));
    var LogMessageNotification;
    (function(LogMessageNotification2) {
      LogMessageNotification2.type = new messages_1.ProtocolNotificationType("window/logMessage");
    })(LogMessageNotification = exports2.LogMessageNotification || (exports2.LogMessageNotification = {}));
    var TelemetryEventNotification;
    (function(TelemetryEventNotification2) {
      TelemetryEventNotification2.type = new messages_1.ProtocolNotificationType("telemetry/event");
    })(TelemetryEventNotification = exports2.TelemetryEventNotification || (exports2.TelemetryEventNotification = {}));
    var TextDocumentSyncKind;
    (function(TextDocumentSyncKind2) {
      TextDocumentSyncKind2.None = 0;
      TextDocumentSyncKind2.Full = 1;
      TextDocumentSyncKind2.Incremental = 2;
    })(TextDocumentSyncKind = exports2.TextDocumentSyncKind || (exports2.TextDocumentSyncKind = {}));
    var DidOpenTextDocumentNotification;
    (function(DidOpenTextDocumentNotification2) {
      DidOpenTextDocumentNotification2.method = "textDocument/didOpen";
      DidOpenTextDocumentNotification2.type = new messages_1.ProtocolNotificationType(DidOpenTextDocumentNotification2.method);
    })(DidOpenTextDocumentNotification = exports2.DidOpenTextDocumentNotification || (exports2.DidOpenTextDocumentNotification = {}));
    var TextDocumentContentChangeEvent;
    (function(TextDocumentContentChangeEvent2) {
      function isIncremental(event) {
        let candidate = event;
        return candidate !== void 0 && candidate !== null && typeof candidate.text === "string" && candidate.range !== void 0 && (candidate.rangeLength === void 0 || typeof candidate.rangeLength === "number");
      }
      TextDocumentContentChangeEvent2.isIncremental = isIncremental;
      function isFull(event) {
        let candidate = event;
        return candidate !== void 0 && candidate !== null && typeof candidate.text === "string" && candidate.range === void 0 && candidate.rangeLength === void 0;
      }
      TextDocumentContentChangeEvent2.isFull = isFull;
    })(TextDocumentContentChangeEvent = exports2.TextDocumentContentChangeEvent || (exports2.TextDocumentContentChangeEvent = {}));
    var DidChangeTextDocumentNotification;
    (function(DidChangeTextDocumentNotification2) {
      DidChangeTextDocumentNotification2.method = "textDocument/didChange";
      DidChangeTextDocumentNotification2.type = new messages_1.ProtocolNotificationType(DidChangeTextDocumentNotification2.method);
    })(DidChangeTextDocumentNotification = exports2.DidChangeTextDocumentNotification || (exports2.DidChangeTextDocumentNotification = {}));
    var DidCloseTextDocumentNotification;
    (function(DidCloseTextDocumentNotification2) {
      DidCloseTextDocumentNotification2.method = "textDocument/didClose";
      DidCloseTextDocumentNotification2.type = new messages_1.ProtocolNotificationType(DidCloseTextDocumentNotification2.method);
    })(DidCloseTextDocumentNotification = exports2.DidCloseTextDocumentNotification || (exports2.DidCloseTextDocumentNotification = {}));
    var DidSaveTextDocumentNotification;
    (function(DidSaveTextDocumentNotification2) {
      DidSaveTextDocumentNotification2.method = "textDocument/didSave";
      DidSaveTextDocumentNotification2.type = new messages_1.ProtocolNotificationType(DidSaveTextDocumentNotification2.method);
    })(DidSaveTextDocumentNotification = exports2.DidSaveTextDocumentNotification || (exports2.DidSaveTextDocumentNotification = {}));
    var TextDocumentSaveReason;
    (function(TextDocumentSaveReason2) {
      TextDocumentSaveReason2.Manual = 1;
      TextDocumentSaveReason2.AfterDelay = 2;
      TextDocumentSaveReason2.FocusOut = 3;
    })(TextDocumentSaveReason = exports2.TextDocumentSaveReason || (exports2.TextDocumentSaveReason = {}));
    var WillSaveTextDocumentNotification;
    (function(WillSaveTextDocumentNotification2) {
      WillSaveTextDocumentNotification2.method = "textDocument/willSave";
      WillSaveTextDocumentNotification2.type = new messages_1.ProtocolNotificationType(WillSaveTextDocumentNotification2.method);
    })(WillSaveTextDocumentNotification = exports2.WillSaveTextDocumentNotification || (exports2.WillSaveTextDocumentNotification = {}));
    var WillSaveTextDocumentWaitUntilRequest;
    (function(WillSaveTextDocumentWaitUntilRequest2) {
      WillSaveTextDocumentWaitUntilRequest2.method = "textDocument/willSaveWaitUntil";
      WillSaveTextDocumentWaitUntilRequest2.type = new messages_1.ProtocolRequestType(WillSaveTextDocumentWaitUntilRequest2.method);
    })(WillSaveTextDocumentWaitUntilRequest = exports2.WillSaveTextDocumentWaitUntilRequest || (exports2.WillSaveTextDocumentWaitUntilRequest = {}));
    var DidChangeWatchedFilesNotification;
    (function(DidChangeWatchedFilesNotification2) {
      DidChangeWatchedFilesNotification2.type = new messages_1.ProtocolNotificationType("workspace/didChangeWatchedFiles");
    })(DidChangeWatchedFilesNotification = exports2.DidChangeWatchedFilesNotification || (exports2.DidChangeWatchedFilesNotification = {}));
    var FileChangeType;
    (function(FileChangeType2) {
      FileChangeType2.Created = 1;
      FileChangeType2.Changed = 2;
      FileChangeType2.Deleted = 3;
    })(FileChangeType = exports2.FileChangeType || (exports2.FileChangeType = {}));
    var WatchKind;
    (function(WatchKind2) {
      WatchKind2.Create = 1;
      WatchKind2.Change = 2;
      WatchKind2.Delete = 4;
    })(WatchKind = exports2.WatchKind || (exports2.WatchKind = {}));
    var PublishDiagnosticsNotification;
    (function(PublishDiagnosticsNotification2) {
      PublishDiagnosticsNotification2.type = new messages_1.ProtocolNotificationType("textDocument/publishDiagnostics");
    })(PublishDiagnosticsNotification = exports2.PublishDiagnosticsNotification || (exports2.PublishDiagnosticsNotification = {}));
    var CompletionTriggerKind;
    (function(CompletionTriggerKind2) {
      CompletionTriggerKind2.Invoked = 1;
      CompletionTriggerKind2.TriggerCharacter = 2;
      CompletionTriggerKind2.TriggerForIncompleteCompletions = 3;
    })(CompletionTriggerKind = exports2.CompletionTriggerKind || (exports2.CompletionTriggerKind = {}));
    var CompletionRequest;
    (function(CompletionRequest2) {
      CompletionRequest2.method = "textDocument/completion";
      CompletionRequest2.type = new messages_1.ProtocolRequestType(CompletionRequest2.method);
    })(CompletionRequest = exports2.CompletionRequest || (exports2.CompletionRequest = {}));
    var CompletionResolveRequest;
    (function(CompletionResolveRequest2) {
      CompletionResolveRequest2.method = "completionItem/resolve";
      CompletionResolveRequest2.type = new messages_1.ProtocolRequestType(CompletionResolveRequest2.method);
    })(CompletionResolveRequest = exports2.CompletionResolveRequest || (exports2.CompletionResolveRequest = {}));
    var HoverRequest;
    (function(HoverRequest2) {
      HoverRequest2.method = "textDocument/hover";
      HoverRequest2.type = new messages_1.ProtocolRequestType(HoverRequest2.method);
    })(HoverRequest = exports2.HoverRequest || (exports2.HoverRequest = {}));
    var SignatureHelpTriggerKind;
    (function(SignatureHelpTriggerKind2) {
      SignatureHelpTriggerKind2.Invoked = 1;
      SignatureHelpTriggerKind2.TriggerCharacter = 2;
      SignatureHelpTriggerKind2.ContentChange = 3;
    })(SignatureHelpTriggerKind = exports2.SignatureHelpTriggerKind || (exports2.SignatureHelpTriggerKind = {}));
    var SignatureHelpRequest;
    (function(SignatureHelpRequest2) {
      SignatureHelpRequest2.method = "textDocument/signatureHelp";
      SignatureHelpRequest2.type = new messages_1.ProtocolRequestType(SignatureHelpRequest2.method);
    })(SignatureHelpRequest = exports2.SignatureHelpRequest || (exports2.SignatureHelpRequest = {}));
    var DefinitionRequest;
    (function(DefinitionRequest2) {
      DefinitionRequest2.method = "textDocument/definition";
      DefinitionRequest2.type = new messages_1.ProtocolRequestType(DefinitionRequest2.method);
    })(DefinitionRequest = exports2.DefinitionRequest || (exports2.DefinitionRequest = {}));
    var ReferencesRequest;
    (function(ReferencesRequest2) {
      ReferencesRequest2.method = "textDocument/references";
      ReferencesRequest2.type = new messages_1.ProtocolRequestType(ReferencesRequest2.method);
    })(ReferencesRequest = exports2.ReferencesRequest || (exports2.ReferencesRequest = {}));
    var DocumentHighlightRequest;
    (function(DocumentHighlightRequest2) {
      DocumentHighlightRequest2.method = "textDocument/documentHighlight";
      DocumentHighlightRequest2.type = new messages_1.ProtocolRequestType(DocumentHighlightRequest2.method);
    })(DocumentHighlightRequest = exports2.DocumentHighlightRequest || (exports2.DocumentHighlightRequest = {}));
    var DocumentSymbolRequest;
    (function(DocumentSymbolRequest2) {
      DocumentSymbolRequest2.method = "textDocument/documentSymbol";
      DocumentSymbolRequest2.type = new messages_1.ProtocolRequestType(DocumentSymbolRequest2.method);
    })(DocumentSymbolRequest = exports2.DocumentSymbolRequest || (exports2.DocumentSymbolRequest = {}));
    var CodeActionRequest;
    (function(CodeActionRequest2) {
      CodeActionRequest2.method = "textDocument/codeAction";
      CodeActionRequest2.type = new messages_1.ProtocolRequestType(CodeActionRequest2.method);
    })(CodeActionRequest = exports2.CodeActionRequest || (exports2.CodeActionRequest = {}));
    var CodeActionResolveRequest;
    (function(CodeActionResolveRequest2) {
      CodeActionResolveRequest2.method = "codeAction/resolve";
      CodeActionResolveRequest2.type = new messages_1.ProtocolRequestType(CodeActionResolveRequest2.method);
    })(CodeActionResolveRequest = exports2.CodeActionResolveRequest || (exports2.CodeActionResolveRequest = {}));
    var WorkspaceSymbolRequest;
    (function(WorkspaceSymbolRequest2) {
      WorkspaceSymbolRequest2.method = "workspace/symbol";
      WorkspaceSymbolRequest2.type = new messages_1.ProtocolRequestType(WorkspaceSymbolRequest2.method);
    })(WorkspaceSymbolRequest = exports2.WorkspaceSymbolRequest || (exports2.WorkspaceSymbolRequest = {}));
    var CodeLensRequest;
    (function(CodeLensRequest2) {
      CodeLensRequest2.method = "textDocument/codeLens";
      CodeLensRequest2.type = new messages_1.ProtocolRequestType(CodeLensRequest2.method);
    })(CodeLensRequest = exports2.CodeLensRequest || (exports2.CodeLensRequest = {}));
    var CodeLensResolveRequest;
    (function(CodeLensResolveRequest2) {
      CodeLensResolveRequest2.method = "codeLens/resolve";
      CodeLensResolveRequest2.type = new messages_1.ProtocolRequestType(CodeLensResolveRequest2.method);
    })(CodeLensResolveRequest = exports2.CodeLensResolveRequest || (exports2.CodeLensResolveRequest = {}));
    var CodeLensRefreshRequest;
    (function(CodeLensRefreshRequest2) {
      CodeLensRefreshRequest2.method = `workspace/codeLens/refresh`;
      CodeLensRefreshRequest2.type = new messages_1.ProtocolRequestType0(CodeLensRefreshRequest2.method);
    })(CodeLensRefreshRequest = exports2.CodeLensRefreshRequest || (exports2.CodeLensRefreshRequest = {}));
    var DocumentLinkRequest;
    (function(DocumentLinkRequest2) {
      DocumentLinkRequest2.method = "textDocument/documentLink";
      DocumentLinkRequest2.type = new messages_1.ProtocolRequestType(DocumentLinkRequest2.method);
    })(DocumentLinkRequest = exports2.DocumentLinkRequest || (exports2.DocumentLinkRequest = {}));
    var DocumentLinkResolveRequest;
    (function(DocumentLinkResolveRequest2) {
      DocumentLinkResolveRequest2.method = "documentLink/resolve";
      DocumentLinkResolveRequest2.type = new messages_1.ProtocolRequestType(DocumentLinkResolveRequest2.method);
    })(DocumentLinkResolveRequest = exports2.DocumentLinkResolveRequest || (exports2.DocumentLinkResolveRequest = {}));
    var DocumentFormattingRequest;
    (function(DocumentFormattingRequest2) {
      DocumentFormattingRequest2.method = "textDocument/formatting";
      DocumentFormattingRequest2.type = new messages_1.ProtocolRequestType(DocumentFormattingRequest2.method);
    })(DocumentFormattingRequest = exports2.DocumentFormattingRequest || (exports2.DocumentFormattingRequest = {}));
    var DocumentRangeFormattingRequest;
    (function(DocumentRangeFormattingRequest2) {
      DocumentRangeFormattingRequest2.method = "textDocument/rangeFormatting";
      DocumentRangeFormattingRequest2.type = new messages_1.ProtocolRequestType(DocumentRangeFormattingRequest2.method);
    })(DocumentRangeFormattingRequest = exports2.DocumentRangeFormattingRequest || (exports2.DocumentRangeFormattingRequest = {}));
    var DocumentOnTypeFormattingRequest;
    (function(DocumentOnTypeFormattingRequest2) {
      DocumentOnTypeFormattingRequest2.method = "textDocument/onTypeFormatting";
      DocumentOnTypeFormattingRequest2.type = new messages_1.ProtocolRequestType(DocumentOnTypeFormattingRequest2.method);
    })(DocumentOnTypeFormattingRequest = exports2.DocumentOnTypeFormattingRequest || (exports2.DocumentOnTypeFormattingRequest = {}));
    var PrepareSupportDefaultBehavior;
    (function(PrepareSupportDefaultBehavior2) {
      PrepareSupportDefaultBehavior2.Identifier = 1;
    })(PrepareSupportDefaultBehavior = exports2.PrepareSupportDefaultBehavior || (exports2.PrepareSupportDefaultBehavior = {}));
    var RenameRequest;
    (function(RenameRequest2) {
      RenameRequest2.method = "textDocument/rename";
      RenameRequest2.type = new messages_1.ProtocolRequestType(RenameRequest2.method);
    })(RenameRequest = exports2.RenameRequest || (exports2.RenameRequest = {}));
    var PrepareRenameRequest;
    (function(PrepareRenameRequest2) {
      PrepareRenameRequest2.method = "textDocument/prepareRename";
      PrepareRenameRequest2.type = new messages_1.ProtocolRequestType(PrepareRenameRequest2.method);
    })(PrepareRenameRequest = exports2.PrepareRenameRequest || (exports2.PrepareRenameRequest = {}));
    var ExecuteCommandRequest;
    (function(ExecuteCommandRequest2) {
      ExecuteCommandRequest2.type = new messages_1.ProtocolRequestType("workspace/executeCommand");
    })(ExecuteCommandRequest = exports2.ExecuteCommandRequest || (exports2.ExecuteCommandRequest = {}));
    var ApplyWorkspaceEditRequest;
    (function(ApplyWorkspaceEditRequest2) {
      ApplyWorkspaceEditRequest2.type = new messages_1.ProtocolRequestType("workspace/applyEdit");
    })(ApplyWorkspaceEditRequest = exports2.ApplyWorkspaceEditRequest || (exports2.ApplyWorkspaceEditRequest = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/connection.js
var require_connection2 = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/connection.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.createProtocolConnection = void 0;
    var vscode_jsonrpc_1 = require_main();
    function createProtocolConnection(input, output, logger, options) {
      if (vscode_jsonrpc_1.ConnectionStrategy.is(options)) {
        options = { connectionStrategy: options };
      }
      return vscode_jsonrpc_1.createMessageConnection(input, output, logger, options);
    }
    exports2.createProtocolConnection = createProtocolConnection;
  }
});

// node_modules/vscode-languageserver-protocol/lib/common/api.js
var require_api2 = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/common/api.js"(exports2) {
    "use strict";
    var __createBinding = exports2 && exports2.__createBinding || (Object.create ? function(o, m, k, k2) {
      if (k2 === void 0) k2 = k;
      Object.defineProperty(o, k2, { enumerable: true, get: function() {
        return m[k];
      } });
    } : function(o, m, k, k2) {
      if (k2 === void 0) k2 = k;
      o[k2] = m[k];
    });
    var __exportStar = exports2 && exports2.__exportStar || function(m, exports3) {
      for (var p in m) if (p !== "default" && !Object.prototype.hasOwnProperty.call(exports3, p)) __createBinding(exports3, m, p);
    };
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.LSPErrorCodes = exports2.createProtocolConnection = void 0;
    __exportStar(require_main(), exports2);
    __exportStar(require_main2(), exports2);
    __exportStar(require_messages2(), exports2);
    __exportStar(require_protocol(), exports2);
    var connection_1 = require_connection2();
    Object.defineProperty(exports2, "createProtocolConnection", { enumerable: true, get: function() {
      return connection_1.createProtocolConnection;
    } });
    var LSPErrorCodes;
    (function(LSPErrorCodes2) {
      LSPErrorCodes2.lspReservedErrorRangeStart = -32899;
      LSPErrorCodes2.ContentModified = -32801;
      LSPErrorCodes2.RequestCancelled = -32800;
      LSPErrorCodes2.lspReservedErrorRangeEnd = -32800;
    })(LSPErrorCodes = exports2.LSPErrorCodes || (exports2.LSPErrorCodes = {}));
  }
});

// node_modules/vscode-languageserver-protocol/lib/node/main.js
var require_main3 = __commonJS({
  "node_modules/vscode-languageserver-protocol/lib/node/main.js"(exports2) {
    "use strict";
    var __createBinding = exports2 && exports2.__createBinding || (Object.create ? function(o, m, k, k2) {
      if (k2 === void 0) k2 = k;
      Object.defineProperty(o, k2, { enumerable: true, get: function() {
        return m[k];
      } });
    } : function(o, m, k, k2) {
      if (k2 === void 0) k2 = k;
      o[k2] = m[k];
    });
    var __exportStar = exports2 && exports2.__exportStar || function(m, exports3) {
      for (var p in m) if (p !== "default" && !Object.prototype.hasOwnProperty.call(exports3, p)) __createBinding(exports3, m, p);
    };
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.createProtocolConnection = void 0;
    var node_1 = require_node();
    __exportStar(require_node(), exports2);
    __exportStar(require_api2(), exports2);
    function createProtocolConnection(input, output, logger, options) {
      return node_1.createMessageConnection(input, output, logger, options);
    }
    exports2.createProtocolConnection = createProtocolConnection;
  }
});

// node_modules/vscode-languageclient/lib/common/configuration.js
var require_configuration = __commonJS({
  "node_modules/vscode-languageclient/lib/common/configuration.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.toJSONObject = exports2.ConfigurationFeature = void 0;
    var vscode_1 = require("vscode");
    var vscode_languageserver_protocol_1 = require_main3();
    var ConfigurationFeature = class {
      constructor(_client) {
        this._client = _client;
      }
      fillClientCapabilities(capabilities) {
        capabilities.workspace = capabilities.workspace || {};
        capabilities.workspace.configuration = true;
      }
      initialize() {
        let client2 = this._client;
        client2.onRequest(vscode_languageserver_protocol_1.ConfigurationRequest.type, (params, token) => {
          let configuration = (params2) => {
            let result = [];
            for (let item of params2.items) {
              let resource = item.scopeUri !== void 0 && item.scopeUri !== null ? this._client.protocol2CodeConverter.asUri(item.scopeUri) : void 0;
              result.push(this.getConfiguration(resource, item.section !== null ? item.section : void 0));
            }
            return result;
          };
          let middleware = client2.clientOptions.middleware.workspace;
          return middleware && middleware.configuration ? middleware.configuration(params, token, configuration) : configuration(params, token);
        });
      }
      getConfiguration(resource, section) {
        let result = null;
        if (section) {
          let index = section.lastIndexOf(".");
          if (index === -1) {
            result = toJSONObject(vscode_1.workspace.getConfiguration(void 0, resource).get(section));
          } else {
            let config = vscode_1.workspace.getConfiguration(section.substr(0, index), resource);
            if (config) {
              result = toJSONObject(config.get(section.substr(index + 1)));
            }
          }
        } else {
          let config = vscode_1.workspace.getConfiguration(void 0, resource);
          result = {};
          for (let key of Object.keys(config)) {
            if (config.has(key)) {
              result[key] = toJSONObject(config.get(key));
            }
          }
        }
        if (result === void 0) {
          result = null;
        }
        return result;
      }
      dispose() {
      }
    };
    exports2.ConfigurationFeature = ConfigurationFeature;
    function toJSONObject(obj) {
      if (obj) {
        if (Array.isArray(obj)) {
          return obj.map(toJSONObject);
        } else if (typeof obj === "object") {
          const res = /* @__PURE__ */ Object.create(null);
          for (const key in obj) {
            if (Object.prototype.hasOwnProperty.call(obj, key)) {
              res[key] = toJSONObject(obj[key]);
            }
          }
          return res;
        }
      }
      return obj;
    }
    exports2.toJSONObject = toJSONObject;
  }
});

// node_modules/vscode-languageclient/lib/common/protocolCompletionItem.js
var require_protocolCompletionItem = __commonJS({
  "node_modules/vscode-languageclient/lib/common/protocolCompletionItem.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    var code = require("vscode");
    var ProtocolCompletionItem = class extends code.CompletionItem {
      constructor(label) {
        super(label);
      }
    };
    exports2.default = ProtocolCompletionItem;
  }
});

// node_modules/vscode-languageclient/lib/common/protocolCodeLens.js
var require_protocolCodeLens = __commonJS({
  "node_modules/vscode-languageclient/lib/common/protocolCodeLens.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    var code = require("vscode");
    var ProtocolCodeLens = class extends code.CodeLens {
      constructor(range) {
        super(range);
      }
    };
    exports2.default = ProtocolCodeLens;
  }
});

// node_modules/vscode-languageclient/lib/common/protocolDocumentLink.js
var require_protocolDocumentLink = __commonJS({
  "node_modules/vscode-languageclient/lib/common/protocolDocumentLink.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    var code = require("vscode");
    var ProtocolDocumentLink = class extends code.DocumentLink {
      constructor(range, target) {
        super(range, target);
      }
    };
    exports2.default = ProtocolDocumentLink;
  }
});

// node_modules/vscode-languageclient/lib/common/protocolCodeAction.js
var require_protocolCodeAction = __commonJS({
  "node_modules/vscode-languageclient/lib/common/protocolCodeAction.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    var vscode = require("vscode");
    var ProtocolCodeAction = class extends vscode.CodeAction {
      constructor(title, data) {
        super(title);
        this.data = data;
      }
    };
    exports2.default = ProtocolCodeAction;
  }
});

// node_modules/vscode-languageclient/lib/common/protocolDiagnostic.js
var require_protocolDiagnostic = __commonJS({
  "node_modules/vscode-languageclient/lib/common/protocolDiagnostic.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.ProtocolDiagnostic = exports2.DiagnosticCode = void 0;
    var vscode = require("vscode");
    var Is = require_is();
    var DiagnosticCode;
    (function(DiagnosticCode2) {
      function is(value) {
        const candidate = value;
        return candidate !== void 0 && candidate !== null && (Is.number(candidate.value) || Is.string(candidate.value)) && Is.string(candidate.target);
      }
      DiagnosticCode2.is = is;
    })(DiagnosticCode = exports2.DiagnosticCode || (exports2.DiagnosticCode = {}));
    var ProtocolDiagnostic = class extends vscode.Diagnostic {
      constructor(range, message, severity, data) {
        super(range, message, severity);
        this.data = data;
        this.hasDiagnosticCode = false;
      }
    };
    exports2.ProtocolDiagnostic = ProtocolDiagnostic;
  }
});

// node_modules/vscode-languageclient/lib/common/protocolCallHierarchyItem.js
var require_protocolCallHierarchyItem = __commonJS({
  "node_modules/vscode-languageclient/lib/common/protocolCallHierarchyItem.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    var code = require("vscode");
    var ProtocolCallHierarchyItem = class extends code.CallHierarchyItem {
      constructor(kind, name, detail, uri, range, selectionRange, data) {
        super(kind, name, detail, uri, range, selectionRange);
        if (data !== void 0) {
          this.data = data;
        }
      }
    };
    exports2.default = ProtocolCallHierarchyItem;
  }
});

// node_modules/vscode-languageclient/lib/common/codeConverter.js
var require_codeConverter = __commonJS({
  "node_modules/vscode-languageclient/lib/common/codeConverter.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.createConverter = void 0;
    var code = require("vscode");
    var proto = require_main3();
    var Is = require_is();
    var protocolCompletionItem_1 = require_protocolCompletionItem();
    var protocolCodeLens_1 = require_protocolCodeLens();
    var protocolDocumentLink_1 = require_protocolDocumentLink();
    var protocolCodeAction_1 = require_protocolCodeAction();
    var protocolDiagnostic_1 = require_protocolDiagnostic();
    var protocolCallHierarchyItem_1 = require_protocolCallHierarchyItem();
    var vscode_languageserver_protocol_1 = require_main3();
    var InsertReplaceRange;
    (function(InsertReplaceRange2) {
      function is(value) {
        const candidate = value;
        return candidate && !!candidate.inserting && !!candidate.replacing;
      }
      InsertReplaceRange2.is = is;
    })(InsertReplaceRange || (InsertReplaceRange = {}));
    function createConverter(uriConverter) {
      const nullConverter = (value) => value.toString();
      const _uriConverter = uriConverter || nullConverter;
      function asUri(value) {
        return _uriConverter(value);
      }
      function asTextDocumentIdentifier(textDocument) {
        return {
          uri: _uriConverter(textDocument.uri)
        };
      }
      function asVersionedTextDocumentIdentifier(textDocument) {
        return {
          uri: _uriConverter(textDocument.uri),
          version: textDocument.version
        };
      }
      function asOpenTextDocumentParams(textDocument) {
        return {
          textDocument: {
            uri: _uriConverter(textDocument.uri),
            languageId: textDocument.languageId,
            version: textDocument.version,
            text: textDocument.getText()
          }
        };
      }
      function isTextDocumentChangeEvent(value) {
        let candidate = value;
        return !!candidate.document && !!candidate.contentChanges;
      }
      function isTextDocument(value) {
        let candidate = value;
        return !!candidate.uri && !!candidate.version;
      }
      function asChangeTextDocumentParams(arg) {
        if (isTextDocument(arg)) {
          let result = {
            textDocument: {
              uri: _uriConverter(arg.uri),
              version: arg.version
            },
            contentChanges: [{ text: arg.getText() }]
          };
          return result;
        } else if (isTextDocumentChangeEvent(arg)) {
          let document = arg.document;
          let result = {
            textDocument: {
              uri: _uriConverter(document.uri),
              version: document.version
            },
            contentChanges: arg.contentChanges.map((change) => {
              let range = change.range;
              return {
                range: {
                  start: { line: range.start.line, character: range.start.character },
                  end: { line: range.end.line, character: range.end.character }
                },
                rangeLength: change.rangeLength,
                text: change.text
              };
            })
          };
          return result;
        } else {
          throw Error("Unsupported text document change parameter");
        }
      }
      function asCloseTextDocumentParams(textDocument) {
        return {
          textDocument: asTextDocumentIdentifier(textDocument)
        };
      }
      function asSaveTextDocumentParams(textDocument, includeContent = false) {
        let result = {
          textDocument: asTextDocumentIdentifier(textDocument)
        };
        if (includeContent) {
          result.text = textDocument.getText();
        }
        return result;
      }
      function asTextDocumentSaveReason(reason) {
        switch (reason) {
          case code.TextDocumentSaveReason.Manual:
            return proto.TextDocumentSaveReason.Manual;
          case code.TextDocumentSaveReason.AfterDelay:
            return proto.TextDocumentSaveReason.AfterDelay;
          case code.TextDocumentSaveReason.FocusOut:
            return proto.TextDocumentSaveReason.FocusOut;
        }
        return proto.TextDocumentSaveReason.Manual;
      }
      function asWillSaveTextDocumentParams(event) {
        return {
          textDocument: asTextDocumentIdentifier(event.document),
          reason: asTextDocumentSaveReason(event.reason)
        };
      }
      function asDidCreateFilesParams(event) {
        return {
          files: event.files.map((fileUri) => ({
            uri: _uriConverter(fileUri)
          }))
        };
      }
      function asDidRenameFilesParams(event) {
        return {
          files: event.files.map((file) => ({
            oldUri: _uriConverter(file.oldUri),
            newUri: _uriConverter(file.newUri)
          }))
        };
      }
      function asDidDeleteFilesParams(event) {
        return {
          files: event.files.map((fileUri) => ({
            uri: _uriConverter(fileUri)
          }))
        };
      }
      function asWillCreateFilesParams(event) {
        return {
          files: event.files.map((fileUri) => ({
            uri: _uriConverter(fileUri)
          }))
        };
      }
      function asWillRenameFilesParams(event) {
        return {
          files: event.files.map((file) => ({
            oldUri: _uriConverter(file.oldUri),
            newUri: _uriConverter(file.newUri)
          }))
        };
      }
      function asWillDeleteFilesParams(event) {
        return {
          files: event.files.map((fileUri) => ({
            uri: _uriConverter(fileUri)
          }))
        };
      }
      function asTextDocumentPositionParams(textDocument, position) {
        return {
          textDocument: asTextDocumentIdentifier(textDocument),
          position: asWorkerPosition(position)
        };
      }
      function asCompletionTriggerKind(triggerKind) {
        switch (triggerKind) {
          case code.CompletionTriggerKind.TriggerCharacter:
            return proto.CompletionTriggerKind.TriggerCharacter;
          case code.CompletionTriggerKind.TriggerForIncompleteCompletions:
            return proto.CompletionTriggerKind.TriggerForIncompleteCompletions;
          default:
            return proto.CompletionTriggerKind.Invoked;
        }
      }
      function asCompletionParams(textDocument, position, context) {
        return {
          textDocument: asTextDocumentIdentifier(textDocument),
          position: asWorkerPosition(position),
          context: {
            triggerKind: asCompletionTriggerKind(context.triggerKind),
            triggerCharacter: context.triggerCharacter
          }
        };
      }
      function asSignatureHelpTriggerKind(triggerKind) {
        switch (triggerKind) {
          case code.SignatureHelpTriggerKind.Invoke:
            return proto.SignatureHelpTriggerKind.Invoked;
          case code.SignatureHelpTriggerKind.TriggerCharacter:
            return proto.SignatureHelpTriggerKind.TriggerCharacter;
          case code.SignatureHelpTriggerKind.ContentChange:
            return proto.SignatureHelpTriggerKind.ContentChange;
        }
      }
      function asParameterInformation(value) {
        return {
          label: value.label
        };
      }
      function asParameterInformations(values) {
        return values.map(asParameterInformation);
      }
      function asSignatureInformation(value) {
        return {
          label: value.label,
          parameters: asParameterInformations(value.parameters)
        };
      }
      function asSignatureInformations(values) {
        return values.map(asSignatureInformation);
      }
      function asSignatureHelp(value) {
        if (value === void 0) {
          return value;
        }
        return {
          signatures: asSignatureInformations(value.signatures),
          activeSignature: value.activeSignature,
          activeParameter: value.activeParameter
        };
      }
      function asSignatureHelpParams(textDocument, position, context) {
        return {
          textDocument: asTextDocumentIdentifier(textDocument),
          position: asWorkerPosition(position),
          context: {
            isRetrigger: context.isRetrigger,
            triggerCharacter: context.triggerCharacter,
            triggerKind: asSignatureHelpTriggerKind(context.triggerKind),
            activeSignatureHelp: asSignatureHelp(context.activeSignatureHelp)
          }
        };
      }
      function asWorkerPosition(position) {
        return { line: position.line, character: position.character };
      }
      function asPosition(value) {
        if (value === void 0 || value === null) {
          return value;
        }
        return { line: value.line, character: value.character };
      }
      function asPositions(value) {
        let result = [];
        for (let elem of value) {
          result.push(asPosition(elem));
        }
        return result;
      }
      function asRange(value) {
        if (value === void 0 || value === null) {
          return value;
        }
        return { start: asPosition(value.start), end: asPosition(value.end) };
      }
      function asLocation(value) {
        if (value === void 0 || value === null) {
          return value;
        }
        return proto.Location.create(asUri(value.uri), asRange(value.range));
      }
      function asDiagnosticSeverity(value) {
        switch (value) {
          case code.DiagnosticSeverity.Error:
            return proto.DiagnosticSeverity.Error;
          case code.DiagnosticSeverity.Warning:
            return proto.DiagnosticSeverity.Warning;
          case code.DiagnosticSeverity.Information:
            return proto.DiagnosticSeverity.Information;
          case code.DiagnosticSeverity.Hint:
            return proto.DiagnosticSeverity.Hint;
        }
      }
      function asDiagnosticTags(tags) {
        if (!tags) {
          return void 0;
        }
        let result = [];
        for (let tag of tags) {
          let converted = asDiagnosticTag(tag);
          if (converted !== void 0) {
            result.push(converted);
          }
        }
        return result.length > 0 ? result : void 0;
      }
      function asDiagnosticTag(tag) {
        switch (tag) {
          case code.DiagnosticTag.Unnecessary:
            return proto.DiagnosticTag.Unnecessary;
          case code.DiagnosticTag.Deprecated:
            return proto.DiagnosticTag.Deprecated;
          default:
            return void 0;
        }
      }
      function asRelatedInformation(item) {
        return {
          message: item.message,
          location: asLocation(item.location)
        };
      }
      function asRelatedInformations(items) {
        return items.map(asRelatedInformation);
      }
      function asDiagnosticCode(value) {
        if (value === void 0 || value === null) {
          return void 0;
        }
        if (Is.number(value) || Is.string(value)) {
          return value;
        }
        return { value: value.value, target: asUri(value.target) };
      }
      function asDiagnostic(item) {
        const result = proto.Diagnostic.create(asRange(item.range), item.message);
        const protocolDiagnostic = item instanceof protocolDiagnostic_1.ProtocolDiagnostic ? item : void 0;
        if (protocolDiagnostic !== void 0 && protocolDiagnostic.data !== void 0) {
          result.data = protocolDiagnostic.data;
        }
        const code2 = asDiagnosticCode(item.code);
        if (protocolDiagnostic_1.DiagnosticCode.is(code2)) {
          if (protocolDiagnostic !== void 0 && protocolDiagnostic.hasDiagnosticCode) {
            result.code = code2;
          } else {
            result.code = code2.value;
            result.codeDescription = { href: code2.target };
          }
        } else {
          result.code = code2;
        }
        if (Is.number(item.severity)) {
          result.severity = asDiagnosticSeverity(item.severity);
        }
        if (Array.isArray(item.tags)) {
          result.tags = asDiagnosticTags(item.tags);
        }
        if (item.relatedInformation) {
          result.relatedInformation = asRelatedInformations(item.relatedInformation);
        }
        if (item.source) {
          result.source = item.source;
        }
        return result;
      }
      function asDiagnostics(items) {
        if (items === void 0 || items === null) {
          return items;
        }
        return items.map(asDiagnostic);
      }
      function asDocumentation(format, documentation) {
        switch (format) {
          case "$string":
            return documentation;
          case proto.MarkupKind.PlainText:
            return { kind: format, value: documentation };
          case proto.MarkupKind.Markdown:
            return { kind: format, value: documentation.value };
          default:
            return `Unsupported Markup content received. Kind is: ${format}`;
        }
      }
      function asCompletionItemTag(tag) {
        switch (tag) {
          case code.CompletionItemTag.Deprecated:
            return proto.CompletionItemTag.Deprecated;
        }
        return void 0;
      }
      function asCompletionItemTags(tags) {
        if (tags === void 0) {
          return tags;
        }
        const result = [];
        for (let tag of tags) {
          const converted = asCompletionItemTag(tag);
          if (converted !== void 0) {
            result.push(converted);
          }
        }
        return result;
      }
      function asCompletionItemKind(value, original) {
        if (original !== void 0) {
          return original;
        }
        return value + 1;
      }
      function asCompletionItem(item) {
        let result = { label: item.label };
        let protocolItem = item instanceof protocolCompletionItem_1.default ? item : void 0;
        if (item.detail) {
          result.detail = item.detail;
        }
        if (item.documentation) {
          if (!protocolItem || protocolItem.documentationFormat === "$string") {
            result.documentation = item.documentation;
          } else {
            result.documentation = asDocumentation(protocolItem.documentationFormat, item.documentation);
          }
        }
        if (item.filterText) {
          result.filterText = item.filterText;
        }
        fillPrimaryInsertText(result, item);
        if (Is.number(item.kind)) {
          result.kind = asCompletionItemKind(item.kind, protocolItem && protocolItem.originalItemKind);
        }
        if (item.sortText) {
          result.sortText = item.sortText;
        }
        if (item.additionalTextEdits) {
          result.additionalTextEdits = asTextEdits(item.additionalTextEdits);
        }
        if (item.commitCharacters) {
          result.commitCharacters = item.commitCharacters.slice();
        }
        if (item.command) {
          result.command = asCommand(item.command);
        }
        if (item.preselect === true || item.preselect === false) {
          result.preselect = item.preselect;
        }
        const tags = asCompletionItemTags(item.tags);
        if (protocolItem) {
          if (protocolItem.data !== void 0) {
            result.data = protocolItem.data;
          }
          if (protocolItem.deprecated === true || protocolItem.deprecated === false) {
            if (protocolItem.deprecated === true && tags !== void 0 && tags.length > 0) {
              const index = tags.indexOf(code.CompletionItemTag.Deprecated);
              if (index !== -1) {
                tags.splice(index, 1);
              }
            }
            result.deprecated = protocolItem.deprecated;
          }
          if (protocolItem.insertTextMode !== void 0) {
            result.insertTextMode = protocolItem.insertTextMode;
          }
        }
        if (tags !== void 0 && tags.length > 0) {
          result.tags = tags;
        }
        if (result.insertTextMode === void 0 && item.keepWhitespace === true) {
          result.insertTextMode = vscode_languageserver_protocol_1.InsertTextMode.adjustIndentation;
        }
        return result;
      }
      function fillPrimaryInsertText(target, source) {
        let format = proto.InsertTextFormat.PlainText;
        let text = void 0;
        let range = void 0;
        if (source.textEdit) {
          text = source.textEdit.newText;
          range = source.textEdit.range;
        } else if (source.insertText instanceof code.SnippetString) {
          format = proto.InsertTextFormat.Snippet;
          text = source.insertText.value;
        } else {
          text = source.insertText;
        }
        if (source.range) {
          range = source.range;
        }
        target.insertTextFormat = format;
        if (source.fromEdit && text !== void 0 && range !== void 0) {
          target.textEdit = asCompletionTextEdit(text, range);
        } else {
          target.insertText = text;
        }
      }
      function asCompletionTextEdit(newText, range) {
        if (InsertReplaceRange.is(range)) {
          return proto.InsertReplaceEdit.create(newText, asRange(range.inserting), asRange(range.replacing));
        } else {
          return { newText, range: asRange(range) };
        }
      }
      function asTextEdit(edit) {
        return { range: asRange(edit.range), newText: edit.newText };
      }
      function asTextEdits(edits) {
        if (edits === void 0 || edits === null) {
          return edits;
        }
        return edits.map(asTextEdit);
      }
      function asSymbolKind(item) {
        if (item <= code.SymbolKind.TypeParameter) {
          return item + 1;
        }
        return proto.SymbolKind.Property;
      }
      function asSymbolTag(item) {
        return item;
      }
      function asSymbolTags(items) {
        return items.map(asSymbolTag);
      }
      function asReferenceParams(textDocument, position, options) {
        return {
          textDocument: asTextDocumentIdentifier(textDocument),
          position: asWorkerPosition(position),
          context: { includeDeclaration: options.includeDeclaration }
        };
      }
      function asCodeAction(item) {
        let result = proto.CodeAction.create(item.title);
        if (item instanceof protocolCodeAction_1.default && item.data !== void 0) {
          result.data = item.data;
        }
        if (item.kind !== void 0) {
          result.kind = asCodeActionKind(item.kind);
        }
        if (item.diagnostics !== void 0) {
          result.diagnostics = asDiagnostics(item.diagnostics);
        }
        if (item.edit !== void 0) {
          throw new Error(`VS Code code actions can only be converted to a protocol code action without an edit.`);
        }
        if (item.command !== void 0) {
          result.command = asCommand(item.command);
        }
        if (item.isPreferred !== void 0) {
          result.isPreferred = item.isPreferred;
        }
        if (item.disabled !== void 0) {
          result.disabled = { reason: item.disabled.reason };
        }
        return result;
      }
      function asCodeActionContext(context) {
        if (context === void 0 || context === null) {
          return context;
        }
        let only;
        if (context.only && Is.string(context.only.value)) {
          only = [context.only.value];
        }
        return proto.CodeActionContext.create(asDiagnostics(context.diagnostics), only);
      }
      function asCodeActionKind(item) {
        if (item === void 0 || item === null) {
          return void 0;
        }
        return item.value;
      }
      function asCommand(item) {
        let result = proto.Command.create(item.title, item.command);
        if (item.arguments) {
          result.arguments = item.arguments;
        }
        return result;
      }
      function asCodeLens(item) {
        let result = proto.CodeLens.create(asRange(item.range));
        if (item.command) {
          result.command = asCommand(item.command);
        }
        if (item instanceof protocolCodeLens_1.default) {
          if (item.data) {
            result.data = item.data;
          }
        }
        return result;
      }
      function asFormattingOptions(options, fileOptions) {
        const result = { tabSize: options.tabSize, insertSpaces: options.insertSpaces };
        if (fileOptions.trimTrailingWhitespace) {
          result.trimTrailingWhitespace = true;
        }
        if (fileOptions.trimFinalNewlines) {
          result.trimFinalNewlines = true;
        }
        if (fileOptions.insertFinalNewline) {
          result.insertFinalNewline = true;
        }
        return result;
      }
      function asDocumentSymbolParams(textDocument) {
        return {
          textDocument: asTextDocumentIdentifier(textDocument)
        };
      }
      function asCodeLensParams(textDocument) {
        return {
          textDocument: asTextDocumentIdentifier(textDocument)
        };
      }
      function asDocumentLink(item) {
        let result = proto.DocumentLink.create(asRange(item.range));
        if (item.target) {
          result.target = asUri(item.target);
        }
        if (item.tooltip !== void 0) {
          result.tooltip = item.tooltip;
        }
        let protocolItem = item instanceof protocolDocumentLink_1.default ? item : void 0;
        if (protocolItem && protocolItem.data) {
          result.data = protocolItem.data;
        }
        return result;
      }
      function asDocumentLinkParams(textDocument) {
        return {
          textDocument: asTextDocumentIdentifier(textDocument)
        };
      }
      function asCallHierarchyItem(value) {
        const result = {
          name: value.name,
          kind: asSymbolKind(value.kind),
          uri: asUri(value.uri),
          range: asRange(value.range),
          selectionRange: asRange(value.selectionRange)
        };
        if (value.detail !== void 0 && value.detail.length > 0) {
          result.detail = value.detail;
        }
        if (value.tags !== void 0) {
          result.tags = asSymbolTags(value.tags);
        }
        if (value instanceof protocolCallHierarchyItem_1.default && value.data !== void 0) {
          result.data = value.data;
        }
        return result;
      }
      return {
        asUri,
        asTextDocumentIdentifier,
        asVersionedTextDocumentIdentifier,
        asOpenTextDocumentParams,
        asChangeTextDocumentParams,
        asCloseTextDocumentParams,
        asSaveTextDocumentParams,
        asWillSaveTextDocumentParams,
        asDidCreateFilesParams,
        asDidRenameFilesParams,
        asDidDeleteFilesParams,
        asWillCreateFilesParams,
        asWillRenameFilesParams,
        asWillDeleteFilesParams,
        asTextDocumentPositionParams,
        asCompletionParams,
        asSignatureHelpParams,
        asWorkerPosition,
        asRange,
        asPosition,
        asPositions,
        asLocation,
        asDiagnosticSeverity,
        asDiagnosticTag,
        asDiagnostic,
        asDiagnostics,
        asCompletionItem,
        asTextEdit,
        asSymbolKind,
        asSymbolTag,
        asSymbolTags,
        asReferenceParams,
        asCodeAction,
        asCodeActionContext,
        asCommand,
        asCodeLens,
        asFormattingOptions,
        asDocumentSymbolParams,
        asCodeLensParams,
        asDocumentLink,
        asDocumentLinkParams,
        asCallHierarchyItem
      };
    }
    exports2.createConverter = createConverter;
  }
});

// node_modules/vscode-languageclient/lib/common/protocolConverter.js
var require_protocolConverter = __commonJS({
  "node_modules/vscode-languageclient/lib/common/protocolConverter.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.createConverter = void 0;
    var code = require("vscode");
    var ls = require_main3();
    var Is = require_is();
    var protocolCompletionItem_1 = require_protocolCompletionItem();
    var protocolCodeLens_1 = require_protocolCodeLens();
    var protocolDocumentLink_1 = require_protocolDocumentLink();
    var protocolCodeAction_1 = require_protocolCodeAction();
    var protocolDiagnostic_1 = require_protocolDiagnostic();
    var protocolCallHierarchyItem_1 = require_protocolCallHierarchyItem();
    var vscode_languageserver_protocol_1 = require_main3();
    var CodeBlock;
    (function(CodeBlock2) {
      function is(value) {
        let candidate = value;
        return candidate && Is.string(candidate.language) && Is.string(candidate.value);
      }
      CodeBlock2.is = is;
    })(CodeBlock || (CodeBlock = {}));
    function createConverter(uriConverter, trustMarkdown) {
      const nullConverter = (value) => code.Uri.parse(value);
      const _uriConverter = uriConverter || nullConverter;
      function asUri(value) {
        return _uriConverter(value);
      }
      function asDiagnostics(diagnostics) {
        return diagnostics.map(asDiagnostic);
      }
      function asDiagnostic(diagnostic) {
        let result = new protocolDiagnostic_1.ProtocolDiagnostic(asRange(diagnostic.range), diagnostic.message, asDiagnosticSeverity(diagnostic.severity), diagnostic.data);
        if (diagnostic.code !== void 0) {
          if (ls.CodeDescription.is(diagnostic.codeDescription)) {
            result.code = {
              value: diagnostic.code,
              target: asUri(diagnostic.codeDescription.href)
            };
          } else if (protocolDiagnostic_1.DiagnosticCode.is(diagnostic.code)) {
            result.hasDiagnosticCode = true;
            result.code = {
              value: diagnostic.code.value,
              target: asUri(diagnostic.code.target)
            };
          } else {
            result.code = diagnostic.code;
          }
        }
        if (diagnostic.source) {
          result.source = diagnostic.source;
        }
        if (diagnostic.relatedInformation) {
          result.relatedInformation = asRelatedInformation(diagnostic.relatedInformation);
        }
        if (Array.isArray(diagnostic.tags)) {
          result.tags = asDiagnosticTags(diagnostic.tags);
        }
        return result;
      }
      function asRelatedInformation(relatedInformation) {
        return relatedInformation.map(asDiagnosticRelatedInformation);
      }
      function asDiagnosticRelatedInformation(information) {
        return new code.DiagnosticRelatedInformation(asLocation(information.location), information.message);
      }
      function asDiagnosticTags(tags) {
        if (!tags) {
          return void 0;
        }
        let result = [];
        for (let tag of tags) {
          let converted = asDiagnosticTag(tag);
          if (converted !== void 0) {
            result.push(converted);
          }
        }
        return result.length > 0 ? result : void 0;
      }
      function asDiagnosticTag(tag) {
        switch (tag) {
          case ls.DiagnosticTag.Unnecessary:
            return code.DiagnosticTag.Unnecessary;
          case ls.DiagnosticTag.Deprecated:
            return code.DiagnosticTag.Deprecated;
          default:
            return void 0;
        }
      }
      function asPosition(value) {
        if (!value) {
          return void 0;
        }
        return new code.Position(value.line, value.character);
      }
      function asRange(value) {
        if (!value) {
          return void 0;
        }
        return new code.Range(asPosition(value.start), asPosition(value.end));
      }
      function asRanges(value) {
        return value.map((value2) => asRange(value2));
      }
      function asDiagnosticSeverity(value) {
        if (value === void 0 || value === null) {
          return code.DiagnosticSeverity.Error;
        }
        switch (value) {
          case ls.DiagnosticSeverity.Error:
            return code.DiagnosticSeverity.Error;
          case ls.DiagnosticSeverity.Warning:
            return code.DiagnosticSeverity.Warning;
          case ls.DiagnosticSeverity.Information:
            return code.DiagnosticSeverity.Information;
          case ls.DiagnosticSeverity.Hint:
            return code.DiagnosticSeverity.Hint;
        }
        return code.DiagnosticSeverity.Error;
      }
      function asHoverContent(value) {
        if (Is.string(value)) {
          return asMarkdownString(value);
        } else if (CodeBlock.is(value)) {
          let result = asMarkdownString();
          return result.appendCodeblock(value.value, value.language);
        } else if (Array.isArray(value)) {
          let result = [];
          for (let element of value) {
            let item = asMarkdownString();
            if (CodeBlock.is(element)) {
              item.appendCodeblock(element.value, element.language);
            } else {
              item.appendMarkdown(element);
            }
            result.push(item);
          }
          return result;
        } else {
          let result;
          switch (value.kind) {
            case ls.MarkupKind.Markdown:
              return asMarkdownString(value.value);
            case ls.MarkupKind.PlainText:
              result = asMarkdownString();
              result.appendText(value.value);
              return result;
            default:
              result = asMarkdownString();
              result.appendText(`Unsupported Markup content received. Kind is: ${value.kind}`);
              return result;
          }
        }
      }
      function asDocumentation(value) {
        if (Is.string(value)) {
          return value;
        } else {
          switch (value.kind) {
            case ls.MarkupKind.Markdown:
              return asMarkdownString(value.value);
            case ls.MarkupKind.PlainText:
              return value.value;
            default:
              return `Unsupported Markup content received. Kind is: ${value.kind}`;
          }
        }
      }
      function asMarkdownString(value) {
        const result = new code.MarkdownString(value);
        if (trustMarkdown === true) {
          result.isTrusted = trustMarkdown;
        }
        return result;
      }
      function asHover(hover) {
        if (!hover) {
          return void 0;
        }
        return new code.Hover(asHoverContent(hover.contents), asRange(hover.range));
      }
      function asCompletionResult(result) {
        if (!result) {
          return void 0;
        }
        if (Array.isArray(result)) {
          let items = result;
          return items.map(asCompletionItem);
        }
        let list = result;
        return new code.CompletionList(list.items.map(asCompletionItem), list.isIncomplete);
      }
      function asCompletionItemKind(value) {
        if (ls.CompletionItemKind.Text <= value && value <= ls.CompletionItemKind.TypeParameter) {
          return [value - 1, void 0];
        }
        return [code.CompletionItemKind.Text, value];
      }
      function asCompletionItemTag(tag) {
        switch (tag) {
          case ls.CompletionItemTag.Deprecated:
            return code.CompletionItemTag.Deprecated;
        }
        return void 0;
      }
      function asCompletionItemTags(tags) {
        if (tags === void 0 || tags === null) {
          return [];
        }
        const result = [];
        for (let tag of tags) {
          const converted = asCompletionItemTag(tag);
          if (converted !== void 0) {
            result.push(converted);
          }
        }
        return result;
      }
      function asCompletionItem(item) {
        let tags = asCompletionItemTags(item.tags);
        let result = new protocolCompletionItem_1.default(item.label);
        if (item.detail) {
          result.detail = item.detail;
        }
        if (item.documentation) {
          result.documentation = asDocumentation(item.documentation);
          result.documentationFormat = Is.string(item.documentation) ? "$string" : item.documentation.kind;
        }
        if (item.filterText) {
          result.filterText = item.filterText;
        }
        let insertText = asCompletionInsertText(item);
        if (insertText) {
          result.insertText = insertText.text;
          result.range = insertText.range;
          result.fromEdit = insertText.fromEdit;
        }
        if (Is.number(item.kind)) {
          let [itemKind, original] = asCompletionItemKind(item.kind);
          result.kind = itemKind;
          if (original) {
            result.originalItemKind = original;
          }
        }
        if (item.sortText) {
          result.sortText = item.sortText;
        }
        if (item.additionalTextEdits) {
          result.additionalTextEdits = asTextEdits(item.additionalTextEdits);
        }
        if (Is.stringArray(item.commitCharacters)) {
          result.commitCharacters = item.commitCharacters.slice();
        }
        if (item.command) {
          result.command = asCommand(item.command);
        }
        if (item.deprecated === true || item.deprecated === false) {
          result.deprecated = item.deprecated;
          if (item.deprecated === true) {
            tags.push(code.CompletionItemTag.Deprecated);
          }
        }
        if (item.preselect === true || item.preselect === false) {
          result.preselect = item.preselect;
        }
        if (item.data !== void 0) {
          result.data = item.data;
        }
        if (tags.length > 0) {
          result.tags = tags;
        }
        if (item.insertTextMode !== void 0) {
          result.insertTextMode = item.insertTextMode;
          if (item.insertTextMode === vscode_languageserver_protocol_1.InsertTextMode.asIs) {
            result.keepWhitespace = true;
          }
        }
        return result;
      }
      function asCompletionInsertText(item) {
        if (item.textEdit) {
          if (item.insertTextFormat === ls.InsertTextFormat.Snippet) {
            return { text: new code.SnippetString(item.textEdit.newText), range: asCompletionRange(item.textEdit), fromEdit: true };
          } else {
            return { text: item.textEdit.newText, range: asCompletionRange(item.textEdit), fromEdit: true };
          }
        } else if (item.insertText) {
          if (item.insertTextFormat === ls.InsertTextFormat.Snippet) {
            return { text: new code.SnippetString(item.insertText), fromEdit: false };
          } else {
            return { text: item.insertText, fromEdit: false };
          }
        } else {
          return void 0;
        }
      }
      function asCompletionRange(value) {
        if (ls.InsertReplaceEdit.is(value)) {
          return { inserting: asRange(value.insert), replacing: asRange(value.replace) };
        } else {
          return asRange(value.range);
        }
      }
      function asTextEdit(edit) {
        if (!edit) {
          return void 0;
        }
        return new code.TextEdit(asRange(edit.range), edit.newText);
      }
      function asTextEdits(items) {
        if (!items) {
          return void 0;
        }
        return items.map(asTextEdit);
      }
      function asSignatureHelp(item) {
        if (!item) {
          return void 0;
        }
        let result = new code.SignatureHelp();
        if (Is.number(item.activeSignature)) {
          result.activeSignature = item.activeSignature;
        } else {
          result.activeSignature = 0;
        }
        if (Is.number(item.activeParameter)) {
          result.activeParameter = item.activeParameter;
        } else {
          result.activeParameter = 0;
        }
        if (item.signatures) {
          result.signatures = asSignatureInformations(item.signatures);
        }
        return result;
      }
      function asSignatureInformations(items) {
        return items.map(asSignatureInformation);
      }
      function asSignatureInformation(item) {
        let result = new code.SignatureInformation(item.label);
        if (item.documentation !== void 0) {
          result.documentation = asDocumentation(item.documentation);
        }
        if (item.parameters !== void 0) {
          result.parameters = asParameterInformations(item.parameters);
        }
        if (item.activeParameter !== void 0) {
          result.activeParameter = item.activeParameter;
        }
        {
          return result;
        }
      }
      function asParameterInformations(item) {
        return item.map(asParameterInformation);
      }
      function asParameterInformation(item) {
        let result = new code.ParameterInformation(item.label);
        if (item.documentation) {
          result.documentation = asDocumentation(item.documentation);
        }
        return result;
      }
      function asLocation(item) {
        if (!item) {
          return void 0;
        }
        return new code.Location(_uriConverter(item.uri), asRange(item.range));
      }
      function asDeclarationResult(item) {
        if (!item) {
          return void 0;
        }
        return asLocationResult(item);
      }
      function asDefinitionResult(item) {
        if (!item) {
          return void 0;
        }
        return asLocationResult(item);
      }
      function asLocationLink(item) {
        if (!item) {
          return void 0;
        }
        let result = {
          targetUri: _uriConverter(item.targetUri),
          targetRange: asRange(item.targetRange),
          originSelectionRange: asRange(item.originSelectionRange),
          targetSelectionRange: asRange(item.targetSelectionRange)
        };
        if (!result.targetSelectionRange) {
          throw new Error(`targetSelectionRange must not be undefined or null`);
        }
        return result;
      }
      function asLocationResult(item) {
        if (!item) {
          return void 0;
        }
        if (Is.array(item)) {
          if (item.length === 0) {
            return [];
          } else if (ls.LocationLink.is(item[0])) {
            let links = item;
            return links.map((link) => asLocationLink(link));
          } else {
            let locations = item;
            return locations.map((location) => asLocation(location));
          }
        } else if (ls.LocationLink.is(item)) {
          return [asLocationLink(item)];
        } else {
          return asLocation(item);
        }
      }
      function asReferences(values) {
        if (!values) {
          return void 0;
        }
        return values.map((location) => asLocation(location));
      }
      function asDocumentHighlights(values) {
        if (!values) {
          return void 0;
        }
        return values.map(asDocumentHighlight);
      }
      function asDocumentHighlight(item) {
        let result = new code.DocumentHighlight(asRange(item.range));
        if (Is.number(item.kind)) {
          result.kind = asDocumentHighlightKind(item.kind);
        }
        return result;
      }
      function asDocumentHighlightKind(item) {
        switch (item) {
          case ls.DocumentHighlightKind.Text:
            return code.DocumentHighlightKind.Text;
          case ls.DocumentHighlightKind.Read:
            return code.DocumentHighlightKind.Read;
          case ls.DocumentHighlightKind.Write:
            return code.DocumentHighlightKind.Write;
        }
        return code.DocumentHighlightKind.Text;
      }
      function asSymbolInformations(values, uri) {
        if (!values) {
          return void 0;
        }
        return values.map((information) => asSymbolInformation(information, uri));
      }
      function asSymbolKind(item) {
        if (item <= ls.SymbolKind.TypeParameter) {
          return item - 1;
        }
        return code.SymbolKind.Property;
      }
      function asSymbolTag(value) {
        switch (value) {
          case ls.SymbolTag.Deprecated:
            return code.SymbolTag.Deprecated;
          default:
            return void 0;
        }
      }
      function asSymbolTags(items) {
        if (items === void 0 || items === null) {
          return void 0;
        }
        const result = [];
        for (const item of items) {
          const converted = asSymbolTag(item);
          if (converted !== void 0) {
            result.push(converted);
          }
        }
        return result.length === 0 ? void 0 : result;
      }
      function asSymbolInformation(item, uri) {
        let result = new code.SymbolInformation(item.name, asSymbolKind(item.kind), asRange(item.location.range), item.location.uri ? _uriConverter(item.location.uri) : uri);
        fillTags(result, item);
        if (item.containerName) {
          result.containerName = item.containerName;
        }
        return result;
      }
      function asDocumentSymbols(values) {
        if (values === void 0 || values === null) {
          return void 0;
        }
        return values.map(asDocumentSymbol);
      }
      function asDocumentSymbol(value) {
        let result = new code.DocumentSymbol(value.name, value.detail || "", asSymbolKind(value.kind), asRange(value.range), asRange(value.selectionRange));
        fillTags(result, value);
        if (value.children !== void 0 && value.children.length > 0) {
          let children = [];
          for (let child of value.children) {
            children.push(asDocumentSymbol(child));
          }
          result.children = children;
        }
        return result;
      }
      function fillTags(result, value) {
        result.tags = asSymbolTags(value.tags);
        if (value.deprecated) {
          if (!result.tags) {
            result.tags = [code.SymbolTag.Deprecated];
          } else {
            if (!result.tags.includes(code.SymbolTag.Deprecated)) {
              result.tags = result.tags.concat(code.SymbolTag.Deprecated);
            }
          }
        }
      }
      function asCommand(item) {
        let result = { title: item.title, command: item.command };
        if (item.arguments) {
          result.arguments = item.arguments;
        }
        return result;
      }
      function asCommands(items) {
        if (!items) {
          return void 0;
        }
        return items.map(asCommand);
      }
      const kindMapping = /* @__PURE__ */ new Map();
      kindMapping.set(ls.CodeActionKind.Empty, code.CodeActionKind.Empty);
      kindMapping.set(ls.CodeActionKind.QuickFix, code.CodeActionKind.QuickFix);
      kindMapping.set(ls.CodeActionKind.Refactor, code.CodeActionKind.Refactor);
      kindMapping.set(ls.CodeActionKind.RefactorExtract, code.CodeActionKind.RefactorExtract);
      kindMapping.set(ls.CodeActionKind.RefactorInline, code.CodeActionKind.RefactorInline);
      kindMapping.set(ls.CodeActionKind.RefactorRewrite, code.CodeActionKind.RefactorRewrite);
      kindMapping.set(ls.CodeActionKind.Source, code.CodeActionKind.Source);
      kindMapping.set(ls.CodeActionKind.SourceOrganizeImports, code.CodeActionKind.SourceOrganizeImports);
      function asCodeActionKind(item) {
        if (item === void 0 || item === null) {
          return void 0;
        }
        let result = kindMapping.get(item);
        if (result) {
          return result;
        }
        let parts = item.split(".");
        result = code.CodeActionKind.Empty;
        for (let part of parts) {
          result = result.append(part);
        }
        return result;
      }
      function asCodeActionKinds(items) {
        if (items === void 0 || items === null) {
          return void 0;
        }
        return items.map((kind) => asCodeActionKind(kind));
      }
      function asCodeAction(item) {
        if (item === void 0 || item === null) {
          return void 0;
        }
        let result = new protocolCodeAction_1.default(item.title, item.data);
        if (item.kind !== void 0) {
          result.kind = asCodeActionKind(item.kind);
        }
        if (item.diagnostics !== void 0) {
          result.diagnostics = asDiagnostics(item.diagnostics);
        }
        if (item.edit !== void 0) {
          result.edit = asWorkspaceEdit(item.edit);
        }
        if (item.command !== void 0) {
          result.command = asCommand(item.command);
        }
        if (item.isPreferred !== void 0) {
          result.isPreferred = item.isPreferred;
        }
        if (item.disabled !== void 0) {
          result.disabled = { reason: item.disabled.reason };
        }
        return result;
      }
      function asCodeLens(item) {
        if (!item) {
          return void 0;
        }
        let result = new protocolCodeLens_1.default(asRange(item.range));
        if (item.command) {
          result.command = asCommand(item.command);
        }
        if (item.data !== void 0 && item.data !== null) {
          result.data = item.data;
        }
        return result;
      }
      function asCodeLenses(items) {
        if (!items) {
          return void 0;
        }
        return items.map((codeLens) => asCodeLens(codeLens));
      }
      function asWorkspaceEdit(item) {
        if (!item) {
          return void 0;
        }
        const sharedMetadata = /* @__PURE__ */ new Map();
        if (item.changeAnnotations !== void 0) {
          for (const key of Object.keys(item.changeAnnotations)) {
            const metaData = asWorkspaceEditEntryMetadata(item.changeAnnotations[key]);
            sharedMetadata.set(key, metaData);
          }
        }
        const asMetadata = (annotation) => {
          if (annotation === void 0) {
            return void 0;
          } else {
            return sharedMetadata.get(annotation);
          }
        };
        const result = new code.WorkspaceEdit();
        if (item.documentChanges) {
          for (const change of item.documentChanges) {
            if (ls.CreateFile.is(change)) {
              result.createFile(_uriConverter(change.uri), change.options, asMetadata(change.annotationId));
            } else if (ls.RenameFile.is(change)) {
              result.renameFile(_uriConverter(change.oldUri), _uriConverter(change.newUri), change.options, asMetadata(change.annotationId));
            } else if (ls.DeleteFile.is(change)) {
              result.deleteFile(_uriConverter(change.uri), change.options, asMetadata(change.annotationId));
            } else if (ls.TextDocumentEdit.is(change)) {
              const uri = _uriConverter(change.textDocument.uri);
              for (const edit of change.edits) {
                if (vscode_languageserver_protocol_1.AnnotatedTextEdit.is(edit)) {
                  result.replace(uri, asRange(edit.range), edit.newText, asMetadata(edit.annotationId));
                } else {
                  result.replace(uri, asRange(edit.range), edit.newText);
                }
              }
            } else {
              throw new Error(`Unknown workspace edit change received:
${JSON.stringify(change, void 0, 4)}`);
            }
          }
        } else if (item.changes) {
          Object.keys(item.changes).forEach((key) => {
            result.set(_uriConverter(key), asTextEdits(item.changes[key]));
          });
        }
        return result;
      }
      function asWorkspaceEditEntryMetadata(annotation) {
        if (annotation === void 0) {
          return void 0;
        }
        return { label: annotation.label, needsConfirmation: !!annotation.needsConfirmation, description: annotation.description };
      }
      function asDocumentLink(item) {
        let range = asRange(item.range);
        let target = item.target ? asUri(item.target) : void 0;
        let link = new protocolDocumentLink_1.default(range, target);
        if (item.tooltip !== void 0) {
          link.tooltip = item.tooltip;
        }
        if (item.data !== void 0 && item.data !== null) {
          link.data = item.data;
        }
        return link;
      }
      function asDocumentLinks(items) {
        if (!items) {
          return void 0;
        }
        return items.map(asDocumentLink);
      }
      function asColor(color) {
        return new code.Color(color.red, color.green, color.blue, color.alpha);
      }
      function asColorInformation(ci) {
        return new code.ColorInformation(asRange(ci.range), asColor(ci.color));
      }
      function asColorInformations(colorInformation) {
        if (Array.isArray(colorInformation)) {
          return colorInformation.map(asColorInformation);
        }
        return void 0;
      }
      function asColorPresentation(cp) {
        let presentation = new code.ColorPresentation(cp.label);
        presentation.additionalTextEdits = asTextEdits(cp.additionalTextEdits);
        if (cp.textEdit) {
          presentation.textEdit = asTextEdit(cp.textEdit);
        }
        return presentation;
      }
      function asColorPresentations(colorPresentations) {
        if (Array.isArray(colorPresentations)) {
          return colorPresentations.map(asColorPresentation);
        }
        return void 0;
      }
      function asFoldingRangeKind(kind) {
        if (kind) {
          switch (kind) {
            case ls.FoldingRangeKind.Comment:
              return code.FoldingRangeKind.Comment;
            case ls.FoldingRangeKind.Imports:
              return code.FoldingRangeKind.Imports;
            case ls.FoldingRangeKind.Region:
              return code.FoldingRangeKind.Region;
          }
        }
        return void 0;
      }
      function asFoldingRange(r) {
        return new code.FoldingRange(r.startLine, r.endLine, asFoldingRangeKind(r.kind));
      }
      function asFoldingRanges(foldingRanges) {
        if (Array.isArray(foldingRanges)) {
          return foldingRanges.map(asFoldingRange);
        }
        return void 0;
      }
      function asSelectionRange(selectionRange) {
        return new code.SelectionRange(asRange(selectionRange.range), selectionRange.parent ? asSelectionRange(selectionRange.parent) : void 0);
      }
      function asSelectionRanges(selectionRanges) {
        if (!Array.isArray(selectionRanges)) {
          return [];
        }
        let result = [];
        for (let range of selectionRanges) {
          result.push(asSelectionRange(range));
        }
        return result;
      }
      function asCallHierarchyItem(item) {
        if (item === null) {
          return void 0;
        }
        let result = new protocolCallHierarchyItem_1.default(asSymbolKind(item.kind), item.name, item.detail || "", asUri(item.uri), asRange(item.range), asRange(item.selectionRange), item.data);
        if (item.tags !== void 0) {
          result.tags = asSymbolTags(item.tags);
        }
        return result;
      }
      function asCallHierarchyItems(items) {
        if (items === null) {
          return void 0;
        }
        return items.map((item) => asCallHierarchyItem(item));
      }
      function asCallHierarchyIncomingCall(item) {
        return new code.CallHierarchyIncomingCall(asCallHierarchyItem(item.from), asRanges(item.fromRanges));
      }
      function asCallHierarchyIncomingCalls(items) {
        if (items === null) {
          return void 0;
        }
        return items.map((item) => asCallHierarchyIncomingCall(item));
      }
      function asCallHierarchyOutgoingCall(item) {
        return new code.CallHierarchyOutgoingCall(asCallHierarchyItem(item.to), asRanges(item.fromRanges));
      }
      function asCallHierarchyOutgoingCalls(items) {
        if (items === null) {
          return void 0;
        }
        return items.map((item) => asCallHierarchyOutgoingCall(item));
      }
      function asSemanticTokens(value) {
        if (value === void 0 || value === null) {
          return void 0;
        }
        return new code.SemanticTokens(new Uint32Array(value.data), value.resultId);
      }
      function asSemanticTokensEdit(value) {
        return new code.SemanticTokensEdit(value.start, value.deleteCount, value.data !== void 0 ? new Uint32Array(value.data) : void 0);
      }
      function asSemanticTokensEdits(value) {
        if (value === void 0 || value === null) {
          return void 0;
        }
        return new code.SemanticTokensEdits(value.edits.map(asSemanticTokensEdit), value.resultId);
      }
      function asSemanticTokensLegend(value) {
        return value;
      }
      function asLinkedEditingRanges(value) {
        if (value === null || value === void 0) {
          return void 0;
        }
        return new code.LinkedEditingRanges(asRanges(value.ranges), asRegularExpression(value.wordPattern));
      }
      function asRegularExpression(value) {
        if (value === null || value === void 0) {
          return void 0;
        }
        return new RegExp(value);
      }
      return {
        asUri,
        asDiagnostics,
        asDiagnostic,
        asRange,
        asRanges,
        asPosition,
        asDiagnosticSeverity,
        asDiagnosticTag,
        asHover,
        asCompletionResult,
        asCompletionItem,
        asTextEdit,
        asTextEdits,
        asSignatureHelp,
        asSignatureInformations,
        asSignatureInformation,
        asParameterInformations,
        asParameterInformation,
        asDeclarationResult,
        asDefinitionResult,
        asLocation,
        asReferences,
        asDocumentHighlights,
        asDocumentHighlight,
        asDocumentHighlightKind,
        asSymbolKind,
        asSymbolTag,
        asSymbolTags,
        asSymbolInformations,
        asSymbolInformation,
        asDocumentSymbols,
        asDocumentSymbol,
        asCommand,
        asCommands,
        asCodeAction,
        asCodeActionKind,
        asCodeActionKinds,
        asCodeLens,
        asCodeLenses,
        asWorkspaceEdit,
        asDocumentLink,
        asDocumentLinks,
        asFoldingRangeKind,
        asFoldingRange,
        asFoldingRanges,
        asColor,
        asColorInformation,
        asColorInformations,
        asColorPresentation,
        asColorPresentations,
        asSelectionRange,
        asSelectionRanges,
        asSemanticTokensLegend,
        asSemanticTokens,
        asSemanticTokensEdit,
        asSemanticTokensEdits,
        asCallHierarchyItem,
        asCallHierarchyItems,
        asCallHierarchyIncomingCall,
        asCallHierarchyIncomingCalls,
        asCallHierarchyOutgoingCall,
        asCallHierarchyOutgoingCalls,
        asLinkedEditingRanges
      };
    }
    exports2.createConverter = createConverter;
  }
});

// node_modules/vscode-languageclient/lib/common/utils/async.js
var require_async = __commonJS({
  "node_modules/vscode-languageclient/lib/common/utils/async.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.Delayer = void 0;
    var vscode_languageserver_protocol_1 = require_main3();
    var Delayer = class {
      constructor(defaultDelay) {
        this.defaultDelay = defaultDelay;
        this.timeout = void 0;
        this.completionPromise = void 0;
        this.onSuccess = void 0;
        this.task = void 0;
      }
      trigger(task, delay = this.defaultDelay) {
        this.task = task;
        if (delay >= 0) {
          this.cancelTimeout();
        }
        if (!this.completionPromise) {
          this.completionPromise = new Promise((resolve) => {
            this.onSuccess = resolve;
          }).then(() => {
            this.completionPromise = void 0;
            this.onSuccess = void 0;
            var result = this.task();
            this.task = void 0;
            return result;
          });
        }
        if (delay >= 0 || this.timeout === void 0) {
          this.timeout = vscode_languageserver_protocol_1.RAL().timer.setTimeout(() => {
            this.timeout = void 0;
            this.onSuccess(void 0);
          }, delay >= 0 ? delay : this.defaultDelay);
        }
        return this.completionPromise;
      }
      forceDelivery() {
        if (!this.completionPromise) {
          return void 0;
        }
        this.cancelTimeout();
        let result = this.task();
        this.completionPromise = void 0;
        this.onSuccess = void 0;
        this.task = void 0;
        return result;
      }
      isTriggered() {
        return this.timeout !== void 0;
      }
      cancel() {
        this.cancelTimeout();
        this.completionPromise = void 0;
      }
      cancelTimeout() {
        if (this.timeout !== void 0) {
          vscode_languageserver_protocol_1.RAL().timer.clearTimeout(this.timeout);
          this.timeout = void 0;
        }
      }
    };
    exports2.Delayer = Delayer;
  }
});

// node_modules/vscode-languageclient/lib/common/utils/uuid.js
var require_uuid = __commonJS({
  "node_modules/vscode-languageclient/lib/common/utils/uuid.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.generateUuid = exports2.parse = exports2.isUUID = exports2.v4 = exports2.empty = void 0;
    var ValueUUID = class {
      constructor(_value) {
        this._value = _value;
      }
      asHex() {
        return this._value;
      }
      equals(other) {
        return this.asHex() === other.asHex();
      }
    };
    var V4UUID = class _V4UUID extends ValueUUID {
      constructor() {
        super([
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          "-",
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          "-",
          "4",
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          "-",
          _V4UUID._oneOf(_V4UUID._timeHighBits),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          "-",
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex(),
          _V4UUID._randomHex()
        ].join(""));
      }
      static _oneOf(array) {
        return array[Math.floor(array.length * Math.random())];
      }
      static _randomHex() {
        return _V4UUID._oneOf(_V4UUID._chars);
      }
    };
    V4UUID._chars = ["0", "1", "2", "3", "4", "5", "6", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"];
    V4UUID._timeHighBits = ["8", "9", "a", "b"];
    exports2.empty = new ValueUUID("00000000-0000-0000-0000-000000000000");
    function v4() {
      return new V4UUID();
    }
    exports2.v4 = v4;
    var _UUIDPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    function isUUID(value) {
      return _UUIDPattern.test(value);
    }
    exports2.isUUID = isUUID;
    function parse(value) {
      if (!isUUID(value)) {
        throw new Error("invalid uuid");
      }
      return new ValueUUID(value);
    }
    exports2.parse = parse;
    function generateUuid() {
      return v4().asHex();
    }
    exports2.generateUuid = generateUuid;
  }
});

// node_modules/vscode-languageclient/lib/common/progressPart.js
var require_progressPart = __commonJS({
  "node_modules/vscode-languageclient/lib/common/progressPart.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.ProgressPart = void 0;
    var vscode_1 = require("vscode");
    var vscode_languageserver_protocol_1 = require_main3();
    var Is = require_is();
    var ProgressPart = class {
      constructor(_client, _token, done) {
        this._client = _client;
        this._token = _token;
        this._reported = 0;
        this._disposable = this._client.onProgress(vscode_languageserver_protocol_1.WorkDoneProgress.type, this._token, (value) => {
          switch (value.kind) {
            case "begin":
              this.begin(value);
              break;
            case "report":
              this.report(value);
              break;
            case "end":
              this.done();
              done && done(this);
              break;
          }
        });
      }
      begin(params) {
        vscode_1.window.withProgress({ location: vscode_1.ProgressLocation.Window, cancellable: params.cancellable, title: params.title }, async (progress, cancellationToken) => {
          this._progress = progress;
          this._infinite = params.percentage === void 0;
          this._cancellationToken = cancellationToken;
          this._cancellationToken.onCancellationRequested(() => {
            this._client.sendNotification(vscode_languageserver_protocol_1.WorkDoneProgressCancelNotification.type, { token: this._token });
          });
          this.report(params);
          return new Promise((resolve, reject) => {
            this._resolve = resolve;
            this._reject = reject;
          });
        });
      }
      report(params) {
        if (this._infinite && Is.string(params.message)) {
          this._progress.report({ message: params.message });
        } else if (Is.number(params.percentage)) {
          let percentage = Math.max(0, Math.min(params.percentage, 100));
          let delta = Math.max(0, percentage - this._reported);
          this._progress.report({ message: params.message, increment: delta });
          this._reported += delta;
        }
      }
      cancel() {
        if (this._disposable) {
          this._disposable.dispose();
          this._disposable = void 0;
        }
        if (this._reject) {
          this._reject();
          this._resolve = void 0;
          this._reject = void 0;
        }
      }
      done() {
        if (this._disposable) {
          this._disposable.dispose();
          this._disposable = void 0;
        }
        if (this._resolve) {
          this._resolve();
          this._resolve = void 0;
          this._reject = void 0;
        }
      }
    };
    exports2.ProgressPart = ProgressPart;
  }
});

// node_modules/vscode-languageclient/lib/common/client.js
var require_client = __commonJS({
  "node_modules/vscode-languageclient/lib/common/client.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.BaseLanguageClient = exports2.MessageTransports = exports2.TextDocumentFeature = exports2.State = exports2.RevealOutputChannelOn = exports2.CloseAction = exports2.ErrorAction = void 0;
    var vscode_1 = require("vscode");
    var vscode_languageserver_protocol_1 = require_main3();
    var configuration_1 = require_configuration();
    var c2p = require_codeConverter();
    var p2c = require_protocolConverter();
    var Is = require_is();
    var async_1 = require_async();
    var UUID = require_uuid();
    var progressPart_1 = require_progressPart();
    var ConsoleLogger = class {
      error(message) {
        vscode_languageserver_protocol_1.RAL().console.error(message);
      }
      warn(message) {
        vscode_languageserver_protocol_1.RAL().console.warn(message);
      }
      info(message) {
        vscode_languageserver_protocol_1.RAL().console.info(message);
      }
      log(message) {
        vscode_languageserver_protocol_1.RAL().console.log(message);
      }
    };
    function createConnection(input, output, errorHandler, closeHandler, options) {
      let logger = new ConsoleLogger();
      let connection = vscode_languageserver_protocol_1.createProtocolConnection(input, output, logger, options);
      connection.onError((data) => {
        errorHandler(data[0], data[1], data[2]);
      });
      connection.onClose(closeHandler);
      let result = {
        listen: () => connection.listen(),
        sendRequest: (type, ...params) => connection.sendRequest(Is.string(type) ? type : type.method, ...params),
        onRequest: (type, handler) => connection.onRequest(Is.string(type) ? type : type.method, handler),
        sendNotification: (type, params) => connection.sendNotification(Is.string(type) ? type : type.method, params),
        onNotification: (type, handler) => connection.onNotification(Is.string(type) ? type : type.method, handler),
        onProgress: connection.onProgress,
        sendProgress: connection.sendProgress,
        trace: (value, tracer, sendNotificationOrTraceOptions) => {
          const defaultTraceOptions = {
            sendNotification: false,
            traceFormat: vscode_languageserver_protocol_1.TraceFormat.Text
          };
          if (sendNotificationOrTraceOptions === void 0) {
            connection.trace(value, tracer, defaultTraceOptions);
          } else if (Is.boolean(sendNotificationOrTraceOptions)) {
            connection.trace(value, tracer, sendNotificationOrTraceOptions);
          } else {
            connection.trace(value, tracer, sendNotificationOrTraceOptions);
          }
        },
        initialize: (params) => connection.sendRequest(vscode_languageserver_protocol_1.InitializeRequest.type, params),
        shutdown: () => connection.sendRequest(vscode_languageserver_protocol_1.ShutdownRequest.type, void 0),
        exit: () => connection.sendNotification(vscode_languageserver_protocol_1.ExitNotification.type),
        onLogMessage: (handler) => connection.onNotification(vscode_languageserver_protocol_1.LogMessageNotification.type, handler),
        onShowMessage: (handler) => connection.onNotification(vscode_languageserver_protocol_1.ShowMessageNotification.type, handler),
        onTelemetry: (handler) => connection.onNotification(vscode_languageserver_protocol_1.TelemetryEventNotification.type, handler),
        didChangeConfiguration: (params) => connection.sendNotification(vscode_languageserver_protocol_1.DidChangeConfigurationNotification.type, params),
        didChangeWatchedFiles: (params) => connection.sendNotification(vscode_languageserver_protocol_1.DidChangeWatchedFilesNotification.type, params),
        didOpenTextDocument: (params) => connection.sendNotification(vscode_languageserver_protocol_1.DidOpenTextDocumentNotification.type, params),
        didChangeTextDocument: (params) => connection.sendNotification(vscode_languageserver_protocol_1.DidChangeTextDocumentNotification.type, params),
        didCloseTextDocument: (params) => connection.sendNotification(vscode_languageserver_protocol_1.DidCloseTextDocumentNotification.type, params),
        didSaveTextDocument: (params) => connection.sendNotification(vscode_languageserver_protocol_1.DidSaveTextDocumentNotification.type, params),
        onDiagnostics: (handler) => connection.onNotification(vscode_languageserver_protocol_1.PublishDiagnosticsNotification.type, handler),
        end: () => connection.end(),
        dispose: () => connection.dispose()
      };
      return result;
    }
    var ErrorAction;
    (function(ErrorAction2) {
      ErrorAction2[ErrorAction2["Continue"] = 1] = "Continue";
      ErrorAction2[ErrorAction2["Shutdown"] = 2] = "Shutdown";
    })(ErrorAction = exports2.ErrorAction || (exports2.ErrorAction = {}));
    var CloseAction;
    (function(CloseAction2) {
      CloseAction2[CloseAction2["DoNotRestart"] = 1] = "DoNotRestart";
      CloseAction2[CloseAction2["Restart"] = 2] = "Restart";
    })(CloseAction = exports2.CloseAction || (exports2.CloseAction = {}));
    var DefaultErrorHandler = class {
      constructor(name, maxRestartCount) {
        this.name = name;
        this.maxRestartCount = maxRestartCount;
        this.restarts = [];
      }
      error(_error, _message, count) {
        if (count && count <= 3) {
          return ErrorAction.Continue;
        }
        return ErrorAction.Shutdown;
      }
      closed() {
        this.restarts.push(Date.now());
        if (this.restarts.length <= this.maxRestartCount) {
          return CloseAction.Restart;
        } else {
          let diff = this.restarts[this.restarts.length - 1] - this.restarts[0];
          if (diff <= 3 * 60 * 1e3) {
            vscode_1.window.showErrorMessage(`The ${this.name} server crashed ${this.maxRestartCount + 1} times in the last 3 minutes. The server will not be restarted.`);
            return CloseAction.DoNotRestart;
          } else {
            this.restarts.shift();
            return CloseAction.Restart;
          }
        }
      }
    };
    var RevealOutputChannelOn;
    (function(RevealOutputChannelOn2) {
      RevealOutputChannelOn2[RevealOutputChannelOn2["Info"] = 1] = "Info";
      RevealOutputChannelOn2[RevealOutputChannelOn2["Warn"] = 2] = "Warn";
      RevealOutputChannelOn2[RevealOutputChannelOn2["Error"] = 3] = "Error";
      RevealOutputChannelOn2[RevealOutputChannelOn2["Never"] = 4] = "Never";
    })(RevealOutputChannelOn = exports2.RevealOutputChannelOn || (exports2.RevealOutputChannelOn = {}));
    var State;
    (function(State2) {
      State2[State2["Stopped"] = 1] = "Stopped";
      State2[State2["Starting"] = 3] = "Starting";
      State2[State2["Running"] = 2] = "Running";
    })(State = exports2.State || (exports2.State = {}));
    var ClientState;
    (function(ClientState2) {
      ClientState2[ClientState2["Initial"] = 0] = "Initial";
      ClientState2[ClientState2["Starting"] = 1] = "Starting";
      ClientState2[ClientState2["StartFailed"] = 2] = "StartFailed";
      ClientState2[ClientState2["Running"] = 3] = "Running";
      ClientState2[ClientState2["Stopping"] = 4] = "Stopping";
      ClientState2[ClientState2["Stopped"] = 5] = "Stopped";
    })(ClientState || (ClientState = {}));
    var SupportedSymbolKinds = [
      vscode_languageserver_protocol_1.SymbolKind.File,
      vscode_languageserver_protocol_1.SymbolKind.Module,
      vscode_languageserver_protocol_1.SymbolKind.Namespace,
      vscode_languageserver_protocol_1.SymbolKind.Package,
      vscode_languageserver_protocol_1.SymbolKind.Class,
      vscode_languageserver_protocol_1.SymbolKind.Method,
      vscode_languageserver_protocol_1.SymbolKind.Property,
      vscode_languageserver_protocol_1.SymbolKind.Field,
      vscode_languageserver_protocol_1.SymbolKind.Constructor,
      vscode_languageserver_protocol_1.SymbolKind.Enum,
      vscode_languageserver_protocol_1.SymbolKind.Interface,
      vscode_languageserver_protocol_1.SymbolKind.Function,
      vscode_languageserver_protocol_1.SymbolKind.Variable,
      vscode_languageserver_protocol_1.SymbolKind.Constant,
      vscode_languageserver_protocol_1.SymbolKind.String,
      vscode_languageserver_protocol_1.SymbolKind.Number,
      vscode_languageserver_protocol_1.SymbolKind.Boolean,
      vscode_languageserver_protocol_1.SymbolKind.Array,
      vscode_languageserver_protocol_1.SymbolKind.Object,
      vscode_languageserver_protocol_1.SymbolKind.Key,
      vscode_languageserver_protocol_1.SymbolKind.Null,
      vscode_languageserver_protocol_1.SymbolKind.EnumMember,
      vscode_languageserver_protocol_1.SymbolKind.Struct,
      vscode_languageserver_protocol_1.SymbolKind.Event,
      vscode_languageserver_protocol_1.SymbolKind.Operator,
      vscode_languageserver_protocol_1.SymbolKind.TypeParameter
    ];
    var SupportedCompletionItemKinds = [
      vscode_languageserver_protocol_1.CompletionItemKind.Text,
      vscode_languageserver_protocol_1.CompletionItemKind.Method,
      vscode_languageserver_protocol_1.CompletionItemKind.Function,
      vscode_languageserver_protocol_1.CompletionItemKind.Constructor,
      vscode_languageserver_protocol_1.CompletionItemKind.Field,
      vscode_languageserver_protocol_1.CompletionItemKind.Variable,
      vscode_languageserver_protocol_1.CompletionItemKind.Class,
      vscode_languageserver_protocol_1.CompletionItemKind.Interface,
      vscode_languageserver_protocol_1.CompletionItemKind.Module,
      vscode_languageserver_protocol_1.CompletionItemKind.Property,
      vscode_languageserver_protocol_1.CompletionItemKind.Unit,
      vscode_languageserver_protocol_1.CompletionItemKind.Value,
      vscode_languageserver_protocol_1.CompletionItemKind.Enum,
      vscode_languageserver_protocol_1.CompletionItemKind.Keyword,
      vscode_languageserver_protocol_1.CompletionItemKind.Snippet,
      vscode_languageserver_protocol_1.CompletionItemKind.Color,
      vscode_languageserver_protocol_1.CompletionItemKind.File,
      vscode_languageserver_protocol_1.CompletionItemKind.Reference,
      vscode_languageserver_protocol_1.CompletionItemKind.Folder,
      vscode_languageserver_protocol_1.CompletionItemKind.EnumMember,
      vscode_languageserver_protocol_1.CompletionItemKind.Constant,
      vscode_languageserver_protocol_1.CompletionItemKind.Struct,
      vscode_languageserver_protocol_1.CompletionItemKind.Event,
      vscode_languageserver_protocol_1.CompletionItemKind.Operator,
      vscode_languageserver_protocol_1.CompletionItemKind.TypeParameter
    ];
    var SupportedSymbolTags = [
      vscode_languageserver_protocol_1.SymbolTag.Deprecated
    ];
    function ensure(target, key) {
      if (target[key] === void 0) {
        target[key] = {};
      }
      return target[key];
    }
    var FileFormattingOptions;
    (function(FileFormattingOptions2) {
      function fromConfiguration(document) {
        const filesConfig = vscode_1.workspace.getConfiguration("files", document);
        return {
          trimTrailingWhitespace: filesConfig.get("trimTrailingWhitespace"),
          trimFinalNewlines: filesConfig.get("trimFinalNewlines"),
          insertFinalNewline: filesConfig.get("insertFinalNewline")
        };
      }
      FileFormattingOptions2.fromConfiguration = fromConfiguration;
    })(FileFormattingOptions || (FileFormattingOptions = {}));
    var DynamicFeature;
    (function(DynamicFeature2) {
      function is(value) {
        let candidate = value;
        return candidate && Is.func(candidate.register) && Is.func(candidate.unregister) && Is.func(candidate.dispose) && candidate.registrationType !== void 0;
      }
      DynamicFeature2.is = is;
    })(DynamicFeature || (DynamicFeature = {}));
    var DocumentNotifications = class {
      constructor(_client, _event, _type, _middleware, _createParams, _selectorFilter) {
        this._client = _client;
        this._event = _event;
        this._type = _type;
        this._middleware = _middleware;
        this._createParams = _createParams;
        this._selectorFilter = _selectorFilter;
        this._selectors = /* @__PURE__ */ new Map();
      }
      static textDocumentFilter(selectors, textDocument) {
        for (const selector of selectors) {
          if (vscode_1.languages.match(selector, textDocument)) {
            return true;
          }
        }
        return false;
      }
      register(data) {
        if (!data.registerOptions.documentSelector) {
          return;
        }
        if (!this._listener) {
          this._listener = this._event(this.callback, this);
        }
        this._selectors.set(data.id, data.registerOptions.documentSelector);
      }
      callback(data) {
        if (!this._selectorFilter || this._selectorFilter(this._selectors.values(), data)) {
          if (this._middleware) {
            this._middleware(data, (data2) => this._client.sendNotification(this._type, this._createParams(data2)));
          } else {
            this._client.sendNotification(this._type, this._createParams(data));
          }
          this.notificationSent(data);
        }
      }
      notificationSent(_data) {
      }
      unregister(id) {
        this._selectors.delete(id);
        if (this._selectors.size === 0 && this._listener) {
          this._listener.dispose();
          this._listener = void 0;
        }
      }
      dispose() {
        this._selectors.clear();
        if (this._listener) {
          this._listener.dispose();
          this._listener = void 0;
        }
      }
      getProvider(document) {
        for (const selector of this._selectors.values()) {
          if (vscode_1.languages.match(selector, document)) {
            return {
              send: (data) => {
                this.callback(data);
              }
            };
          }
        }
        return void 0;
      }
    };
    var DidOpenTextDocumentFeature = class extends DocumentNotifications {
      constructor(client2, _syncedDocuments) {
        super(client2, vscode_1.workspace.onDidOpenTextDocument, vscode_languageserver_protocol_1.DidOpenTextDocumentNotification.type, client2.clientOptions.middleware.didOpen, (textDocument) => client2.code2ProtocolConverter.asOpenTextDocumentParams(textDocument), DocumentNotifications.textDocumentFilter);
        this._syncedDocuments = _syncedDocuments;
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "textDocument"), "synchronization").dynamicRegistration = true;
      }
      initialize(capabilities, documentSelector) {
        let textDocumentSyncOptions = capabilities.resolvedTextDocumentSync;
        if (documentSelector && textDocumentSyncOptions && textDocumentSyncOptions.openClose) {
          this.register({ id: UUID.generateUuid(), registerOptions: { documentSelector } });
        }
      }
      get registrationType() {
        return vscode_languageserver_protocol_1.DidOpenTextDocumentNotification.type;
      }
      register(data) {
        super.register(data);
        if (!data.registerOptions.documentSelector) {
          return;
        }
        let documentSelector = data.registerOptions.documentSelector;
        vscode_1.workspace.textDocuments.forEach((textDocument) => {
          let uri = textDocument.uri.toString();
          if (this._syncedDocuments.has(uri)) {
            return;
          }
          if (vscode_1.languages.match(documentSelector, textDocument)) {
            let middleware = this._client.clientOptions.middleware;
            let didOpen = (textDocument2) => {
              this._client.sendNotification(this._type, this._createParams(textDocument2));
            };
            if (middleware.didOpen) {
              middleware.didOpen(textDocument, didOpen);
            } else {
              didOpen(textDocument);
            }
            this._syncedDocuments.set(uri, textDocument);
          }
        });
      }
      notificationSent(textDocument) {
        super.notificationSent(textDocument);
        this._syncedDocuments.set(textDocument.uri.toString(), textDocument);
      }
    };
    var DidCloseTextDocumentFeature = class extends DocumentNotifications {
      constructor(client2, _syncedDocuments) {
        super(client2, vscode_1.workspace.onDidCloseTextDocument, vscode_languageserver_protocol_1.DidCloseTextDocumentNotification.type, client2.clientOptions.middleware.didClose, (textDocument) => client2.code2ProtocolConverter.asCloseTextDocumentParams(textDocument), DocumentNotifications.textDocumentFilter);
        this._syncedDocuments = _syncedDocuments;
      }
      get registrationType() {
        return vscode_languageserver_protocol_1.DidCloseTextDocumentNotification.type;
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "textDocument"), "synchronization").dynamicRegistration = true;
      }
      initialize(capabilities, documentSelector) {
        let textDocumentSyncOptions = capabilities.resolvedTextDocumentSync;
        if (documentSelector && textDocumentSyncOptions && textDocumentSyncOptions.openClose) {
          this.register({ id: UUID.generateUuid(), registerOptions: { documentSelector } });
        }
      }
      notificationSent(textDocument) {
        super.notificationSent(textDocument);
        this._syncedDocuments.delete(textDocument.uri.toString());
      }
      unregister(id) {
        let selector = this._selectors.get(id);
        super.unregister(id);
        let selectors = this._selectors.values();
        this._syncedDocuments.forEach((textDocument) => {
          if (vscode_1.languages.match(selector, textDocument) && !this._selectorFilter(selectors, textDocument)) {
            let middleware = this._client.clientOptions.middleware;
            let didClose = (textDocument2) => {
              this._client.sendNotification(this._type, this._createParams(textDocument2));
            };
            this._syncedDocuments.delete(textDocument.uri.toString());
            if (middleware.didClose) {
              middleware.didClose(textDocument, didClose);
            } else {
              didClose(textDocument);
            }
          }
        });
      }
    };
    var DidChangeTextDocumentFeature = class {
      constructor(_client) {
        this._client = _client;
        this._changeData = /* @__PURE__ */ new Map();
        this._forcingDelivery = false;
      }
      get registrationType() {
        return vscode_languageserver_protocol_1.DidChangeTextDocumentNotification.type;
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "textDocument"), "synchronization").dynamicRegistration = true;
      }
      initialize(capabilities, documentSelector) {
        let textDocumentSyncOptions = capabilities.resolvedTextDocumentSync;
        if (documentSelector && textDocumentSyncOptions && textDocumentSyncOptions.change !== void 0 && textDocumentSyncOptions.change !== vscode_languageserver_protocol_1.TextDocumentSyncKind.None) {
          this.register({
            id: UUID.generateUuid(),
            registerOptions: Object.assign({}, { documentSelector }, { syncKind: textDocumentSyncOptions.change })
          });
        }
      }
      register(data) {
        if (!data.registerOptions.documentSelector) {
          return;
        }
        if (!this._listener) {
          this._listener = vscode_1.workspace.onDidChangeTextDocument(this.callback, this);
        }
        this._changeData.set(data.id, {
          documentSelector: data.registerOptions.documentSelector,
          syncKind: data.registerOptions.syncKind
        });
      }
      callback(event) {
        if (event.contentChanges.length === 0) {
          return;
        }
        for (const changeData of this._changeData.values()) {
          if (vscode_1.languages.match(changeData.documentSelector, event.document)) {
            let middleware = this._client.clientOptions.middleware;
            if (changeData.syncKind === vscode_languageserver_protocol_1.TextDocumentSyncKind.Incremental) {
              let params = this._client.code2ProtocolConverter.asChangeTextDocumentParams(event);
              if (middleware.didChange) {
                middleware.didChange(event, () => this._client.sendNotification(vscode_languageserver_protocol_1.DidChangeTextDocumentNotification.type, params));
              } else {
                this._client.sendNotification(vscode_languageserver_protocol_1.DidChangeTextDocumentNotification.type, params);
              }
            } else if (changeData.syncKind === vscode_languageserver_protocol_1.TextDocumentSyncKind.Full) {
              let didChange = (event2) => {
                if (this._changeDelayer) {
                  if (this._changeDelayer.uri !== event2.document.uri.toString()) {
                    this.forceDelivery();
                    this._changeDelayer.uri = event2.document.uri.toString();
                  }
                  this._changeDelayer.delayer.trigger(() => {
                    this._client.sendNotification(vscode_languageserver_protocol_1.DidChangeTextDocumentNotification.type, this._client.code2ProtocolConverter.asChangeTextDocumentParams(event2.document));
                  });
                } else {
                  this._changeDelayer = {
                    uri: event2.document.uri.toString(),
                    delayer: new async_1.Delayer(200)
                  };
                  this._changeDelayer.delayer.trigger(() => {
                    this._client.sendNotification(vscode_languageserver_protocol_1.DidChangeTextDocumentNotification.type, this._client.code2ProtocolConverter.asChangeTextDocumentParams(event2.document));
                  }, -1);
                }
              };
              if (middleware.didChange) {
                middleware.didChange(event, didChange);
              } else {
                didChange(event);
              }
            }
          }
        }
      }
      unregister(id) {
        this._changeData.delete(id);
        if (this._changeData.size === 0 && this._listener) {
          this._listener.dispose();
          this._listener = void 0;
        }
      }
      dispose() {
        this._changeDelayer = void 0;
        this._forcingDelivery = false;
        this._changeData.clear();
        if (this._listener) {
          this._listener.dispose();
          this._listener = void 0;
        }
      }
      forceDelivery() {
        if (this._forcingDelivery || !this._changeDelayer) {
          return;
        }
        try {
          this._forcingDelivery = true;
          this._changeDelayer.delayer.forceDelivery();
        } finally {
          this._forcingDelivery = false;
        }
      }
      getProvider(document) {
        for (const changeData of this._changeData.values()) {
          if (vscode_1.languages.match(changeData.documentSelector, document)) {
            return {
              send: (event) => {
                this.callback(event);
              }
            };
          }
        }
        return void 0;
      }
    };
    var WillSaveFeature = class extends DocumentNotifications {
      constructor(client2) {
        super(client2, vscode_1.workspace.onWillSaveTextDocument, vscode_languageserver_protocol_1.WillSaveTextDocumentNotification.type, client2.clientOptions.middleware.willSave, (willSaveEvent) => client2.code2ProtocolConverter.asWillSaveTextDocumentParams(willSaveEvent), (selectors, willSaveEvent) => DocumentNotifications.textDocumentFilter(selectors, willSaveEvent.document));
      }
      get registrationType() {
        return vscode_languageserver_protocol_1.WillSaveTextDocumentNotification.type;
      }
      fillClientCapabilities(capabilities) {
        let value = ensure(ensure(capabilities, "textDocument"), "synchronization");
        value.willSave = true;
      }
      initialize(capabilities, documentSelector) {
        let textDocumentSyncOptions = capabilities.resolvedTextDocumentSync;
        if (documentSelector && textDocumentSyncOptions && textDocumentSyncOptions.willSave) {
          this.register({
            id: UUID.generateUuid(),
            registerOptions: { documentSelector }
          });
        }
      }
    };
    var WillSaveWaitUntilFeature = class {
      constructor(_client) {
        this._client = _client;
        this._selectors = /* @__PURE__ */ new Map();
      }
      get registrationType() {
        return vscode_languageserver_protocol_1.WillSaveTextDocumentWaitUntilRequest.type;
      }
      fillClientCapabilities(capabilities) {
        let value = ensure(ensure(capabilities, "textDocument"), "synchronization");
        value.willSaveWaitUntil = true;
      }
      initialize(capabilities, documentSelector) {
        let textDocumentSyncOptions = capabilities.resolvedTextDocumentSync;
        if (documentSelector && textDocumentSyncOptions && textDocumentSyncOptions.willSaveWaitUntil) {
          this.register({
            id: UUID.generateUuid(),
            registerOptions: { documentSelector }
          });
        }
      }
      register(data) {
        if (!data.registerOptions.documentSelector) {
          return;
        }
        if (!this._listener) {
          this._listener = vscode_1.workspace.onWillSaveTextDocument(this.callback, this);
        }
        this._selectors.set(data.id, data.registerOptions.documentSelector);
      }
      callback(event) {
        if (DocumentNotifications.textDocumentFilter(this._selectors.values(), event.document)) {
          let middleware = this._client.clientOptions.middleware;
          let willSaveWaitUntil = (event2) => {
            return this._client.sendRequest(vscode_languageserver_protocol_1.WillSaveTextDocumentWaitUntilRequest.type, this._client.code2ProtocolConverter.asWillSaveTextDocumentParams(event2)).then((edits) => {
              let vEdits = this._client.protocol2CodeConverter.asTextEdits(edits);
              return vEdits === void 0 ? [] : vEdits;
            });
          };
          event.waitUntil(middleware.willSaveWaitUntil ? middleware.willSaveWaitUntil(event, willSaveWaitUntil) : willSaveWaitUntil(event));
        }
      }
      unregister(id) {
        this._selectors.delete(id);
        if (this._selectors.size === 0 && this._listener) {
          this._listener.dispose();
          this._listener = void 0;
        }
      }
      dispose() {
        this._selectors.clear();
        if (this._listener) {
          this._listener.dispose();
          this._listener = void 0;
        }
      }
    };
    var DidSaveTextDocumentFeature = class extends DocumentNotifications {
      constructor(client2) {
        super(client2, vscode_1.workspace.onDidSaveTextDocument, vscode_languageserver_protocol_1.DidSaveTextDocumentNotification.type, client2.clientOptions.middleware.didSave, (textDocument) => client2.code2ProtocolConverter.asSaveTextDocumentParams(textDocument, this._includeText), DocumentNotifications.textDocumentFilter);
        this._includeText = false;
      }
      get registrationType() {
        return vscode_languageserver_protocol_1.DidSaveTextDocumentNotification.type;
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "textDocument"), "synchronization").didSave = true;
      }
      initialize(capabilities, documentSelector) {
        const textDocumentSyncOptions = capabilities.resolvedTextDocumentSync;
        if (documentSelector && textDocumentSyncOptions && textDocumentSyncOptions.save) {
          const saveOptions = typeof textDocumentSyncOptions.save === "boolean" ? { includeText: false } : { includeText: !!textDocumentSyncOptions.save.includeText };
          this.register({
            id: UUID.generateUuid(),
            registerOptions: Object.assign({}, { documentSelector }, saveOptions)
          });
        }
      }
      register(data) {
        this._includeText = !!data.registerOptions.includeText;
        super.register(data);
      }
    };
    var FileSystemWatcherFeature = class {
      constructor(_client, _notifyFileEvent) {
        this._client = _client;
        this._notifyFileEvent = _notifyFileEvent;
        this._watchers = /* @__PURE__ */ new Map();
      }
      get registrationType() {
        return vscode_languageserver_protocol_1.DidChangeWatchedFilesNotification.type;
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "workspace"), "didChangeWatchedFiles").dynamicRegistration = true;
      }
      initialize(_capabilities, _documentSelector) {
      }
      register(data) {
        if (!Array.isArray(data.registerOptions.watchers)) {
          return;
        }
        let disposables = [];
        for (let watcher of data.registerOptions.watchers) {
          if (!Is.string(watcher.globPattern)) {
            continue;
          }
          let watchCreate = true, watchChange = true, watchDelete = true;
          if (watcher.kind !== void 0 && watcher.kind !== null) {
            watchCreate = (watcher.kind & vscode_languageserver_protocol_1.WatchKind.Create) !== 0;
            watchChange = (watcher.kind & vscode_languageserver_protocol_1.WatchKind.Change) !== 0;
            watchDelete = (watcher.kind & vscode_languageserver_protocol_1.WatchKind.Delete) !== 0;
          }
          let fileSystemWatcher = vscode_1.workspace.createFileSystemWatcher(watcher.globPattern, !watchCreate, !watchChange, !watchDelete);
          this.hookListeners(fileSystemWatcher, watchCreate, watchChange, watchDelete);
          disposables.push(fileSystemWatcher);
        }
        this._watchers.set(data.id, disposables);
      }
      registerRaw(id, fileSystemWatchers) {
        let disposables = [];
        for (let fileSystemWatcher of fileSystemWatchers) {
          this.hookListeners(fileSystemWatcher, true, true, true, disposables);
        }
        this._watchers.set(id, disposables);
      }
      hookListeners(fileSystemWatcher, watchCreate, watchChange, watchDelete, listeners) {
        if (watchCreate) {
          fileSystemWatcher.onDidCreate((resource) => this._notifyFileEvent({
            uri: this._client.code2ProtocolConverter.asUri(resource),
            type: vscode_languageserver_protocol_1.FileChangeType.Created
          }), null, listeners);
        }
        if (watchChange) {
          fileSystemWatcher.onDidChange((resource) => this._notifyFileEvent({
            uri: this._client.code2ProtocolConverter.asUri(resource),
            type: vscode_languageserver_protocol_1.FileChangeType.Changed
          }), null, listeners);
        }
        if (watchDelete) {
          fileSystemWatcher.onDidDelete((resource) => this._notifyFileEvent({
            uri: this._client.code2ProtocolConverter.asUri(resource),
            type: vscode_languageserver_protocol_1.FileChangeType.Deleted
          }), null, listeners);
        }
      }
      unregister(id) {
        let disposables = this._watchers.get(id);
        if (disposables) {
          for (let disposable of disposables) {
            disposable.dispose();
          }
        }
      }
      dispose() {
        this._watchers.forEach((disposables) => {
          for (let disposable of disposables) {
            disposable.dispose();
          }
        });
        this._watchers.clear();
      }
    };
    var TextDocumentFeature = class {
      constructor(_client, _registrationType) {
        this._client = _client;
        this._registrationType = _registrationType;
        this._registrations = /* @__PURE__ */ new Map();
      }
      get registrationType() {
        return this._registrationType;
      }
      register(data) {
        if (!data.registerOptions.documentSelector) {
          return;
        }
        let registration = this.registerLanguageProvider(data.registerOptions);
        this._registrations.set(data.id, { disposable: registration[0], data, provider: registration[1] });
      }
      unregister(id) {
        let registration = this._registrations.get(id);
        if (registration !== void 0) {
          registration.disposable.dispose();
        }
      }
      dispose() {
        this._registrations.forEach((value) => {
          value.disposable.dispose();
        });
        this._registrations.clear();
      }
      getRegistration(documentSelector, capability) {
        if (!capability) {
          return [void 0, void 0];
        } else if (vscode_languageserver_protocol_1.TextDocumentRegistrationOptions.is(capability)) {
          const id = vscode_languageserver_protocol_1.StaticRegistrationOptions.hasId(capability) ? capability.id : UUID.generateUuid();
          const selector = capability.documentSelector || documentSelector;
          if (selector) {
            return [id, Object.assign({}, capability, { documentSelector: selector })];
          }
        } else if (Is.boolean(capability) && capability === true || vscode_languageserver_protocol_1.WorkDoneProgressOptions.is(capability)) {
          if (!documentSelector) {
            return [void 0, void 0];
          }
          let options = Is.boolean(capability) && capability === true ? { documentSelector } : Object.assign({}, capability, { documentSelector });
          return [UUID.generateUuid(), options];
        }
        return [void 0, void 0];
      }
      getRegistrationOptions(documentSelector, capability) {
        if (!documentSelector || !capability) {
          return void 0;
        }
        return Is.boolean(capability) && capability === true ? { documentSelector } : Object.assign({}, capability, { documentSelector });
      }
      getProvider(textDocument) {
        for (const registration of this._registrations.values()) {
          let selector = registration.data.registerOptions.documentSelector;
          if (selector !== null && vscode_1.languages.match(selector, textDocument)) {
            return registration.provider;
          }
        }
        return void 0;
      }
      getAllProviders() {
        const result = [];
        for (const item of this._registrations.values()) {
          result.push(item.provider);
        }
        return result;
      }
    };
    exports2.TextDocumentFeature = TextDocumentFeature;
    var WorkspaceFeature = class {
      constructor(_client, _registrationType) {
        this._client = _client;
        this._registrationType = _registrationType;
        this._registrations = /* @__PURE__ */ new Map();
      }
      get registrationType() {
        return this._registrationType;
      }
      register(data) {
        const registration = this.registerLanguageProvider(data.registerOptions);
        this._registrations.set(data.id, { disposable: registration[0], provider: registration[1] });
      }
      unregister(id) {
        let registration = this._registrations.get(id);
        if (registration !== void 0) {
          registration.disposable.dispose();
        }
      }
      dispose() {
        this._registrations.forEach((registration) => {
          registration.disposable.dispose();
        });
        this._registrations.clear();
      }
      getProviders() {
        const result = [];
        for (const registration of this._registrations.values()) {
          result.push(registration.provider);
        }
        return result;
      }
    };
    var CompletionItemFeature = class extends TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.CompletionRequest.type);
      }
      fillClientCapabilities(capabilities) {
        let completion = ensure(ensure(capabilities, "textDocument"), "completion");
        completion.dynamicRegistration = true;
        completion.contextSupport = true;
        completion.completionItem = {
          snippetSupport: true,
          commitCharactersSupport: true,
          documentationFormat: [vscode_languageserver_protocol_1.MarkupKind.Markdown, vscode_languageserver_protocol_1.MarkupKind.PlainText],
          deprecatedSupport: true,
          preselectSupport: true,
          tagSupport: { valueSet: [vscode_languageserver_protocol_1.CompletionItemTag.Deprecated] },
          insertReplaceSupport: true,
          resolveSupport: {
            properties: ["documentation", "detail", "additionalTextEdits"]
          },
          insertTextModeSupport: { valueSet: [vscode_languageserver_protocol_1.InsertTextMode.asIs, vscode_languageserver_protocol_1.InsertTextMode.adjustIndentation] }
        };
        completion.completionItemKind = { valueSet: SupportedCompletionItemKinds };
      }
      initialize(capabilities, documentSelector) {
        const options = this.getRegistrationOptions(documentSelector, capabilities.completionProvider);
        if (!options) {
          return;
        }
        this.register({
          id: UUID.generateUuid(),
          registerOptions: options
        });
      }
      registerLanguageProvider(options) {
        const triggerCharacters = options.triggerCharacters || [];
        const provider = {
          provideCompletionItems: (document, position, token, context) => {
            const client2 = this._client;
            const middleware = this._client.clientOptions.middleware;
            const provideCompletionItems = (document2, position2, context2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.CompletionRequest.type, client2.code2ProtocolConverter.asCompletionParams(document2, position2, context2), token2).then(client2.protocol2CodeConverter.asCompletionResult, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.CompletionRequest.type, error, null);
              });
            };
            return middleware.provideCompletionItem ? middleware.provideCompletionItem(document, position, context, token, provideCompletionItems) : provideCompletionItems(document, position, context, token);
          },
          resolveCompletionItem: options.resolveProvider ? (item, token) => {
            const client2 = this._client;
            const middleware = this._client.clientOptions.middleware;
            const resolveCompletionItem = (item2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.CompletionResolveRequest.type, client2.code2ProtocolConverter.asCompletionItem(item2), token2).then(client2.protocol2CodeConverter.asCompletionItem, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.CompletionResolveRequest.type, error, item2);
              });
            };
            return middleware.resolveCompletionItem ? middleware.resolveCompletionItem(item, token, resolveCompletionItem) : resolveCompletionItem(item, token);
          } : void 0
        };
        return [vscode_1.languages.registerCompletionItemProvider(options.documentSelector, provider, ...triggerCharacters), provider];
      }
    };
    var HoverFeature = class extends TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.HoverRequest.type);
      }
      fillClientCapabilities(capabilities) {
        const hoverCapability = ensure(ensure(capabilities, "textDocument"), "hover");
        hoverCapability.dynamicRegistration = true;
        hoverCapability.contentFormat = [vscode_languageserver_protocol_1.MarkupKind.Markdown, vscode_languageserver_protocol_1.MarkupKind.PlainText];
      }
      initialize(capabilities, documentSelector) {
        const options = this.getRegistrationOptions(documentSelector, capabilities.hoverProvider);
        if (!options) {
          return;
        }
        this.register({
          id: UUID.generateUuid(),
          registerOptions: options
        });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideHover: (document, position, token) => {
            const client2 = this._client;
            const provideHover = (document2, position2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.HoverRequest.type, client2.code2ProtocolConverter.asTextDocumentPositionParams(document2, position2), token2).then(client2.protocol2CodeConverter.asHover, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.HoverRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideHover ? middleware.provideHover(document, position, token, provideHover) : provideHover(document, position, token);
          }
        };
        return [vscode_1.languages.registerHoverProvider(options.documentSelector, provider), provider];
      }
    };
    var SignatureHelpFeature = class extends TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.SignatureHelpRequest.type);
      }
      fillClientCapabilities(capabilities) {
        let config = ensure(ensure(capabilities, "textDocument"), "signatureHelp");
        config.dynamicRegistration = true;
        config.signatureInformation = { documentationFormat: [vscode_languageserver_protocol_1.MarkupKind.Markdown, vscode_languageserver_protocol_1.MarkupKind.PlainText] };
        config.signatureInformation.parameterInformation = { labelOffsetSupport: true };
        config.signatureInformation.activeParameterSupport = true;
        config.contextSupport = true;
      }
      initialize(capabilities, documentSelector) {
        const options = this.getRegistrationOptions(documentSelector, capabilities.signatureHelpProvider);
        if (!options) {
          return;
        }
        this.register({
          id: UUID.generateUuid(),
          registerOptions: options
        });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideSignatureHelp: (document, position, token, context) => {
            const client2 = this._client;
            const providerSignatureHelp = (document2, position2, context2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.SignatureHelpRequest.type, client2.code2ProtocolConverter.asSignatureHelpParams(document2, position2, context2), token2).then(client2.protocol2CodeConverter.asSignatureHelp, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.SignatureHelpRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideSignatureHelp ? middleware.provideSignatureHelp(document, position, context, token, providerSignatureHelp) : providerSignatureHelp(document, position, context, token);
          }
        };
        let disposable;
        if (options.retriggerCharacters === void 0) {
          const triggerCharacters = options.triggerCharacters || [];
          disposable = vscode_1.languages.registerSignatureHelpProvider(options.documentSelector, provider, ...triggerCharacters);
        } else {
          const metaData = {
            triggerCharacters: options.triggerCharacters || [],
            retriggerCharacters: options.retriggerCharacters || []
          };
          disposable = vscode_1.languages.registerSignatureHelpProvider(options.documentSelector, provider, metaData);
        }
        return [disposable, provider];
      }
    };
    var DefinitionFeature = class extends TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.DefinitionRequest.type);
      }
      fillClientCapabilities(capabilities) {
        let definitionSupport = ensure(ensure(capabilities, "textDocument"), "definition");
        definitionSupport.dynamicRegistration = true;
        definitionSupport.linkSupport = true;
      }
      initialize(capabilities, documentSelector) {
        const options = this.getRegistrationOptions(documentSelector, capabilities.definitionProvider);
        if (!options) {
          return;
        }
        this.register({ id: UUID.generateUuid(), registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideDefinition: (document, position, token) => {
            const client2 = this._client;
            const provideDefinition = (document2, position2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.DefinitionRequest.type, client2.code2ProtocolConverter.asTextDocumentPositionParams(document2, position2), token2).then(client2.protocol2CodeConverter.asDefinitionResult, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.DefinitionRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideDefinition ? middleware.provideDefinition(document, position, token, provideDefinition) : provideDefinition(document, position, token);
          }
        };
        return [vscode_1.languages.registerDefinitionProvider(options.documentSelector, provider), provider];
      }
    };
    var ReferencesFeature = class extends TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.ReferencesRequest.type);
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "textDocument"), "references").dynamicRegistration = true;
      }
      initialize(capabilities, documentSelector) {
        const options = this.getRegistrationOptions(documentSelector, capabilities.referencesProvider);
        if (!options) {
          return;
        }
        this.register({ id: UUID.generateUuid(), registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideReferences: (document, position, options2, token) => {
            const client2 = this._client;
            const _providerReferences = (document2, position2, options3, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.ReferencesRequest.type, client2.code2ProtocolConverter.asReferenceParams(document2, position2, options3), token2).then(client2.protocol2CodeConverter.asReferences, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.ReferencesRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideReferences ? middleware.provideReferences(document, position, options2, token, _providerReferences) : _providerReferences(document, position, options2, token);
          }
        };
        return [vscode_1.languages.registerReferenceProvider(options.documentSelector, provider), provider];
      }
    };
    var DocumentHighlightFeature = class extends TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.DocumentHighlightRequest.type);
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "textDocument"), "documentHighlight").dynamicRegistration = true;
      }
      initialize(capabilities, documentSelector) {
        const options = this.getRegistrationOptions(documentSelector, capabilities.documentHighlightProvider);
        if (!options) {
          return;
        }
        this.register({ id: UUID.generateUuid(), registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideDocumentHighlights: (document, position, token) => {
            const client2 = this._client;
            const _provideDocumentHighlights = (document2, position2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.DocumentHighlightRequest.type, client2.code2ProtocolConverter.asTextDocumentPositionParams(document2, position2), token2).then(client2.protocol2CodeConverter.asDocumentHighlights, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.DocumentHighlightRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideDocumentHighlights ? middleware.provideDocumentHighlights(document, position, token, _provideDocumentHighlights) : _provideDocumentHighlights(document, position, token);
          }
        };
        return [vscode_1.languages.registerDocumentHighlightProvider(options.documentSelector, provider), provider];
      }
    };
    var DocumentSymbolFeature = class extends TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.DocumentSymbolRequest.type);
      }
      fillClientCapabilities(capabilities) {
        let symbolCapabilities = ensure(ensure(capabilities, "textDocument"), "documentSymbol");
        symbolCapabilities.dynamicRegistration = true;
        symbolCapabilities.symbolKind = {
          valueSet: SupportedSymbolKinds
        };
        symbolCapabilities.hierarchicalDocumentSymbolSupport = true;
        symbolCapabilities.tagSupport = {
          valueSet: SupportedSymbolTags
        };
        symbolCapabilities.labelSupport = true;
      }
      initialize(capabilities, documentSelector) {
        const options = this.getRegistrationOptions(documentSelector, capabilities.documentSymbolProvider);
        if (!options) {
          return;
        }
        this.register({ id: UUID.generateUuid(), registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideDocumentSymbols: (document, token) => {
            const client2 = this._client;
            const _provideDocumentSymbols = (document2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.DocumentSymbolRequest.type, client2.code2ProtocolConverter.asDocumentSymbolParams(document2), token2).then((data) => {
                if (data === null) {
                  return void 0;
                }
                if (data.length === 0) {
                  return [];
                } else {
                  let element = data[0];
                  if (vscode_languageserver_protocol_1.DocumentSymbol.is(element)) {
                    return client2.protocol2CodeConverter.asDocumentSymbols(data);
                  } else {
                    return client2.protocol2CodeConverter.asSymbolInformations(data);
                  }
                }
              }, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.DocumentSymbolRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideDocumentSymbols ? middleware.provideDocumentSymbols(document, token, _provideDocumentSymbols) : _provideDocumentSymbols(document, token);
          }
        };
        const metaData = options.label !== void 0 ? { label: options.label } : void 0;
        return [vscode_1.languages.registerDocumentSymbolProvider(options.documentSelector, provider, metaData), provider];
      }
    };
    var WorkspaceSymbolFeature = class extends WorkspaceFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.WorkspaceSymbolRequest.type);
      }
      fillClientCapabilities(capabilities) {
        let symbolCapabilities = ensure(ensure(capabilities, "workspace"), "symbol");
        symbolCapabilities.dynamicRegistration = true;
        symbolCapabilities.symbolKind = {
          valueSet: SupportedSymbolKinds
        };
        symbolCapabilities.tagSupport = {
          valueSet: SupportedSymbolTags
        };
      }
      initialize(capabilities) {
        if (!capabilities.workspaceSymbolProvider) {
          return;
        }
        this.register({
          id: UUID.generateUuid(),
          registerOptions: capabilities.workspaceSymbolProvider === true ? { workDoneProgress: false } : capabilities.workspaceSymbolProvider
        });
      }
      registerLanguageProvider(_options) {
        const provider = {
          provideWorkspaceSymbols: (query, token) => {
            const client2 = this._client;
            const provideWorkspaceSymbols = (query2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.WorkspaceSymbolRequest.type, { query: query2 }, token2).then(client2.protocol2CodeConverter.asSymbolInformations, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.WorkspaceSymbolRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideWorkspaceSymbols ? middleware.provideWorkspaceSymbols(query, token, provideWorkspaceSymbols) : provideWorkspaceSymbols(query, token);
          }
        };
        return [vscode_1.languages.registerWorkspaceSymbolProvider(provider), provider];
      }
    };
    var CodeActionFeature = class extends TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.CodeActionRequest.type);
      }
      fillClientCapabilities(capabilities) {
        const cap = ensure(ensure(capabilities, "textDocument"), "codeAction");
        cap.dynamicRegistration = true;
        cap.isPreferredSupport = true;
        cap.disabledSupport = true;
        cap.dataSupport = true;
        cap.resolveSupport = {
          properties: ["edit"]
        };
        cap.codeActionLiteralSupport = {
          codeActionKind: {
            valueSet: [
              vscode_languageserver_protocol_1.CodeActionKind.Empty,
              vscode_languageserver_protocol_1.CodeActionKind.QuickFix,
              vscode_languageserver_protocol_1.CodeActionKind.Refactor,
              vscode_languageserver_protocol_1.CodeActionKind.RefactorExtract,
              vscode_languageserver_protocol_1.CodeActionKind.RefactorInline,
              vscode_languageserver_protocol_1.CodeActionKind.RefactorRewrite,
              vscode_languageserver_protocol_1.CodeActionKind.Source,
              vscode_languageserver_protocol_1.CodeActionKind.SourceOrganizeImports
            ]
          }
        };
        cap.honorsChangeAnnotations = false;
      }
      initialize(capabilities, documentSelector) {
        const options = this.getRegistrationOptions(documentSelector, capabilities.codeActionProvider);
        if (!options) {
          return;
        }
        this.register({ id: UUID.generateUuid(), registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideCodeActions: (document, range, context, token) => {
            const client2 = this._client;
            const _provideCodeActions = (document2, range2, context2, token2) => {
              const params = {
                textDocument: client2.code2ProtocolConverter.asTextDocumentIdentifier(document2),
                range: client2.code2ProtocolConverter.asRange(range2),
                context: client2.code2ProtocolConverter.asCodeActionContext(context2)
              };
              return client2.sendRequest(vscode_languageserver_protocol_1.CodeActionRequest.type, params, token2).then((values) => {
                if (values === null) {
                  return void 0;
                }
                const result = [];
                for (let item of values) {
                  if (vscode_languageserver_protocol_1.Command.is(item)) {
                    result.push(client2.protocol2CodeConverter.asCommand(item));
                  } else {
                    result.push(client2.protocol2CodeConverter.asCodeAction(item));
                  }
                }
                return result;
              }, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.CodeActionRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideCodeActions ? middleware.provideCodeActions(document, range, context, token, _provideCodeActions) : _provideCodeActions(document, range, context, token);
          },
          resolveCodeAction: options.resolveProvider ? (item, token) => {
            const client2 = this._client;
            const middleware = this._client.clientOptions.middleware;
            const resolveCodeAction = (item2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.CodeActionResolveRequest.type, client2.code2ProtocolConverter.asCodeAction(item2), token2).then(client2.protocol2CodeConverter.asCodeAction, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.CodeActionResolveRequest.type, error, item2);
              });
            };
            return middleware.resolveCodeAction ? middleware.resolveCodeAction(item, token, resolveCodeAction) : resolveCodeAction(item, token);
          } : void 0
        };
        return [vscode_1.languages.registerCodeActionsProvider(options.documentSelector, provider, options.codeActionKinds ? { providedCodeActionKinds: this._client.protocol2CodeConverter.asCodeActionKinds(options.codeActionKinds) } : void 0), provider];
      }
    };
    var CodeLensFeature = class extends TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.CodeLensRequest.type);
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "textDocument"), "codeLens").dynamicRegistration = true;
        ensure(ensure(capabilities, "workspace"), "codeLens").refreshSupport = true;
      }
      initialize(capabilities, documentSelector) {
        const client2 = this._client;
        client2.onRequest(vscode_languageserver_protocol_1.CodeLensRefreshRequest.type, async () => {
          for (const provider of this.getAllProviders()) {
            provider.onDidChangeCodeLensEmitter.fire();
          }
        });
        const options = this.getRegistrationOptions(documentSelector, capabilities.codeLensProvider);
        if (!options) {
          return;
        }
        this.register({ id: UUID.generateUuid(), registerOptions: options });
      }
      registerLanguageProvider(options) {
        const eventEmitter = new vscode_1.EventEmitter();
        const provider = {
          onDidChangeCodeLenses: eventEmitter.event,
          provideCodeLenses: (document, token) => {
            const client2 = this._client;
            const provideCodeLenses = (document2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.CodeLensRequest.type, client2.code2ProtocolConverter.asCodeLensParams(document2), token2).then(client2.protocol2CodeConverter.asCodeLenses, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.CodeLensRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideCodeLenses ? middleware.provideCodeLenses(document, token, provideCodeLenses) : provideCodeLenses(document, token);
          },
          resolveCodeLens: options.resolveProvider ? (codeLens, token) => {
            const client2 = this._client;
            const resolveCodeLens = (codeLens2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.CodeLensResolveRequest.type, client2.code2ProtocolConverter.asCodeLens(codeLens2), token2).then(client2.protocol2CodeConverter.asCodeLens, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.CodeLensResolveRequest.type, error, codeLens2);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.resolveCodeLens ? middleware.resolveCodeLens(codeLens, token, resolveCodeLens) : resolveCodeLens(codeLens, token);
          } : void 0
        };
        return [vscode_1.languages.registerCodeLensProvider(options.documentSelector, provider), { provider, onDidChangeCodeLensEmitter: eventEmitter }];
      }
    };
    var DocumentFormattingFeature = class extends TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.DocumentFormattingRequest.type);
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "textDocument"), "formatting").dynamicRegistration = true;
      }
      initialize(capabilities, documentSelector) {
        const options = this.getRegistrationOptions(documentSelector, capabilities.documentFormattingProvider);
        if (!options) {
          return;
        }
        this.register({ id: UUID.generateUuid(), registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideDocumentFormattingEdits: (document, options2, token) => {
            const client2 = this._client;
            const provideDocumentFormattingEdits = (document2, options3, token2) => {
              const params = {
                textDocument: client2.code2ProtocolConverter.asTextDocumentIdentifier(document2),
                options: client2.code2ProtocolConverter.asFormattingOptions(options3, FileFormattingOptions.fromConfiguration(document2))
              };
              return client2.sendRequest(vscode_languageserver_protocol_1.DocumentFormattingRequest.type, params, token2).then(client2.protocol2CodeConverter.asTextEdits, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.DocumentFormattingRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideDocumentFormattingEdits ? middleware.provideDocumentFormattingEdits(document, options2, token, provideDocumentFormattingEdits) : provideDocumentFormattingEdits(document, options2, token);
          }
        };
        return [vscode_1.languages.registerDocumentFormattingEditProvider(options.documentSelector, provider), provider];
      }
    };
    var DocumentRangeFormattingFeature = class extends TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.DocumentRangeFormattingRequest.type);
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "textDocument"), "rangeFormatting").dynamicRegistration = true;
      }
      initialize(capabilities, documentSelector) {
        const options = this.getRegistrationOptions(documentSelector, capabilities.documentRangeFormattingProvider);
        if (!options) {
          return;
        }
        this.register({ id: UUID.generateUuid(), registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideDocumentRangeFormattingEdits: (document, range, options2, token) => {
            const client2 = this._client;
            const provideDocumentRangeFormattingEdits = (document2, range2, options3, token2) => {
              const params = {
                textDocument: client2.code2ProtocolConverter.asTextDocumentIdentifier(document2),
                range: client2.code2ProtocolConverter.asRange(range2),
                options: client2.code2ProtocolConverter.asFormattingOptions(options3, FileFormattingOptions.fromConfiguration(document2))
              };
              return client2.sendRequest(vscode_languageserver_protocol_1.DocumentRangeFormattingRequest.type, params, token2).then(client2.protocol2CodeConverter.asTextEdits, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.DocumentRangeFormattingRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideDocumentRangeFormattingEdits ? middleware.provideDocumentRangeFormattingEdits(document, range, options2, token, provideDocumentRangeFormattingEdits) : provideDocumentRangeFormattingEdits(document, range, options2, token);
          }
        };
        return [vscode_1.languages.registerDocumentRangeFormattingEditProvider(options.documentSelector, provider), provider];
      }
    };
    var DocumentOnTypeFormattingFeature = class extends TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.DocumentOnTypeFormattingRequest.type);
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "textDocument"), "onTypeFormatting").dynamicRegistration = true;
      }
      initialize(capabilities, documentSelector) {
        const options = this.getRegistrationOptions(documentSelector, capabilities.documentOnTypeFormattingProvider);
        if (!options) {
          return;
        }
        this.register({ id: UUID.generateUuid(), registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideOnTypeFormattingEdits: (document, position, ch, options2, token) => {
            const client2 = this._client;
            const provideOnTypeFormattingEdits = (document2, position2, ch2, options3, token2) => {
              let params = {
                textDocument: client2.code2ProtocolConverter.asTextDocumentIdentifier(document2),
                position: client2.code2ProtocolConverter.asPosition(position2),
                ch: ch2,
                options: client2.code2ProtocolConverter.asFormattingOptions(options3, FileFormattingOptions.fromConfiguration(document2))
              };
              return client2.sendRequest(vscode_languageserver_protocol_1.DocumentOnTypeFormattingRequest.type, params, token2).then(client2.protocol2CodeConverter.asTextEdits, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.DocumentOnTypeFormattingRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideOnTypeFormattingEdits ? middleware.provideOnTypeFormattingEdits(document, position, ch, options2, token, provideOnTypeFormattingEdits) : provideOnTypeFormattingEdits(document, position, ch, options2, token);
          }
        };
        const moreTriggerCharacter = options.moreTriggerCharacter || [];
        return [vscode_1.languages.registerOnTypeFormattingEditProvider(options.documentSelector, provider, options.firstTriggerCharacter, ...moreTriggerCharacter), provider];
      }
    };
    var RenameFeature = class extends TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.RenameRequest.type);
      }
      fillClientCapabilities(capabilities) {
        let rename = ensure(ensure(capabilities, "textDocument"), "rename");
        rename.dynamicRegistration = true;
        rename.prepareSupport = true;
        rename.prepareSupportDefaultBehavior = vscode_languageserver_protocol_1.PrepareSupportDefaultBehavior.Identifier;
        rename.honorsChangeAnnotations = true;
      }
      initialize(capabilities, documentSelector) {
        const options = this.getRegistrationOptions(documentSelector, capabilities.renameProvider);
        if (!options) {
          return;
        }
        if (Is.boolean(capabilities.renameProvider)) {
          options.prepareProvider = false;
        }
        this.register({ id: UUID.generateUuid(), registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideRenameEdits: (document, position, newName, token) => {
            const client2 = this._client;
            const provideRenameEdits = (document2, position2, newName2, token2) => {
              let params = {
                textDocument: client2.code2ProtocolConverter.asTextDocumentIdentifier(document2),
                position: client2.code2ProtocolConverter.asPosition(position2),
                newName: newName2
              };
              return client2.sendRequest(vscode_languageserver_protocol_1.RenameRequest.type, params, token2).then(client2.protocol2CodeConverter.asWorkspaceEdit, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.RenameRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideRenameEdits ? middleware.provideRenameEdits(document, position, newName, token, provideRenameEdits) : provideRenameEdits(document, position, newName, token);
          },
          prepareRename: options.prepareProvider ? (document, position, token) => {
            const client2 = this._client;
            const prepareRename = (document2, position2, token2) => {
              let params = {
                textDocument: client2.code2ProtocolConverter.asTextDocumentIdentifier(document2),
                position: client2.code2ProtocolConverter.asPosition(position2)
              };
              return client2.sendRequest(vscode_languageserver_protocol_1.PrepareRenameRequest.type, params, token2).then((result) => {
                if (vscode_languageserver_protocol_1.Range.is(result)) {
                  return client2.protocol2CodeConverter.asRange(result);
                } else if (this.isDefaultBehavior(result)) {
                  return result.defaultBehavior === true ? null : Promise.reject(new Error(`The element can't be renamed.`));
                } else if (result && vscode_languageserver_protocol_1.Range.is(result.range)) {
                  return {
                    range: client2.protocol2CodeConverter.asRange(result.range),
                    placeholder: result.placeholder
                  };
                }
                return Promise.reject(new Error(`The element can't be renamed.`));
              }, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.PrepareRenameRequest.type, error, void 0);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.prepareRename ? middleware.prepareRename(document, position, token, prepareRename) : prepareRename(document, position, token);
          } : void 0
        };
        return [vscode_1.languages.registerRenameProvider(options.documentSelector, provider), provider];
      }
      isDefaultBehavior(value) {
        const candidate = value;
        return candidate && Is.boolean(candidate.defaultBehavior);
      }
    };
    var DocumentLinkFeature = class extends TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.DocumentLinkRequest.type);
      }
      fillClientCapabilities(capabilities) {
        const documentLinkCapabilities = ensure(ensure(capabilities, "textDocument"), "documentLink");
        documentLinkCapabilities.dynamicRegistration = true;
        documentLinkCapabilities.tooltipSupport = true;
      }
      initialize(capabilities, documentSelector) {
        const options = this.getRegistrationOptions(documentSelector, capabilities.documentLinkProvider);
        if (!options) {
          return;
        }
        this.register({ id: UUID.generateUuid(), registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideDocumentLinks: (document, token) => {
            const client2 = this._client;
            const provideDocumentLinks = (document2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.DocumentLinkRequest.type, client2.code2ProtocolConverter.asDocumentLinkParams(document2), token2).then(client2.protocol2CodeConverter.asDocumentLinks, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.DocumentLinkRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideDocumentLinks ? middleware.provideDocumentLinks(document, token, provideDocumentLinks) : provideDocumentLinks(document, token);
          },
          resolveDocumentLink: options.resolveProvider ? (link, token) => {
            const client2 = this._client;
            let resolveDocumentLink = (link2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.DocumentLinkResolveRequest.type, client2.code2ProtocolConverter.asDocumentLink(link2), token2).then(client2.protocol2CodeConverter.asDocumentLink, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.DocumentLinkResolveRequest.type, error, link2);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.resolveDocumentLink ? middleware.resolveDocumentLink(link, token, resolveDocumentLink) : resolveDocumentLink(link, token);
          } : void 0
        };
        return [vscode_1.languages.registerDocumentLinkProvider(options.documentSelector, provider), provider];
      }
    };
    var ConfigurationFeature = class {
      constructor(_client) {
        this._client = _client;
        this._listeners = /* @__PURE__ */ new Map();
      }
      get registrationType() {
        return vscode_languageserver_protocol_1.DidChangeConfigurationNotification.type;
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "workspace"), "didChangeConfiguration").dynamicRegistration = true;
      }
      initialize() {
        let section = this._client.clientOptions.synchronize.configurationSection;
        if (section !== void 0) {
          this.register({
            id: UUID.generateUuid(),
            registerOptions: {
              section
            }
          });
        }
      }
      register(data) {
        let disposable = vscode_1.workspace.onDidChangeConfiguration((event) => {
          this.onDidChangeConfiguration(data.registerOptions.section, event);
        });
        this._listeners.set(data.id, disposable);
        if (data.registerOptions.section !== void 0) {
          this.onDidChangeConfiguration(data.registerOptions.section, void 0);
        }
      }
      unregister(id) {
        let disposable = this._listeners.get(id);
        if (disposable) {
          this._listeners.delete(id);
          disposable.dispose();
        }
      }
      dispose() {
        for (let disposable of this._listeners.values()) {
          disposable.dispose();
        }
        this._listeners.clear();
      }
      onDidChangeConfiguration(configurationSection, event) {
        let sections;
        if (Is.string(configurationSection)) {
          sections = [configurationSection];
        } else {
          sections = configurationSection;
        }
        if (sections !== void 0 && event !== void 0) {
          let affected = sections.some((section) => event.affectsConfiguration(section));
          if (!affected) {
            return;
          }
        }
        let didChangeConfiguration = (sections2) => {
          if (sections2 === void 0) {
            this._client.sendNotification(vscode_languageserver_protocol_1.DidChangeConfigurationNotification.type, { settings: null });
            return;
          }
          this._client.sendNotification(vscode_languageserver_protocol_1.DidChangeConfigurationNotification.type, { settings: this.extractSettingsInformation(sections2) });
        };
        let middleware = this.getMiddleware();
        middleware ? middleware(sections, didChangeConfiguration) : didChangeConfiguration(sections);
      }
      extractSettingsInformation(keys) {
        function ensurePath(config, path2) {
          let current = config;
          for (let i = 0; i < path2.length - 1; i++) {
            let obj = current[path2[i]];
            if (!obj) {
              obj = /* @__PURE__ */ Object.create(null);
              current[path2[i]] = obj;
            }
            current = obj;
          }
          return current;
        }
        let resource = this._client.clientOptions.workspaceFolder ? this._client.clientOptions.workspaceFolder.uri : void 0;
        let result = /* @__PURE__ */ Object.create(null);
        for (let i = 0; i < keys.length; i++) {
          let key = keys[i];
          let index = key.indexOf(".");
          let config = null;
          if (index >= 0) {
            config = vscode_1.workspace.getConfiguration(key.substr(0, index), resource).get(key.substr(index + 1));
          } else {
            config = vscode_1.workspace.getConfiguration(void 0, resource).get(key);
          }
          if (config) {
            let path2 = keys[i].split(".");
            ensurePath(result, path2)[path2[path2.length - 1]] = configuration_1.toJSONObject(config);
          }
        }
        return result;
      }
      getMiddleware() {
        let middleware = this._client.clientOptions.middleware;
        if (middleware.workspace && middleware.workspace.didChangeConfiguration) {
          return middleware.workspace.didChangeConfiguration;
        } else {
          return void 0;
        }
      }
    };
    var ExecuteCommandFeature = class {
      constructor(_client) {
        this._client = _client;
        this._commands = /* @__PURE__ */ new Map();
      }
      get registrationType() {
        return vscode_languageserver_protocol_1.ExecuteCommandRequest.type;
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "workspace"), "executeCommand").dynamicRegistration = true;
      }
      initialize(capabilities) {
        if (!capabilities.executeCommandProvider) {
          return;
        }
        this.register({
          id: UUID.generateUuid(),
          registerOptions: Object.assign({}, capabilities.executeCommandProvider)
        });
      }
      register(data) {
        const client2 = this._client;
        const middleware = client2.clientOptions.middleware;
        const executeCommand = (command, args) => {
          let params = {
            command,
            arguments: args
          };
          return client2.sendRequest(vscode_languageserver_protocol_1.ExecuteCommandRequest.type, params).then(void 0, (error) => {
            return client2.handleFailedRequest(vscode_languageserver_protocol_1.ExecuteCommandRequest.type, error, void 0);
          });
        };
        if (data.registerOptions.commands) {
          const disposables = [];
          for (const command of data.registerOptions.commands) {
            disposables.push(vscode_1.commands.registerCommand(command, (...args) => {
              return middleware.executeCommand ? middleware.executeCommand(command, args, executeCommand) : executeCommand(command, args);
            }));
          }
          this._commands.set(data.id, disposables);
        }
      }
      unregister(id) {
        let disposables = this._commands.get(id);
        if (disposables) {
          disposables.forEach((disposable) => disposable.dispose());
        }
      }
      dispose() {
        this._commands.forEach((value) => {
          value.forEach((disposable) => disposable.dispose());
        });
        this._commands.clear();
      }
    };
    var MessageTransports;
    (function(MessageTransports2) {
      function is(value) {
        let candidate = value;
        return candidate && vscode_languageserver_protocol_1.MessageReader.is(value.reader) && vscode_languageserver_protocol_1.MessageWriter.is(value.writer);
      }
      MessageTransports2.is = is;
    })(MessageTransports = exports2.MessageTransports || (exports2.MessageTransports = {}));
    var OnReady = class {
      constructor(_resolve, _reject) {
        this._resolve = _resolve;
        this._reject = _reject;
        this._used = false;
      }
      get isUsed() {
        return this._used;
      }
      resolve() {
        this._used = true;
        this._resolve();
      }
      reject(error) {
        this._used = true;
        this._reject(error);
      }
    };
    var BaseLanguageClient = class _BaseLanguageClient {
      constructor(id, name, clientOptions) {
        var _a;
        this._traceFormat = vscode_languageserver_protocol_1.TraceFormat.Text;
        this._features = [];
        this._dynamicFeatures = /* @__PURE__ */ new Map();
        this._id = id;
        this._name = name;
        clientOptions = clientOptions || {};
        const markdown = { isTrusted: false };
        if (clientOptions.markdown !== void 0 && clientOptions.markdown.isTrusted === true) {
          markdown.isTrusted = true;
        }
        this._clientOptions = {
          documentSelector: clientOptions.documentSelector || [],
          synchronize: clientOptions.synchronize || {},
          diagnosticCollectionName: clientOptions.diagnosticCollectionName,
          outputChannelName: clientOptions.outputChannelName || this._name,
          revealOutputChannelOn: clientOptions.revealOutputChannelOn || RevealOutputChannelOn.Error,
          stdioEncoding: clientOptions.stdioEncoding || "utf8",
          initializationOptions: clientOptions.initializationOptions,
          initializationFailedHandler: clientOptions.initializationFailedHandler,
          progressOnInitialization: !!clientOptions.progressOnInitialization,
          errorHandler: clientOptions.errorHandler || this.createDefaultErrorHandler((_a = clientOptions.connectionOptions) === null || _a === void 0 ? void 0 : _a.maxRestartCount),
          middleware: clientOptions.middleware || {},
          uriConverters: clientOptions.uriConverters,
          workspaceFolder: clientOptions.workspaceFolder,
          connectionOptions: clientOptions.connectionOptions,
          markdown
        };
        this._clientOptions.synchronize = this._clientOptions.synchronize || {};
        this._state = ClientState.Initial;
        this._connectionPromise = void 0;
        this._resolvedConnection = void 0;
        this._initializeResult = void 0;
        if (clientOptions.outputChannel) {
          this._outputChannel = clientOptions.outputChannel;
          this._disposeOutputChannel = false;
        } else {
          this._outputChannel = void 0;
          this._disposeOutputChannel = true;
        }
        this._traceOutputChannel = clientOptions.traceOutputChannel;
        this._listeners = void 0;
        this._providers = void 0;
        this._diagnostics = void 0;
        this._fileEvents = [];
        this._fileEventDelayer = new async_1.Delayer(250);
        this._onReady = new Promise((resolve, reject) => {
          this._onReadyCallbacks = new OnReady(resolve, reject);
        });
        this._onStop = void 0;
        this._telemetryEmitter = new vscode_languageserver_protocol_1.Emitter();
        this._stateChangeEmitter = new vscode_languageserver_protocol_1.Emitter();
        this._trace = vscode_languageserver_protocol_1.Trace.Off;
        this._tracer = {
          log: (messageOrDataObject, data) => {
            if (Is.string(messageOrDataObject)) {
              this.logTrace(messageOrDataObject, data);
            } else {
              this.logObjectTrace(messageOrDataObject);
            }
          }
        };
        this._c2p = c2p.createConverter(clientOptions.uriConverters ? clientOptions.uriConverters.code2Protocol : void 0);
        this._p2c = p2c.createConverter(clientOptions.uriConverters ? clientOptions.uriConverters.protocol2Code : void 0, this._clientOptions.markdown.isTrusted);
        this._syncedDocuments = /* @__PURE__ */ new Map();
        this.registerBuiltinFeatures();
      }
      get state() {
        return this._state;
      }
      set state(value) {
        let oldState = this.getPublicState();
        this._state = value;
        let newState = this.getPublicState();
        if (newState !== oldState) {
          this._stateChangeEmitter.fire({ oldState, newState });
        }
      }
      getPublicState() {
        if (this.state === ClientState.Running) {
          return State.Running;
        } else if (this.state === ClientState.Starting) {
          return State.Starting;
        } else {
          return State.Stopped;
        }
      }
      get initializeResult() {
        return this._initializeResult;
      }
      sendRequest(type, ...params) {
        if (!this.isConnectionActive()) {
          throw new Error("Language client is not ready yet");
        }
        this.forceDocumentSync();
        try {
          return this._resolvedConnection.sendRequest(type, ...params);
        } catch (error) {
          this.error(`Sending request ${Is.string(type) ? type : type.method} failed.`, error);
          throw error;
        }
      }
      onRequest(type, handler) {
        if (!this.isConnectionActive()) {
          throw new Error("Language client is not ready yet");
        }
        try {
          return this._resolvedConnection.onRequest(type, handler);
        } catch (error) {
          this.error(`Registering request handler ${Is.string(type) ? type : type.method} failed.`, error);
          throw error;
        }
      }
      sendNotification(type, params) {
        if (!this.isConnectionActive()) {
          throw new Error("Language client is not ready yet");
        }
        this.forceDocumentSync();
        try {
          this._resolvedConnection.sendNotification(type, params);
        } catch (error) {
          this.error(`Sending notification ${Is.string(type) ? type : type.method} failed.`, error);
          throw error;
        }
      }
      onNotification(type, handler) {
        if (!this.isConnectionActive()) {
          throw new Error("Language client is not ready yet");
        }
        try {
          return this._resolvedConnection.onNotification(type, handler);
        } catch (error) {
          this.error(`Registering notification handler ${Is.string(type) ? type : type.method} failed.`, error);
          throw error;
        }
      }
      onProgress(type, token, handler) {
        if (!this.isConnectionActive()) {
          throw new Error("Language client is not ready yet");
        }
        try {
          if (vscode_languageserver_protocol_1.WorkDoneProgress.is(type)) {
            const handleWorkDoneProgress = this._clientOptions.middleware.handleWorkDoneProgress;
            if (handleWorkDoneProgress !== void 0) {
              return this._resolvedConnection.onProgress(type, token, (params) => {
                handleWorkDoneProgress(token, params, () => handler(params));
              });
            }
          }
          return this._resolvedConnection.onProgress(type, token, handler);
        } catch (error) {
          this.error(`Registering progress handler for token ${token} failed.`, error);
          throw error;
        }
      }
      sendProgress(type, token, value) {
        if (!this.isConnectionActive()) {
          throw new Error("Language client is not ready yet");
        }
        this.forceDocumentSync();
        try {
          this._resolvedConnection.sendProgress(type, token, value);
        } catch (error) {
          this.error(`Sending progress for token ${token} failed.`, error);
          throw error;
        }
      }
      get clientOptions() {
        return this._clientOptions;
      }
      get protocol2CodeConverter() {
        return this._p2c;
      }
      get code2ProtocolConverter() {
        return this._c2p;
      }
      get onTelemetry() {
        return this._telemetryEmitter.event;
      }
      get onDidChangeState() {
        return this._stateChangeEmitter.event;
      }
      get outputChannel() {
        if (!this._outputChannel) {
          this._outputChannel = vscode_1.window.createOutputChannel(this._clientOptions.outputChannelName ? this._clientOptions.outputChannelName : this._name);
        }
        return this._outputChannel;
      }
      get traceOutputChannel() {
        if (this._traceOutputChannel) {
          return this._traceOutputChannel;
        }
        return this.outputChannel;
      }
      get diagnostics() {
        return this._diagnostics;
      }
      createDefaultErrorHandler(maxRestartCount) {
        if (maxRestartCount !== void 0 && maxRestartCount < 0) {
          throw new Error(`Invalid maxRestartCount: ${maxRestartCount}`);
        }
        return new DefaultErrorHandler(this._name, maxRestartCount !== null && maxRestartCount !== void 0 ? maxRestartCount : 4);
      }
      set trace(value) {
        this._trace = value;
        this.onReady().then(() => {
          this.resolveConnection().then((connection) => {
            connection.trace(this._trace, this._tracer, {
              sendNotification: false,
              traceFormat: this._traceFormat
            });
          });
        }, () => {
        });
      }
      data2String(data) {
        if (data instanceof vscode_languageserver_protocol_1.ResponseError) {
          const responseError = data;
          return `  Message: ${responseError.message}
  Code: ${responseError.code} ${responseError.data ? "\n" + responseError.data.toString() : ""}`;
        }
        if (data instanceof Error) {
          if (Is.string(data.stack)) {
            return data.stack;
          }
          return data.message;
        }
        if (Is.string(data)) {
          return data;
        }
        return data.toString();
      }
      info(message, data, showNotification = true) {
        this.outputChannel.appendLine(`[Info  - ${(/* @__PURE__ */ new Date()).toLocaleTimeString()}] ${message}`);
        if (data) {
          this.outputChannel.appendLine(this.data2String(data));
        }
        if (showNotification && this._clientOptions.revealOutputChannelOn <= RevealOutputChannelOn.Info) {
          this.showNotificationMessage();
        }
      }
      warn(message, data, showNotification = true) {
        this.outputChannel.appendLine(`[Warn  - ${(/* @__PURE__ */ new Date()).toLocaleTimeString()}] ${message}`);
        if (data) {
          this.outputChannel.appendLine(this.data2String(data));
        }
        if (showNotification && this._clientOptions.revealOutputChannelOn <= RevealOutputChannelOn.Warn) {
          this.showNotificationMessage();
        }
      }
      error(message, data, showNotification = true) {
        this.outputChannel.appendLine(`[Error - ${(/* @__PURE__ */ new Date()).toLocaleTimeString()}] ${message}`);
        if (data) {
          this.outputChannel.appendLine(this.data2String(data));
        }
        if (showNotification && this._clientOptions.revealOutputChannelOn <= RevealOutputChannelOn.Error) {
          this.showNotificationMessage();
        }
      }
      showNotificationMessage() {
        vscode_1.window.showInformationMessage("A request has failed. See the output for more information.", "Go to output").then(() => {
          this.outputChannel.show(true);
        });
      }
      logTrace(message, data) {
        this.traceOutputChannel.appendLine(`[Trace - ${(/* @__PURE__ */ new Date()).toLocaleTimeString()}] ${message}`);
        if (data) {
          this.traceOutputChannel.appendLine(this.data2String(data));
        }
      }
      logObjectTrace(data) {
        if (data.isLSPMessage && data.type) {
          this.traceOutputChannel.append(`[LSP   - ${(/* @__PURE__ */ new Date()).toLocaleTimeString()}] `);
        } else {
          this.traceOutputChannel.append(`[Trace - ${(/* @__PURE__ */ new Date()).toLocaleTimeString()}] `);
        }
        if (data) {
          this.traceOutputChannel.appendLine(`${JSON.stringify(data)}`);
        }
      }
      needsStart() {
        return this.state === ClientState.Initial || this.state === ClientState.Stopping || this.state === ClientState.Stopped;
      }
      needsStop() {
        return this.state === ClientState.Starting || this.state === ClientState.Running;
      }
      onReady() {
        return this._onReady;
      }
      isConnectionActive() {
        return this.state === ClientState.Running && !!this._resolvedConnection;
      }
      start() {
        if (this._onReadyCallbacks.isUsed) {
          this._onReady = new Promise((resolve, reject) => {
            this._onReadyCallbacks = new OnReady(resolve, reject);
          });
        }
        this._listeners = [];
        this._providers = [];
        if (!this._diagnostics) {
          this._diagnostics = this._clientOptions.diagnosticCollectionName ? vscode_1.languages.createDiagnosticCollection(this._clientOptions.diagnosticCollectionName) : vscode_1.languages.createDiagnosticCollection();
        }
        this.state = ClientState.Starting;
        this.resolveConnection().then((connection) => {
          connection.onLogMessage((message) => {
            switch (message.type) {
              case vscode_languageserver_protocol_1.MessageType.Error:
                this.error(message.message, void 0, false);
                break;
              case vscode_languageserver_protocol_1.MessageType.Warning:
                this.warn(message.message, void 0, false);
                break;
              case vscode_languageserver_protocol_1.MessageType.Info:
                this.info(message.message, void 0, false);
                break;
              default:
                this.outputChannel.appendLine(message.message);
            }
          });
          connection.onShowMessage((message) => {
            switch (message.type) {
              case vscode_languageserver_protocol_1.MessageType.Error:
                vscode_1.window.showErrorMessage(message.message);
                break;
              case vscode_languageserver_protocol_1.MessageType.Warning:
                vscode_1.window.showWarningMessage(message.message);
                break;
              case vscode_languageserver_protocol_1.MessageType.Info:
                vscode_1.window.showInformationMessage(message.message);
                break;
              default:
                vscode_1.window.showInformationMessage(message.message);
            }
          });
          connection.onRequest(vscode_languageserver_protocol_1.ShowMessageRequest.type, (params) => {
            let messageFunc;
            switch (params.type) {
              case vscode_languageserver_protocol_1.MessageType.Error:
                messageFunc = vscode_1.window.showErrorMessage;
                break;
              case vscode_languageserver_protocol_1.MessageType.Warning:
                messageFunc = vscode_1.window.showWarningMessage;
                break;
              case vscode_languageserver_protocol_1.MessageType.Info:
                messageFunc = vscode_1.window.showInformationMessage;
                break;
              default:
                messageFunc = vscode_1.window.showInformationMessage;
            }
            let actions = params.actions || [];
            return messageFunc(params.message, ...actions);
          });
          connection.onTelemetry((data) => {
            this._telemetryEmitter.fire(data);
          });
          connection.onRequest(vscode_languageserver_protocol_1.ShowDocumentRequest.type, async (params) => {
            var _a;
            const showDocument = async (params2) => {
              const uri = this.protocol2CodeConverter.asUri(params2.uri);
              try {
                if (params2.external === true) {
                  const success = await vscode_1.env.openExternal(uri);
                  return { success };
                } else {
                  const options = {};
                  if (params2.selection !== void 0) {
                    options.selection = this.protocol2CodeConverter.asRange(params2.selection);
                  }
                  if (params2.takeFocus === void 0 || params2.takeFocus === false) {
                    options.preserveFocus = true;
                  } else if (params2.takeFocus === true) {
                    options.preserveFocus = false;
                  }
                  await vscode_1.window.showTextDocument(uri, options);
                  return { success: true };
                }
              } catch (error) {
                return { success: true };
              }
            };
            const middleware = (_a = this._clientOptions.middleware.window) === null || _a === void 0 ? void 0 : _a.showDocument;
            if (middleware !== void 0) {
              return middleware(params, showDocument);
            } else {
              return showDocument(params);
            }
          });
          connection.listen();
          return this.initialize(connection);
        }).then(void 0, (error) => {
          this.state = ClientState.StartFailed;
          this._onReadyCallbacks.reject(error);
          this.error("Starting client failed", error);
          vscode_1.window.showErrorMessage(`Couldn't start client ${this._name}`);
        });
        return new vscode_1.Disposable(() => {
          if (this.needsStop()) {
            this.stop();
          }
        });
      }
      resolveConnection() {
        if (!this._connectionPromise) {
          this._connectionPromise = this.createConnection();
        }
        return this._connectionPromise;
      }
      initialize(connection) {
        this.refreshTrace(connection, false);
        let initOption = this._clientOptions.initializationOptions;
        let rootPath = this._clientOptions.workspaceFolder ? this._clientOptions.workspaceFolder.uri.fsPath : this._clientGetRootPath();
        let initParams = {
          processId: null,
          clientInfo: {
            name: vscode_1.env.appName,
            version: vscode_1.version
          },
          locale: this.getLocale(),
          rootPath: rootPath ? rootPath : null,
          rootUri: rootPath ? this._c2p.asUri(vscode_1.Uri.file(rootPath)) : null,
          capabilities: this.computeClientCapabilities(),
          initializationOptions: Is.func(initOption) ? initOption() : initOption,
          trace: vscode_languageserver_protocol_1.Trace.toString(this._trace),
          workspaceFolders: null
        };
        this.fillInitializeParams(initParams);
        if (this._clientOptions.progressOnInitialization) {
          const token = UUID.generateUuid();
          const part = new progressPart_1.ProgressPart(connection, token);
          initParams.workDoneToken = token;
          return this.doInitialize(connection, initParams).then((result) => {
            part.done();
            return result;
          }, (error) => {
            part.cancel();
            throw error;
          });
        } else {
          return this.doInitialize(connection, initParams);
        }
      }
      doInitialize(connection, initParams) {
        return connection.initialize(initParams).then((result) => {
          this._resolvedConnection = connection;
          this._initializeResult = result;
          this.state = ClientState.Running;
          let textDocumentSyncOptions = void 0;
          if (Is.number(result.capabilities.textDocumentSync)) {
            if (result.capabilities.textDocumentSync === vscode_languageserver_protocol_1.TextDocumentSyncKind.None) {
              textDocumentSyncOptions = {
                openClose: false,
                change: vscode_languageserver_protocol_1.TextDocumentSyncKind.None,
                save: void 0
              };
            } else {
              textDocumentSyncOptions = {
                openClose: true,
                change: result.capabilities.textDocumentSync,
                save: {
                  includeText: false
                }
              };
            }
          } else if (result.capabilities.textDocumentSync !== void 0 && result.capabilities.textDocumentSync !== null) {
            textDocumentSyncOptions = result.capabilities.textDocumentSync;
          }
          this._capabilities = Object.assign({}, result.capabilities, { resolvedTextDocumentSync: textDocumentSyncOptions });
          connection.onDiagnostics((params) => this.handleDiagnostics(params));
          connection.onRequest(vscode_languageserver_protocol_1.RegistrationRequest.type, (params) => this.handleRegistrationRequest(params));
          connection.onRequest("client/registerFeature", (params) => this.handleRegistrationRequest(params));
          connection.onRequest(vscode_languageserver_protocol_1.UnregistrationRequest.type, (params) => this.handleUnregistrationRequest(params));
          connection.onRequest("client/unregisterFeature", (params) => this.handleUnregistrationRequest(params));
          connection.onRequest(vscode_languageserver_protocol_1.ApplyWorkspaceEditRequest.type, (params) => this.handleApplyWorkspaceEdit(params));
          connection.sendNotification(vscode_languageserver_protocol_1.InitializedNotification.type, {});
          this.hookFileEvents(connection);
          this.hookConfigurationChanged(connection);
          this.initializeFeatures(connection);
          this._onReadyCallbacks.resolve();
          return result;
        }).then(void 0, (error) => {
          if (this._clientOptions.initializationFailedHandler) {
            if (this._clientOptions.initializationFailedHandler(error)) {
              this.initialize(connection);
            } else {
              this.stop();
              this._onReadyCallbacks.reject(error);
            }
          } else if (error instanceof vscode_languageserver_protocol_1.ResponseError && error.data && error.data.retry) {
            vscode_1.window.showErrorMessage(error.message, { title: "Retry", id: "retry" }).then((item) => {
              if (item && item.id === "retry") {
                this.initialize(connection);
              } else {
                this.stop();
                this._onReadyCallbacks.reject(error);
              }
            });
          } else {
            if (error && error.message) {
              vscode_1.window.showErrorMessage(error.message);
            }
            this.error("Server initialization failed.", error);
            this.stop();
            this._onReadyCallbacks.reject(error);
          }
          throw error;
        });
      }
      _clientGetRootPath() {
        let folders = vscode_1.workspace.workspaceFolders;
        if (!folders || folders.length === 0) {
          return void 0;
        }
        let folder = folders[0];
        if (folder.uri.scheme === "file") {
          return folder.uri.fsPath;
        }
        return void 0;
      }
      stop() {
        this._initializeResult = void 0;
        if (!this._connectionPromise) {
          this.state = ClientState.Stopped;
          return Promise.resolve();
        }
        if (this.state === ClientState.Stopping && this._onStop) {
          return this._onStop;
        }
        this.state = ClientState.Stopping;
        this.cleanUp(false);
        return this._onStop = this.resolveConnection().then((connection) => {
          return connection.shutdown().then(() => {
            connection.exit();
            connection.end();
            connection.dispose();
            this.state = ClientState.Stopped;
            this.cleanUpChannel();
            this._onStop = void 0;
            this._connectionPromise = void 0;
            this._resolvedConnection = void 0;
          });
        });
      }
      cleanUp(channel = true, diagnostics = true) {
        if (this._listeners) {
          this._listeners.forEach((listener) => listener.dispose());
          this._listeners = void 0;
        }
        if (this._providers) {
          this._providers.forEach((provider) => provider.dispose());
          this._providers = void 0;
        }
        if (this._syncedDocuments) {
          this._syncedDocuments.clear();
        }
        for (const feature of this._features.values()) {
          feature.dispose();
        }
        if (channel) {
          this.cleanUpChannel();
        }
        if (diagnostics && this._diagnostics) {
          this._diagnostics.dispose();
          this._diagnostics = void 0;
        }
      }
      cleanUpChannel() {
        if (this._outputChannel && this._disposeOutputChannel) {
          this._outputChannel.dispose();
          this._outputChannel = void 0;
        }
      }
      notifyFileEvent(event) {
        var _a;
        const client2 = this;
        function didChangeWatchedFile(event2) {
          client2._fileEvents.push(event2);
          client2._fileEventDelayer.trigger(() => {
            client2.onReady().then(() => {
              client2.resolveConnection().then((connection) => {
                if (client2.isConnectionActive()) {
                  client2.forceDocumentSync();
                  connection.didChangeWatchedFiles({ changes: client2._fileEvents });
                }
                client2._fileEvents = [];
              });
            }, (error) => {
              client2.error(`Notify file events failed.`, error);
            });
          });
        }
        const workSpaceMiddleware = (_a = this.clientOptions.middleware) === null || _a === void 0 ? void 0 : _a.workspace;
        (workSpaceMiddleware === null || workSpaceMiddleware === void 0 ? void 0 : workSpaceMiddleware.didChangeWatchedFile) ? workSpaceMiddleware.didChangeWatchedFile(event, didChangeWatchedFile) : didChangeWatchedFile(event);
      }
      forceDocumentSync() {
        if (this._didChangeTextDocumentFeature === void 0) {
          this._didChangeTextDocumentFeature = this._dynamicFeatures.get(vscode_languageserver_protocol_1.DidChangeTextDocumentNotification.type.method);
        }
        this._didChangeTextDocumentFeature.forceDelivery();
      }
      handleDiagnostics(params) {
        if (!this._diagnostics) {
          return;
        }
        let uri = this._p2c.asUri(params.uri);
        let diagnostics = this._p2c.asDiagnostics(params.diagnostics);
        let middleware = this.clientOptions.middleware;
        if (middleware.handleDiagnostics) {
          middleware.handleDiagnostics(uri, diagnostics, (uri2, diagnostics2) => this.setDiagnostics(uri2, diagnostics2));
        } else {
          this.setDiagnostics(uri, diagnostics);
        }
      }
      setDiagnostics(uri, diagnostics) {
        if (!this._diagnostics) {
          return;
        }
        this._diagnostics.set(uri, diagnostics);
      }
      createConnection() {
        let errorHandler = (error, message, count) => {
          this.handleConnectionError(error, message, count);
        };
        let closeHandler = () => {
          this.handleConnectionClosed();
        };
        return this.createMessageTransports(this._clientOptions.stdioEncoding || "utf8").then((transports) => {
          return createConnection(transports.reader, transports.writer, errorHandler, closeHandler, this._clientOptions.connectionOptions);
        });
      }
      handleConnectionClosed() {
        if (this.state === ClientState.Stopping || this.state === ClientState.Stopped) {
          return;
        }
        try {
          if (this._resolvedConnection) {
            this._resolvedConnection.dispose();
          }
        } catch (error) {
        }
        let action = CloseAction.DoNotRestart;
        try {
          action = this._clientOptions.errorHandler.closed();
        } catch (error) {
        }
        this._connectionPromise = void 0;
        this._resolvedConnection = void 0;
        if (action === CloseAction.DoNotRestart) {
          this.error("Connection to server got closed. Server will not be restarted.");
          if (this.state === ClientState.Starting) {
            this._onReadyCallbacks.reject(new Error(`Connection to server got closed. Server will not be restarted.`));
            this.state = ClientState.StartFailed;
          } else {
            this.state = ClientState.Stopped;
          }
          this.cleanUp(false, true);
        } else if (action === CloseAction.Restart) {
          this.info("Connection to server got closed. Server will restart.");
          this.cleanUp(false, false);
          this.state = ClientState.Initial;
          this.start();
        }
      }
      handleConnectionError(error, message, count) {
        let action = this._clientOptions.errorHandler.error(error, message, count);
        if (action === ErrorAction.Shutdown) {
          this.error("Connection to server is erroring. Shutting down server.");
          this.stop();
        }
      }
      hookConfigurationChanged(connection) {
        vscode_1.workspace.onDidChangeConfiguration(() => {
          this.refreshTrace(connection, true);
        });
      }
      refreshTrace(connection, sendNotification = false) {
        let config = vscode_1.workspace.getConfiguration(this._id);
        let trace = vscode_languageserver_protocol_1.Trace.Off;
        let traceFormat = vscode_languageserver_protocol_1.TraceFormat.Text;
        if (config) {
          const traceConfig = config.get("trace.server", "off");
          if (typeof traceConfig === "string") {
            trace = vscode_languageserver_protocol_1.Trace.fromString(traceConfig);
          } else {
            trace = vscode_languageserver_protocol_1.Trace.fromString(config.get("trace.server.verbosity", "off"));
            traceFormat = vscode_languageserver_protocol_1.TraceFormat.fromString(config.get("trace.server.format", "text"));
          }
        }
        this._trace = trace;
        this._traceFormat = traceFormat;
        connection.trace(this._trace, this._tracer, {
          sendNotification,
          traceFormat: this._traceFormat
        });
      }
      hookFileEvents(_connection) {
        let fileEvents = this._clientOptions.synchronize.fileEvents;
        if (!fileEvents) {
          return;
        }
        let watchers;
        if (Is.array(fileEvents)) {
          watchers = fileEvents;
        } else {
          watchers = [fileEvents];
        }
        if (!watchers) {
          return;
        }
        this._dynamicFeatures.get(vscode_languageserver_protocol_1.DidChangeWatchedFilesNotification.type.method).registerRaw(UUID.generateUuid(), watchers);
      }
      registerFeatures(features) {
        for (let feature of features) {
          this.registerFeature(feature);
        }
      }
      registerFeature(feature) {
        this._features.push(feature);
        if (DynamicFeature.is(feature)) {
          const registrationType = feature.registrationType;
          this._dynamicFeatures.set(registrationType.method, feature);
        }
      }
      getFeature(request) {
        return this._dynamicFeatures.get(request);
      }
      registerBuiltinFeatures() {
        this.registerFeature(new ConfigurationFeature(this));
        this.registerFeature(new DidOpenTextDocumentFeature(this, this._syncedDocuments));
        this.registerFeature(new DidChangeTextDocumentFeature(this));
        this.registerFeature(new WillSaveFeature(this));
        this.registerFeature(new WillSaveWaitUntilFeature(this));
        this.registerFeature(new DidSaveTextDocumentFeature(this));
        this.registerFeature(new DidCloseTextDocumentFeature(this, this._syncedDocuments));
        this.registerFeature(new FileSystemWatcherFeature(this, (event) => this.notifyFileEvent(event)));
        this.registerFeature(new CompletionItemFeature(this));
        this.registerFeature(new HoverFeature(this));
        this.registerFeature(new SignatureHelpFeature(this));
        this.registerFeature(new DefinitionFeature(this));
        this.registerFeature(new ReferencesFeature(this));
        this.registerFeature(new DocumentHighlightFeature(this));
        this.registerFeature(new DocumentSymbolFeature(this));
        this.registerFeature(new WorkspaceSymbolFeature(this));
        this.registerFeature(new CodeActionFeature(this));
        this.registerFeature(new CodeLensFeature(this));
        this.registerFeature(new DocumentFormattingFeature(this));
        this.registerFeature(new DocumentRangeFormattingFeature(this));
        this.registerFeature(new DocumentOnTypeFormattingFeature(this));
        this.registerFeature(new RenameFeature(this));
        this.registerFeature(new DocumentLinkFeature(this));
        this.registerFeature(new ExecuteCommandFeature(this));
      }
      fillInitializeParams(params) {
        for (let feature of this._features) {
          if (Is.func(feature.fillInitializeParams)) {
            feature.fillInitializeParams(params);
          }
        }
      }
      computeClientCapabilities() {
        const result = {};
        ensure(result, "workspace").applyEdit = true;
        const workspaceEdit = ensure(ensure(result, "workspace"), "workspaceEdit");
        workspaceEdit.documentChanges = true;
        workspaceEdit.resourceOperations = [vscode_languageserver_protocol_1.ResourceOperationKind.Create, vscode_languageserver_protocol_1.ResourceOperationKind.Rename, vscode_languageserver_protocol_1.ResourceOperationKind.Delete];
        workspaceEdit.failureHandling = vscode_languageserver_protocol_1.FailureHandlingKind.TextOnlyTransactional;
        workspaceEdit.normalizesLineEndings = true;
        workspaceEdit.changeAnnotationSupport = {
          groupsOnLabel: true
        };
        const diagnostics = ensure(ensure(result, "textDocument"), "publishDiagnostics");
        diagnostics.relatedInformation = true;
        diagnostics.versionSupport = false;
        diagnostics.tagSupport = { valueSet: [vscode_languageserver_protocol_1.DiagnosticTag.Unnecessary, vscode_languageserver_protocol_1.DiagnosticTag.Deprecated] };
        diagnostics.codeDescriptionSupport = true;
        diagnostics.dataSupport = true;
        const windowCapabilities = ensure(result, "window");
        const showMessage = ensure(windowCapabilities, "showMessage");
        showMessage.messageActionItem = { additionalPropertiesSupport: true };
        const showDocument = ensure(windowCapabilities, "showDocument");
        showDocument.support = true;
        const generalCapabilities = ensure(result, "general");
        generalCapabilities.regularExpressions = { engine: "ECMAScript", version: "ES2020" };
        generalCapabilities.markdown = { parser: "marked", version: "1.1.0" };
        for (let feature of this._features) {
          feature.fillClientCapabilities(result);
        }
        return result;
      }
      initializeFeatures(_connection) {
        let documentSelector = this._clientOptions.documentSelector;
        for (let feature of this._features) {
          feature.initialize(this._capabilities, documentSelector);
        }
      }
      handleRegistrationRequest(params) {
        return new Promise((resolve, reject) => {
          for (const registration of params.registrations) {
            const feature = this._dynamicFeatures.get(registration.method);
            if (feature === void 0) {
              reject(new Error(`No feature implementation for ${registration.method} found. Registration failed.`));
              return;
            }
            const options = registration.registerOptions || {};
            options.documentSelector = options.documentSelector || this._clientOptions.documentSelector;
            const data = {
              id: registration.id,
              registerOptions: options
            };
            try {
              feature.register(data);
            } catch (err) {
              reject(err);
              return;
            }
          }
          resolve();
        });
      }
      handleUnregistrationRequest(params) {
        return new Promise((resolve, reject) => {
          for (let unregistration of params.unregisterations) {
            const feature = this._dynamicFeatures.get(unregistration.method);
            if (!feature) {
              reject(new Error(`No feature implementation for ${unregistration.method} found. Unregistration failed.`));
              return;
            }
            feature.unregister(unregistration.id);
          }
          resolve();
        });
      }
      handleApplyWorkspaceEdit(params) {
        let workspaceEdit = params.edit;
        let openTextDocuments = /* @__PURE__ */ new Map();
        vscode_1.workspace.textDocuments.forEach((document) => openTextDocuments.set(document.uri.toString(), document));
        let versionMismatch = false;
        if (workspaceEdit.documentChanges) {
          for (const change of workspaceEdit.documentChanges) {
            if (vscode_languageserver_protocol_1.TextDocumentEdit.is(change) && change.textDocument.version && change.textDocument.version >= 0) {
              let textDocument = openTextDocuments.get(change.textDocument.uri);
              if (textDocument && textDocument.version !== change.textDocument.version) {
                versionMismatch = true;
                break;
              }
            }
          }
        }
        if (versionMismatch) {
          return Promise.resolve({ applied: false });
        }
        return Is.asPromise(vscode_1.workspace.applyEdit(this._p2c.asWorkspaceEdit(params.edit)).then((value) => {
          return { applied: value };
        }));
      }
      handleFailedRequest(type, error, defaultValue) {
        if (error instanceof vscode_languageserver_protocol_1.ResponseError) {
          if (error.code === vscode_languageserver_protocol_1.LSPErrorCodes.RequestCancelled) {
            throw this.makeCancelError();
          } else if (error.code === vscode_languageserver_protocol_1.LSPErrorCodes.ContentModified) {
            return defaultValue;
          }
        }
        this.error(`Request ${type.method} failed.`, error);
        throw error;
      }
      makeCancelError() {
        const result = new Error(_BaseLanguageClient.Canceled);
        result.name = _BaseLanguageClient.Canceled;
        return result;
      }
    };
    exports2.BaseLanguageClient = BaseLanguageClient;
    BaseLanguageClient.Canceled = "Canceled";
  }
});

// node_modules/vscode-languageclient/lib/common/colorProvider.js
var require_colorProvider = __commonJS({
  "node_modules/vscode-languageclient/lib/common/colorProvider.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.ColorProviderFeature = void 0;
    var vscode_1 = require("vscode");
    var vscode_languageserver_protocol_1 = require_main3();
    var client_1 = require_client();
    function ensure(target, key) {
      if (target[key] === void 0) {
        target[key] = {};
      }
      return target[key];
    }
    var ColorProviderFeature = class extends client_1.TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.DocumentColorRequest.type);
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "textDocument"), "colorProvider").dynamicRegistration = true;
      }
      initialize(capabilities, documentSelector) {
        let [id, options] = this.getRegistration(documentSelector, capabilities.colorProvider);
        if (!id || !options) {
          return;
        }
        this.register({ id, registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideColorPresentations: (color, context, token) => {
            const client2 = this._client;
            const provideColorPresentations = (color2, context2, token2) => {
              const requestParams = {
                color: color2,
                textDocument: client2.code2ProtocolConverter.asTextDocumentIdentifier(context2.document),
                range: client2.code2ProtocolConverter.asRange(context2.range)
              };
              return client2.sendRequest(vscode_languageserver_protocol_1.ColorPresentationRequest.type, requestParams, token2).then(this.asColorPresentations.bind(this), (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.ColorPresentationRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideColorPresentations ? middleware.provideColorPresentations(color, context, token, provideColorPresentations) : provideColorPresentations(color, context, token);
          },
          provideDocumentColors: (document, token) => {
            const client2 = this._client;
            const provideDocumentColors = (document2, token2) => {
              const requestParams = {
                textDocument: client2.code2ProtocolConverter.asTextDocumentIdentifier(document2)
              };
              return client2.sendRequest(vscode_languageserver_protocol_1.DocumentColorRequest.type, requestParams, token2).then(this.asColorInformations.bind(this), (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.ColorPresentationRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideDocumentColors ? middleware.provideDocumentColors(document, token, provideDocumentColors) : provideDocumentColors(document, token);
          }
        };
        return [vscode_1.languages.registerColorProvider(options.documentSelector, provider), provider];
      }
      asColor(color) {
        return new vscode_1.Color(color.red, color.green, color.blue, color.alpha);
      }
      asColorInformations(colorInformation) {
        if (Array.isArray(colorInformation)) {
          return colorInformation.map((ci) => {
            return new vscode_1.ColorInformation(this._client.protocol2CodeConverter.asRange(ci.range), this.asColor(ci.color));
          });
        }
        return [];
      }
      asColorPresentations(colorPresentations) {
        if (Array.isArray(colorPresentations)) {
          return colorPresentations.map((cp) => {
            let presentation = new vscode_1.ColorPresentation(cp.label);
            presentation.additionalTextEdits = this._client.protocol2CodeConverter.asTextEdits(cp.additionalTextEdits);
            presentation.textEdit = this._client.protocol2CodeConverter.asTextEdit(cp.textEdit);
            return presentation;
          });
        }
        return [];
      }
    };
    exports2.ColorProviderFeature = ColorProviderFeature;
  }
});

// node_modules/vscode-languageclient/lib/common/implementation.js
var require_implementation = __commonJS({
  "node_modules/vscode-languageclient/lib/common/implementation.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.ImplementationFeature = void 0;
    var vscode_1 = require("vscode");
    var vscode_languageserver_protocol_1 = require_main3();
    var client_1 = require_client();
    function ensure(target, key) {
      if (target[key] === void 0) {
        target[key] = {};
      }
      return target[key];
    }
    var ImplementationFeature = class extends client_1.TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.ImplementationRequest.type);
      }
      fillClientCapabilities(capabilities) {
        let implementationSupport = ensure(ensure(capabilities, "textDocument"), "implementation");
        implementationSupport.dynamicRegistration = true;
        implementationSupport.linkSupport = true;
      }
      initialize(capabilities, documentSelector) {
        let [id, options] = this.getRegistration(documentSelector, capabilities.implementationProvider);
        if (!id || !options) {
          return;
        }
        this.register({ id, registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideImplementation: (document, position, token) => {
            const client2 = this._client;
            const provideImplementation = (document2, position2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.ImplementationRequest.type, client2.code2ProtocolConverter.asTextDocumentPositionParams(document2, position2), token2).then(client2.protocol2CodeConverter.asDefinitionResult, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.ImplementationRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideImplementation ? middleware.provideImplementation(document, position, token, provideImplementation) : provideImplementation(document, position, token);
          }
        };
        return [vscode_1.languages.registerImplementationProvider(options.documentSelector, provider), provider];
      }
    };
    exports2.ImplementationFeature = ImplementationFeature;
  }
});

// node_modules/vscode-languageclient/lib/common/typeDefinition.js
var require_typeDefinition = __commonJS({
  "node_modules/vscode-languageclient/lib/common/typeDefinition.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.TypeDefinitionFeature = void 0;
    var vscode_1 = require("vscode");
    var vscode_languageserver_protocol_1 = require_main3();
    var client_1 = require_client();
    function ensure(target, key) {
      if (target[key] === void 0) {
        target[key] = {};
      }
      return target[key];
    }
    var TypeDefinitionFeature = class extends client_1.TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.TypeDefinitionRequest.type);
      }
      fillClientCapabilities(capabilities) {
        ensure(ensure(capabilities, "textDocument"), "typeDefinition").dynamicRegistration = true;
        let typeDefinitionSupport = ensure(ensure(capabilities, "textDocument"), "typeDefinition");
        typeDefinitionSupport.dynamicRegistration = true;
        typeDefinitionSupport.linkSupport = true;
      }
      initialize(capabilities, documentSelector) {
        let [id, options] = this.getRegistration(documentSelector, capabilities.typeDefinitionProvider);
        if (!id || !options) {
          return;
        }
        this.register({ id, registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideTypeDefinition: (document, position, token) => {
            const client2 = this._client;
            const provideTypeDefinition = (document2, position2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.TypeDefinitionRequest.type, client2.code2ProtocolConverter.asTextDocumentPositionParams(document2, position2), token2).then(client2.protocol2CodeConverter.asDefinitionResult, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.TypeDefinitionRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideTypeDefinition ? middleware.provideTypeDefinition(document, position, token, provideTypeDefinition) : provideTypeDefinition(document, position, token);
          }
        };
        return [vscode_1.languages.registerTypeDefinitionProvider(options.documentSelector, provider), provider];
      }
    };
    exports2.TypeDefinitionFeature = TypeDefinitionFeature;
  }
});

// node_modules/vscode-languageclient/lib/common/workspaceFolders.js
var require_workspaceFolders = __commonJS({
  "node_modules/vscode-languageclient/lib/common/workspaceFolders.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.WorkspaceFoldersFeature = exports2.arrayDiff = void 0;
    var UUID = require_uuid();
    var vscode_1 = require("vscode");
    var vscode_languageserver_protocol_1 = require_main3();
    function access(target, key) {
      if (target === void 0) {
        return void 0;
      }
      return target[key];
    }
    function arrayDiff(left, right) {
      return left.filter((element) => right.indexOf(element) < 0);
    }
    exports2.arrayDiff = arrayDiff;
    var WorkspaceFoldersFeature = class {
      constructor(_client) {
        this._client = _client;
        this._listeners = /* @__PURE__ */ new Map();
      }
      get registrationType() {
        return vscode_languageserver_protocol_1.DidChangeWorkspaceFoldersNotification.type;
      }
      fillInitializeParams(params) {
        const folders = vscode_1.workspace.workspaceFolders;
        this.initializeWithFolders(folders);
        if (folders === void 0) {
          params.workspaceFolders = null;
        } else {
          params.workspaceFolders = folders.map((folder) => this.asProtocol(folder));
        }
      }
      initializeWithFolders(currentWorkspaceFolders) {
        this._initialFolders = currentWorkspaceFolders;
      }
      fillClientCapabilities(capabilities) {
        capabilities.workspace = capabilities.workspace || {};
        capabilities.workspace.workspaceFolders = true;
      }
      initialize(capabilities) {
        const client2 = this._client;
        client2.onRequest(vscode_languageserver_protocol_1.WorkspaceFoldersRequest.type, (token) => {
          const workspaceFolders = () => {
            const folders = vscode_1.workspace.workspaceFolders;
            if (folders === void 0) {
              return null;
            }
            const result = folders.map((folder) => {
              return this.asProtocol(folder);
            });
            return result;
          };
          const middleware = client2.clientOptions.middleware.workspace;
          return middleware && middleware.workspaceFolders ? middleware.workspaceFolders(token, workspaceFolders) : workspaceFolders(token);
        });
        const value = access(access(access(capabilities, "workspace"), "workspaceFolders"), "changeNotifications");
        let id;
        if (typeof value === "string") {
          id = value;
        } else if (value === true) {
          id = UUID.generateUuid();
        }
        if (id) {
          this.register({ id, registerOptions: void 0 });
        }
      }
      sendInitialEvent(currentWorkspaceFolders) {
        if (this._initialFolders && currentWorkspaceFolders) {
          const removed = arrayDiff(this._initialFolders, currentWorkspaceFolders);
          const added = arrayDiff(currentWorkspaceFolders, this._initialFolders);
          if (added.length > 0 || removed.length > 0) {
            this.doSendEvent(added, removed);
          }
        } else if (this._initialFolders) {
          this.doSendEvent([], this._initialFolders);
        } else if (currentWorkspaceFolders) {
          this.doSendEvent(currentWorkspaceFolders, []);
        }
      }
      doSendEvent(addedFolders, removedFolders) {
        let params = {
          event: {
            added: addedFolders.map((folder) => this.asProtocol(folder)),
            removed: removedFolders.map((folder) => this.asProtocol(folder))
          }
        };
        this._client.sendNotification(vscode_languageserver_protocol_1.DidChangeWorkspaceFoldersNotification.type, params);
      }
      register(data) {
        let id = data.id;
        let client2 = this._client;
        let disposable = vscode_1.workspace.onDidChangeWorkspaceFolders((event) => {
          let didChangeWorkspaceFolders = (event2) => {
            this.doSendEvent(event2.added, event2.removed);
          };
          let middleware = client2.clientOptions.middleware.workspace;
          middleware && middleware.didChangeWorkspaceFolders ? middleware.didChangeWorkspaceFolders(event, didChangeWorkspaceFolders) : didChangeWorkspaceFolders(event);
        });
        this._listeners.set(id, disposable);
        this.sendInitialEvent(vscode_1.workspace.workspaceFolders);
      }
      unregister(id) {
        let disposable = this._listeners.get(id);
        if (disposable === void 0) {
          return;
        }
        this._listeners.delete(id);
        disposable.dispose();
      }
      dispose() {
        for (let disposable of this._listeners.values()) {
          disposable.dispose();
        }
        this._listeners.clear();
      }
      asProtocol(workspaceFolder) {
        if (workspaceFolder === void 0) {
          return null;
        }
        return { uri: this._client.code2ProtocolConverter.asUri(workspaceFolder.uri), name: workspaceFolder.name };
      }
    };
    exports2.WorkspaceFoldersFeature = WorkspaceFoldersFeature;
  }
});

// node_modules/vscode-languageclient/lib/common/foldingRange.js
var require_foldingRange = __commonJS({
  "node_modules/vscode-languageclient/lib/common/foldingRange.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.FoldingRangeFeature = void 0;
    var vscode_1 = require("vscode");
    var vscode_languageserver_protocol_1 = require_main3();
    var client_1 = require_client();
    function ensure(target, key) {
      if (target[key] === void 0) {
        target[key] = {};
      }
      return target[key];
    }
    var FoldingRangeFeature = class _FoldingRangeFeature extends client_1.TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.FoldingRangeRequest.type);
      }
      fillClientCapabilities(capabilities) {
        let capability = ensure(ensure(capabilities, "textDocument"), "foldingRange");
        capability.dynamicRegistration = true;
        capability.rangeLimit = 5e3;
        capability.lineFoldingOnly = true;
      }
      initialize(capabilities, documentSelector) {
        let [id, options] = this.getRegistration(documentSelector, capabilities.foldingRangeProvider);
        if (!id || !options) {
          return;
        }
        this.register({ id, registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideFoldingRanges: (document, context, token) => {
            const client2 = this._client;
            const provideFoldingRanges = (document2, _, token2) => {
              const requestParams = {
                textDocument: client2.code2ProtocolConverter.asTextDocumentIdentifier(document2)
              };
              return client2.sendRequest(vscode_languageserver_protocol_1.FoldingRangeRequest.type, requestParams, token2).then(_FoldingRangeFeature.asFoldingRanges, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.FoldingRangeRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideFoldingRanges ? middleware.provideFoldingRanges(document, context, token, provideFoldingRanges) : provideFoldingRanges(document, context, token);
          }
        };
        return [vscode_1.languages.registerFoldingRangeProvider(options.documentSelector, provider), provider];
      }
      static asFoldingRangeKind(kind) {
        if (kind) {
          switch (kind) {
            case vscode_languageserver_protocol_1.FoldingRangeKind.Comment:
              return vscode_1.FoldingRangeKind.Comment;
            case vscode_languageserver_protocol_1.FoldingRangeKind.Imports:
              return vscode_1.FoldingRangeKind.Imports;
            case vscode_languageserver_protocol_1.FoldingRangeKind.Region:
              return vscode_1.FoldingRangeKind.Region;
          }
        }
        return void 0;
      }
      static asFoldingRanges(foldingRanges) {
        if (Array.isArray(foldingRanges)) {
          return foldingRanges.map((r) => {
            return new vscode_1.FoldingRange(r.startLine, r.endLine, _FoldingRangeFeature.asFoldingRangeKind(r.kind));
          });
        }
        return [];
      }
    };
    exports2.FoldingRangeFeature = FoldingRangeFeature;
  }
});

// node_modules/vscode-languageclient/lib/common/declaration.js
var require_declaration = __commonJS({
  "node_modules/vscode-languageclient/lib/common/declaration.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.DeclarationFeature = void 0;
    var vscode_1 = require("vscode");
    var vscode_languageserver_protocol_1 = require_main3();
    var client_1 = require_client();
    function ensure(target, key) {
      if (target[key] === void 0) {
        target[key] = {};
      }
      return target[key];
    }
    var DeclarationFeature = class extends client_1.TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.DeclarationRequest.type);
      }
      fillClientCapabilities(capabilities) {
        const declarationSupport = ensure(ensure(capabilities, "textDocument"), "declaration");
        declarationSupport.dynamicRegistration = true;
        declarationSupport.linkSupport = true;
      }
      initialize(capabilities, documentSelector) {
        const [id, options] = this.getRegistration(documentSelector, capabilities.declarationProvider);
        if (!id || !options) {
          return;
        }
        this.register({ id, registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideDeclaration: (document, position, token) => {
            const client2 = this._client;
            const provideDeclaration = (document2, position2, token2) => {
              return client2.sendRequest(vscode_languageserver_protocol_1.DeclarationRequest.type, client2.code2ProtocolConverter.asTextDocumentPositionParams(document2, position2), token2).then(client2.protocol2CodeConverter.asDeclarationResult, (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.DeclarationRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideDeclaration ? middleware.provideDeclaration(document, position, token, provideDeclaration) : provideDeclaration(document, position, token);
          }
        };
        return [vscode_1.languages.registerDeclarationProvider(options.documentSelector, provider), provider];
      }
    };
    exports2.DeclarationFeature = DeclarationFeature;
  }
});

// node_modules/vscode-languageclient/lib/common/selectionRange.js
var require_selectionRange = __commonJS({
  "node_modules/vscode-languageclient/lib/common/selectionRange.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.SelectionRangeFeature = void 0;
    var vscode_1 = require("vscode");
    var vscode_languageserver_protocol_1 = require_main3();
    var client_1 = require_client();
    function ensure(target, key) {
      if (target[key] === void 0) {
        target[key] = /* @__PURE__ */ Object.create(null);
      }
      return target[key];
    }
    var SelectionRangeFeature = class extends client_1.TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.SelectionRangeRequest.type);
      }
      fillClientCapabilities(capabilities) {
        let capability = ensure(ensure(capabilities, "textDocument"), "selectionRange");
        capability.dynamicRegistration = true;
      }
      initialize(capabilities, documentSelector) {
        let [id, options] = this.getRegistration(documentSelector, capabilities.selectionRangeProvider);
        if (!id || !options) {
          return;
        }
        this.register({ id, registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideSelectionRanges: (document, positions, token) => {
            const client2 = this._client;
            const provideSelectionRanges = (document2, positions2, token2) => {
              const requestParams = {
                textDocument: client2.code2ProtocolConverter.asTextDocumentIdentifier(document2),
                positions: client2.code2ProtocolConverter.asPositions(positions2)
              };
              return client2.sendRequest(vscode_languageserver_protocol_1.SelectionRangeRequest.type, requestParams, token2).then((ranges) => client2.protocol2CodeConverter.asSelectionRanges(ranges), (error) => {
                return client2.handleFailedRequest(vscode_languageserver_protocol_1.SelectionRangeRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideSelectionRanges ? middleware.provideSelectionRanges(document, positions, token, provideSelectionRanges) : provideSelectionRanges(document, positions, token);
          }
        };
        return [vscode_1.languages.registerSelectionRangeProvider(options.documentSelector, provider), provider];
      }
    };
    exports2.SelectionRangeFeature = SelectionRangeFeature;
  }
});

// node_modules/vscode-languageclient/lib/common/progress.js
var require_progress = __commonJS({
  "node_modules/vscode-languageclient/lib/common/progress.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.ProgressFeature = void 0;
    var vscode_languageserver_protocol_1 = require_main3();
    var progressPart_1 = require_progressPart();
    function ensure(target, key) {
      if (target[key] === void 0) {
        target[key] = /* @__PURE__ */ Object.create(null);
      }
      return target[key];
    }
    var ProgressFeature = class {
      constructor(_client) {
        this._client = _client;
        this.activeParts = /* @__PURE__ */ new Set();
      }
      fillClientCapabilities(capabilities) {
        ensure(capabilities, "window").workDoneProgress = true;
      }
      initialize() {
        const client2 = this._client;
        const deleteHandler = (part) => {
          this.activeParts.delete(part);
        };
        const createHandler = (params) => {
          this.activeParts.add(new progressPart_1.ProgressPart(this._client, params.token, deleteHandler));
        };
        client2.onRequest(vscode_languageserver_protocol_1.WorkDoneProgressCreateRequest.type, createHandler);
      }
      dispose() {
        for (const part of this.activeParts) {
          part.done();
        }
        this.activeParts.clear();
      }
    };
    exports2.ProgressFeature = ProgressFeature;
  }
});

// node_modules/vscode-languageclient/lib/common/callHierarchy.js
var require_callHierarchy = __commonJS({
  "node_modules/vscode-languageclient/lib/common/callHierarchy.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.CallHierarchyFeature = void 0;
    var vscode_1 = require("vscode");
    var vscode_languageserver_protocol_1 = require_main3();
    var client_1 = require_client();
    function ensure(target, key) {
      if (target[key] === void 0) {
        target[key] = {};
      }
      return target[key];
    }
    var CallHierarchyProvider = class {
      constructor(client2) {
        this.client = client2;
        this.middleware = client2.clientOptions.middleware;
      }
      prepareCallHierarchy(document, position, token) {
        const client2 = this.client;
        const middleware = this.middleware;
        const prepareCallHierarchy = (document2, position2, token2) => {
          const params = client2.code2ProtocolConverter.asTextDocumentPositionParams(document2, position2);
          return client2.sendRequest(vscode_languageserver_protocol_1.CallHierarchyPrepareRequest.type, params, token2).then((result) => {
            return client2.protocol2CodeConverter.asCallHierarchyItems(result);
          }, (error) => {
            return client2.handleFailedRequest(vscode_languageserver_protocol_1.CallHierarchyPrepareRequest.type, error, null);
          });
        };
        return middleware.prepareCallHierarchy ? middleware.prepareCallHierarchy(document, position, token, prepareCallHierarchy) : prepareCallHierarchy(document, position, token);
      }
      provideCallHierarchyIncomingCalls(item, token) {
        const client2 = this.client;
        const middleware = this.middleware;
        const provideCallHierarchyIncomingCalls = (item2, token2) => {
          const params = {
            item: client2.code2ProtocolConverter.asCallHierarchyItem(item2)
          };
          return client2.sendRequest(vscode_languageserver_protocol_1.CallHierarchyIncomingCallsRequest.type, params, token2).then((result) => {
            return client2.protocol2CodeConverter.asCallHierarchyIncomingCalls(result);
          }, (error) => {
            return client2.handleFailedRequest(vscode_languageserver_protocol_1.CallHierarchyIncomingCallsRequest.type, error, null);
          });
        };
        return middleware.provideCallHierarchyIncomingCalls ? middleware.provideCallHierarchyIncomingCalls(item, token, provideCallHierarchyIncomingCalls) : provideCallHierarchyIncomingCalls(item, token);
      }
      provideCallHierarchyOutgoingCalls(item, token) {
        const client2 = this.client;
        const middleware = this.middleware;
        const provideCallHierarchyOutgoingCalls = (item2, token2) => {
          const params = {
            item: client2.code2ProtocolConverter.asCallHierarchyItem(item2)
          };
          return client2.sendRequest(vscode_languageserver_protocol_1.CallHierarchyOutgoingCallsRequest.type, params, token2).then((result) => {
            return client2.protocol2CodeConverter.asCallHierarchyOutgoingCalls(result);
          }, (error) => {
            return client2.handleFailedRequest(vscode_languageserver_protocol_1.CallHierarchyOutgoingCallsRequest.type, error, null);
          });
        };
        return middleware.provideCallHierarchyOutgoingCalls ? middleware.provideCallHierarchyOutgoingCalls(item, token, provideCallHierarchyOutgoingCalls) : provideCallHierarchyOutgoingCalls(item, token);
      }
    };
    var CallHierarchyFeature = class extends client_1.TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.CallHierarchyPrepareRequest.type);
      }
      fillClientCapabilities(cap) {
        const capabilities = cap;
        const capability = ensure(ensure(capabilities, "textDocument"), "callHierarchy");
        capability.dynamicRegistration = true;
      }
      initialize(capabilities, documentSelector) {
        const [id, options] = this.getRegistration(documentSelector, capabilities.callHierarchyProvider);
        if (!id || !options) {
          return;
        }
        this.register({ id, registerOptions: options });
      }
      registerLanguageProvider(options) {
        const client2 = this._client;
        const provider = new CallHierarchyProvider(client2);
        return [vscode_1.languages.registerCallHierarchyProvider(options.documentSelector, provider), provider];
      }
    };
    exports2.CallHierarchyFeature = CallHierarchyFeature;
  }
});

// node_modules/vscode-languageclient/lib/common/semanticTokens.js
var require_semanticTokens = __commonJS({
  "node_modules/vscode-languageclient/lib/common/semanticTokens.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.SemanticTokensFeature = void 0;
    var vscode = require("vscode");
    var client_1 = require_client();
    var vscode_languageserver_protocol_1 = require_main3();
    var Is = require_is();
    function ensure(target, key) {
      if (target[key] === void 0) {
        target[key] = {};
      }
      return target[key];
    }
    var SemanticTokensFeature = class extends client_1.TextDocumentFeature {
      constructor(client2) {
        super(client2, vscode_languageserver_protocol_1.SemanticTokensRegistrationType.type);
      }
      fillClientCapabilities(capabilities) {
        const capability = ensure(ensure(capabilities, "textDocument"), "semanticTokens");
        capability.dynamicRegistration = true;
        capability.tokenTypes = [
          vscode_languageserver_protocol_1.SemanticTokenTypes.namespace,
          vscode_languageserver_protocol_1.SemanticTokenTypes.type,
          vscode_languageserver_protocol_1.SemanticTokenTypes.class,
          vscode_languageserver_protocol_1.SemanticTokenTypes.enum,
          vscode_languageserver_protocol_1.SemanticTokenTypes.interface,
          vscode_languageserver_protocol_1.SemanticTokenTypes.struct,
          vscode_languageserver_protocol_1.SemanticTokenTypes.typeParameter,
          vscode_languageserver_protocol_1.SemanticTokenTypes.parameter,
          vscode_languageserver_protocol_1.SemanticTokenTypes.variable,
          vscode_languageserver_protocol_1.SemanticTokenTypes.property,
          vscode_languageserver_protocol_1.SemanticTokenTypes.enumMember,
          vscode_languageserver_protocol_1.SemanticTokenTypes.event,
          vscode_languageserver_protocol_1.SemanticTokenTypes.function,
          vscode_languageserver_protocol_1.SemanticTokenTypes.method,
          vscode_languageserver_protocol_1.SemanticTokenTypes.macro,
          vscode_languageserver_protocol_1.SemanticTokenTypes.keyword,
          vscode_languageserver_protocol_1.SemanticTokenTypes.modifier,
          vscode_languageserver_protocol_1.SemanticTokenTypes.comment,
          vscode_languageserver_protocol_1.SemanticTokenTypes.string,
          vscode_languageserver_protocol_1.SemanticTokenTypes.number,
          vscode_languageserver_protocol_1.SemanticTokenTypes.regexp,
          vscode_languageserver_protocol_1.SemanticTokenTypes.operator
        ];
        capability.tokenModifiers = [
          vscode_languageserver_protocol_1.SemanticTokenModifiers.declaration,
          vscode_languageserver_protocol_1.SemanticTokenModifiers.definition,
          vscode_languageserver_protocol_1.SemanticTokenModifiers.readonly,
          vscode_languageserver_protocol_1.SemanticTokenModifiers.static,
          vscode_languageserver_protocol_1.SemanticTokenModifiers.deprecated,
          vscode_languageserver_protocol_1.SemanticTokenModifiers.abstract,
          vscode_languageserver_protocol_1.SemanticTokenModifiers.async,
          vscode_languageserver_protocol_1.SemanticTokenModifiers.modification,
          vscode_languageserver_protocol_1.SemanticTokenModifiers.documentation,
          vscode_languageserver_protocol_1.SemanticTokenModifiers.defaultLibrary
        ];
        capability.formats = [vscode_languageserver_protocol_1.TokenFormat.Relative];
        capability.requests = {
          range: true,
          full: {
            delta: true
          }
        };
        capability.multilineTokenSupport = false;
        capability.overlappingTokenSupport = false;
        ensure(ensure(capabilities, "workspace"), "semanticTokens").refreshSupport = true;
      }
      initialize(capabilities, documentSelector) {
        const client2 = this._client;
        client2.onRequest(vscode_languageserver_protocol_1.SemanticTokensRefreshRequest.type, async () => {
          for (const provider of this.getAllProviders()) {
            provider.onDidChangeSemanticTokensEmitter.fire();
          }
        });
        const [id, options] = this.getRegistration(documentSelector, capabilities.semanticTokensProvider);
        if (!id || !options) {
          return;
        }
        this.register({ id, registerOptions: options });
      }
      registerLanguageProvider(options) {
        const fullProvider = Is.boolean(options.full) ? options.full : options.full !== void 0;
        const hasEditProvider = options.full !== void 0 && typeof options.full !== "boolean" && options.full.delta === true;
        const eventEmitter = new vscode.EventEmitter();
        const documentProvider = fullProvider ? {
          onDidChangeSemanticTokens: eventEmitter.event,
          provideDocumentSemanticTokens: (document, token) => {
            const client3 = this._client;
            const middleware = client3.clientOptions.middleware;
            const provideDocumentSemanticTokens = (document2, token2) => {
              const params = {
                textDocument: client3.code2ProtocolConverter.asTextDocumentIdentifier(document2)
              };
              return client3.sendRequest(vscode_languageserver_protocol_1.SemanticTokensRequest.type, params, token2).then((result) => {
                return client3.protocol2CodeConverter.asSemanticTokens(result);
              }, (error) => {
                return client3.handleFailedRequest(vscode_languageserver_protocol_1.SemanticTokensRequest.type, error, null);
              });
            };
            return middleware.provideDocumentSemanticTokens ? middleware.provideDocumentSemanticTokens(document, token, provideDocumentSemanticTokens) : provideDocumentSemanticTokens(document, token);
          },
          provideDocumentSemanticTokensEdits: hasEditProvider ? (document, previousResultId, token) => {
            const client3 = this._client;
            const middleware = client3.clientOptions.middleware;
            const provideDocumentSemanticTokensEdits = (document2, previousResultId2, token2) => {
              const params = {
                textDocument: client3.code2ProtocolConverter.asTextDocumentIdentifier(document2),
                previousResultId: previousResultId2
              };
              return client3.sendRequest(vscode_languageserver_protocol_1.SemanticTokensDeltaRequest.type, params, token2).then((result) => {
                if (vscode_languageserver_protocol_1.SemanticTokens.is(result)) {
                  return client3.protocol2CodeConverter.asSemanticTokens(result);
                } else {
                  return client3.protocol2CodeConverter.asSemanticTokensEdits(result);
                }
              }, (error) => {
                return client3.handleFailedRequest(vscode_languageserver_protocol_1.SemanticTokensDeltaRequest.type, error, null);
              });
            };
            return middleware.provideDocumentSemanticTokensEdits ? middleware.provideDocumentSemanticTokensEdits(document, previousResultId, token, provideDocumentSemanticTokensEdits) : provideDocumentSemanticTokensEdits(document, previousResultId, token);
          } : void 0
        } : void 0;
        const hasRangeProvider = options.range === true;
        const rangeProvider = hasRangeProvider ? {
          provideDocumentRangeSemanticTokens: (document, range, token) => {
            const client3 = this._client;
            const middleware = client3.clientOptions.middleware;
            const provideDocumentRangeSemanticTokens = (document2, range2, token2) => {
              const params = {
                textDocument: client3.code2ProtocolConverter.asTextDocumentIdentifier(document2),
                range: client3.code2ProtocolConverter.asRange(range2)
              };
              return client3.sendRequest(vscode_languageserver_protocol_1.SemanticTokensRangeRequest.type, params, token2).then((result) => {
                return client3.protocol2CodeConverter.asSemanticTokens(result);
              }, (error) => {
                return client3.handleFailedRequest(vscode_languageserver_protocol_1.SemanticTokensRangeRequest.type, error, null);
              });
            };
            return middleware.provideDocumentRangeSemanticTokens ? middleware.provideDocumentRangeSemanticTokens(document, range, token, provideDocumentRangeSemanticTokens) : provideDocumentRangeSemanticTokens(document, range, token);
          }
        } : void 0;
        const disposables = [];
        const client2 = this._client;
        const legend = client2.protocol2CodeConverter.asSemanticTokensLegend(options.legend);
        if (documentProvider !== void 0) {
          disposables.push(vscode.languages.registerDocumentSemanticTokensProvider(options.documentSelector, documentProvider, legend));
        }
        if (rangeProvider !== void 0) {
          disposables.push(vscode.languages.registerDocumentRangeSemanticTokensProvider(options.documentSelector, rangeProvider, legend));
        }
        return [new vscode.Disposable(() => disposables.forEach((item) => item.dispose())), { range: rangeProvider, full: documentProvider, onDidChangeSemanticTokensEmitter: eventEmitter }];
      }
    };
    exports2.SemanticTokensFeature = SemanticTokensFeature;
  }
});

// node_modules/concat-map/index.js
var require_concat_map = __commonJS({
  "node_modules/concat-map/index.js"(exports2, module2) {
    module2.exports = function(xs, fn) {
      var res = [];
      for (var i = 0; i < xs.length; i++) {
        var x = fn(xs[i], i);
        if (isArray(x)) res.push.apply(res, x);
        else res.push(x);
      }
      return res;
    };
    var isArray = Array.isArray || function(xs) {
      return Object.prototype.toString.call(xs) === "[object Array]";
    };
  }
});

// node_modules/balanced-match/index.js
var require_balanced_match = __commonJS({
  "node_modules/balanced-match/index.js"(exports2, module2) {
    "use strict";
    module2.exports = balanced;
    function balanced(a, b, str) {
      if (a instanceof RegExp) a = maybeMatch(a, str);
      if (b instanceof RegExp) b = maybeMatch(b, str);
      var r = range(a, b, str);
      return r && {
        start: r[0],
        end: r[1],
        pre: str.slice(0, r[0]),
        body: str.slice(r[0] + a.length, r[1]),
        post: str.slice(r[1] + b.length)
      };
    }
    function maybeMatch(reg, str) {
      var m = str.match(reg);
      return m ? m[0] : null;
    }
    balanced.range = range;
    function range(a, b, str) {
      var begs, beg, left, right, result;
      var ai = str.indexOf(a);
      var bi = str.indexOf(b, ai + 1);
      var i = ai;
      if (ai >= 0 && bi > 0) {
        if (a === b) {
          return [ai, bi];
        }
        begs = [];
        left = str.length;
        while (i >= 0 && !result) {
          if (i == ai) {
            begs.push(i);
            ai = str.indexOf(a, i + 1);
          } else if (begs.length == 1) {
            result = [begs.pop(), bi];
          } else {
            beg = begs.pop();
            if (beg < left) {
              left = beg;
              right = bi;
            }
            bi = str.indexOf(b, i + 1);
          }
          i = ai < bi && ai >= 0 ? ai : bi;
        }
        if (begs.length) {
          result = [left, right];
        }
      }
      return result;
    }
  }
});

// node_modules/brace-expansion/index.js
var require_brace_expansion = __commonJS({
  "node_modules/brace-expansion/index.js"(exports2, module2) {
    var concatMap = require_concat_map();
    var balanced = require_balanced_match();
    module2.exports = expandTop;
    var escSlash = "\0SLASH" + Math.random() + "\0";
    var escOpen = "\0OPEN" + Math.random() + "\0";
    var escClose = "\0CLOSE" + Math.random() + "\0";
    var escComma = "\0COMMA" + Math.random() + "\0";
    var escPeriod = "\0PERIOD" + Math.random() + "\0";
    function numeric(str) {
      return parseInt(str, 10) == str ? parseInt(str, 10) : str.charCodeAt(0);
    }
    function escapeBraces(str) {
      return str.split("\\\\").join(escSlash).split("\\{").join(escOpen).split("\\}").join(escClose).split("\\,").join(escComma).split("\\.").join(escPeriod);
    }
    function unescapeBraces(str) {
      return str.split(escSlash).join("\\").split(escOpen).join("{").split(escClose).join("}").split(escComma).join(",").split(escPeriod).join(".");
    }
    function parseCommaParts(str) {
      if (!str)
        return [""];
      var parts = [];
      var m = balanced("{", "}", str);
      if (!m)
        return str.split(",");
      var pre = m.pre;
      var body = m.body;
      var post = m.post;
      var p = pre.split(",");
      p[p.length - 1] += "{" + body + "}";
      var postParts = parseCommaParts(post);
      if (post.length) {
        p[p.length - 1] += postParts.shift();
        p.push.apply(p, postParts);
      }
      parts.push.apply(parts, p);
      return parts;
    }
    function expandTop(str) {
      if (!str)
        return [];
      if (str.substr(0, 2) === "{}") {
        str = "\\{\\}" + str.substr(2);
      }
      return expand(escapeBraces(str), true).map(unescapeBraces);
    }
    function embrace(str) {
      return "{" + str + "}";
    }
    function isPadded(el) {
      return /^-?0\d/.test(el);
    }
    function lte(i, y) {
      return i <= y;
    }
    function gte(i, y) {
      return i >= y;
    }
    function expand(str, isTop) {
      var expansions = [];
      var m = balanced("{", "}", str);
      if (!m || /\$$/.test(m.pre)) return [str];
      var isNumericSequence = /^-?\d+\.\.-?\d+(?:\.\.-?\d+)?$/.test(m.body);
      var isAlphaSequence = /^[a-zA-Z]\.\.[a-zA-Z](?:\.\.-?\d+)?$/.test(m.body);
      var isSequence = isNumericSequence || isAlphaSequence;
      var isOptions = m.body.indexOf(",") >= 0;
      if (!isSequence && !isOptions) {
        if (m.post.match(/,.*\}/)) {
          str = m.pre + "{" + m.body + escClose + m.post;
          return expand(str);
        }
        return [str];
      }
      var n;
      if (isSequence) {
        n = m.body.split(/\.\./);
      } else {
        n = parseCommaParts(m.body);
        if (n.length === 1) {
          n = expand(n[0], false).map(embrace);
          if (n.length === 1) {
            var post = m.post.length ? expand(m.post, false) : [""];
            return post.map(function(p) {
              return m.pre + n[0] + p;
            });
          }
        }
      }
      var pre = m.pre;
      var post = m.post.length ? expand(m.post, false) : [""];
      var N;
      if (isSequence) {
        var x = numeric(n[0]);
        var y = numeric(n[1]);
        var width = Math.max(n[0].length, n[1].length);
        var incr = n.length == 3 ? Math.abs(numeric(n[2])) : 1;
        var test = lte;
        var reverse = y < x;
        if (reverse) {
          incr *= -1;
          test = gte;
        }
        var pad = n.some(isPadded);
        N = [];
        for (var i = x; test(i, y); i += incr) {
          var c;
          if (isAlphaSequence) {
            c = String.fromCharCode(i);
            if (c === "\\")
              c = "";
          } else {
            c = String(i);
            if (pad) {
              var need = width - c.length;
              if (need > 0) {
                var z = new Array(need + 1).join("0");
                if (i < 0)
                  c = "-" + z + c.slice(1);
                else
                  c = z + c;
              }
            }
          }
          N.push(c);
        }
      } else {
        N = concatMap(n, function(el) {
          return expand(el, false);
        });
      }
      for (var j = 0; j < N.length; j++) {
        for (var k = 0; k < post.length; k++) {
          var expansion = pre + N[j] + post[k];
          if (!isTop || isSequence || expansion)
            expansions.push(expansion);
        }
      }
      return expansions;
    }
  }
});

// node_modules/minimatch/minimatch.js
var require_minimatch = __commonJS({
  "node_modules/minimatch/minimatch.js"(exports2, module2) {
    module2.exports = minimatch;
    minimatch.Minimatch = Minimatch;
    var path2 = function() {
      try {
        return require("path");
      } catch (e) {
      }
    }() || {
      sep: "/"
    };
    minimatch.sep = path2.sep;
    var GLOBSTAR = minimatch.GLOBSTAR = Minimatch.GLOBSTAR = {};
    var expand = require_brace_expansion();
    var plTypes = {
      "!": { open: "(?:(?!(?:", close: "))[^/]*?)" },
      "?": { open: "(?:", close: ")?" },
      "+": { open: "(?:", close: ")+" },
      "*": { open: "(?:", close: ")*" },
      "@": { open: "(?:", close: ")" }
    };
    var qmark = "[^/]";
    var star = qmark + "*?";
    var twoStarDot = "(?:(?!(?:\\/|^)(?:\\.{1,2})($|\\/)).)*?";
    var twoStarNoDot = "(?:(?!(?:\\/|^)\\.).)*?";
    var reSpecials = charSet("().*{}+?[]^$\\!");
    function charSet(s) {
      return s.split("").reduce(function(set, c) {
        set[c] = true;
        return set;
      }, {});
    }
    var slashSplit = /\/+/;
    minimatch.filter = filter;
    function filter(pattern, options) {
      options = options || {};
      return function(p, i, list) {
        return minimatch(p, pattern, options);
      };
    }
    function ext(a, b) {
      b = b || {};
      var t = {};
      Object.keys(a).forEach(function(k) {
        t[k] = a[k];
      });
      Object.keys(b).forEach(function(k) {
        t[k] = b[k];
      });
      return t;
    }
    minimatch.defaults = function(def) {
      if (!def || typeof def !== "object" || !Object.keys(def).length) {
        return minimatch;
      }
      var orig = minimatch;
      var m = function minimatch2(p, pattern, options) {
        return orig(p, pattern, ext(def, options));
      };
      m.Minimatch = function Minimatch2(pattern, options) {
        return new orig.Minimatch(pattern, ext(def, options));
      };
      m.Minimatch.defaults = function defaults(options) {
        return orig.defaults(ext(def, options)).Minimatch;
      };
      m.filter = function filter2(pattern, options) {
        return orig.filter(pattern, ext(def, options));
      };
      m.defaults = function defaults(options) {
        return orig.defaults(ext(def, options));
      };
      m.makeRe = function makeRe2(pattern, options) {
        return orig.makeRe(pattern, ext(def, options));
      };
      m.braceExpand = function braceExpand2(pattern, options) {
        return orig.braceExpand(pattern, ext(def, options));
      };
      m.match = function(list, pattern, options) {
        return orig.match(list, pattern, ext(def, options));
      };
      return m;
    };
    Minimatch.defaults = function(def) {
      return minimatch.defaults(def).Minimatch;
    };
    function minimatch(p, pattern, options) {
      assertValidPattern(pattern);
      if (!options) options = {};
      if (!options.nocomment && pattern.charAt(0) === "#") {
        return false;
      }
      return new Minimatch(pattern, options).match(p);
    }
    function Minimatch(pattern, options) {
      if (!(this instanceof Minimatch)) {
        return new Minimatch(pattern, options);
      }
      assertValidPattern(pattern);
      if (!options) options = {};
      pattern = pattern.trim();
      if (!options.allowWindowsEscape && path2.sep !== "/") {
        pattern = pattern.split(path2.sep).join("/");
      }
      this.options = options;
      this.set = [];
      this.pattern = pattern;
      this.regexp = null;
      this.negate = false;
      this.comment = false;
      this.empty = false;
      this.partial = !!options.partial;
      this.make();
    }
    Minimatch.prototype.debug = function() {
    };
    Minimatch.prototype.make = make;
    function make() {
      var pattern = this.pattern;
      var options = this.options;
      if (!options.nocomment && pattern.charAt(0) === "#") {
        this.comment = true;
        return;
      }
      if (!pattern) {
        this.empty = true;
        return;
      }
      this.parseNegate();
      var set = this.globSet = this.braceExpand();
      if (options.debug) this.debug = function debug() {
        console.error.apply(console, arguments);
      };
      this.debug(this.pattern, set);
      set = this.globParts = set.map(function(s) {
        return s.split(slashSplit);
      });
      this.debug(this.pattern, set);
      set = set.map(function(s, si, set2) {
        return s.map(this.parse, this);
      }, this);
      this.debug(this.pattern, set);
      set = set.filter(function(s) {
        return s.indexOf(false) === -1;
      });
      this.debug(this.pattern, set);
      this.set = set;
    }
    Minimatch.prototype.parseNegate = parseNegate;
    function parseNegate() {
      var pattern = this.pattern;
      var negate = false;
      var options = this.options;
      var negateOffset = 0;
      if (options.nonegate) return;
      for (var i = 0, l = pattern.length; i < l && pattern.charAt(i) === "!"; i++) {
        negate = !negate;
        negateOffset++;
      }
      if (negateOffset) this.pattern = pattern.substr(negateOffset);
      this.negate = negate;
    }
    minimatch.braceExpand = function(pattern, options) {
      return braceExpand(pattern, options);
    };
    Minimatch.prototype.braceExpand = braceExpand;
    function braceExpand(pattern, options) {
      if (!options) {
        if (this instanceof Minimatch) {
          options = this.options;
        } else {
          options = {};
        }
      }
      pattern = typeof pattern === "undefined" ? this.pattern : pattern;
      assertValidPattern(pattern);
      if (options.nobrace || !/\{(?:(?!\{).)*\}/.test(pattern)) {
        return [pattern];
      }
      return expand(pattern);
    }
    var MAX_PATTERN_LENGTH = 1024 * 64;
    var assertValidPattern = function(pattern) {
      if (typeof pattern !== "string") {
        throw new TypeError("invalid pattern");
      }
      if (pattern.length > MAX_PATTERN_LENGTH) {
        throw new TypeError("pattern is too long");
      }
    };
    Minimatch.prototype.parse = parse;
    var SUBPARSE = {};
    function parse(pattern, isSub) {
      assertValidPattern(pattern);
      var options = this.options;
      if (pattern === "**") {
        if (!options.noglobstar)
          return GLOBSTAR;
        else
          pattern = "*";
      }
      if (pattern === "") return "";
      var re = "";
      var hasMagic = !!options.nocase;
      var escaping = false;
      var patternListStack = [];
      var negativeLists = [];
      var stateChar;
      var inClass = false;
      var reClassStart = -1;
      var classStart = -1;
      var patternStart = pattern.charAt(0) === "." ? "" : options.dot ? "(?!(?:^|\\/)\\.{1,2}(?:$|\\/))" : "(?!\\.)";
      var self = this;
      function clearStateChar() {
        if (stateChar) {
          switch (stateChar) {
            case "*":
              re += star;
              hasMagic = true;
              break;
            case "?":
              re += qmark;
              hasMagic = true;
              break;
            default:
              re += "\\" + stateChar;
              break;
          }
          self.debug("clearStateChar %j %j", stateChar, re);
          stateChar = false;
        }
      }
      for (var i = 0, len = pattern.length, c; i < len && (c = pattern.charAt(i)); i++) {
        this.debug("%s	%s %s %j", pattern, i, re, c);
        if (escaping && reSpecials[c]) {
          re += "\\" + c;
          escaping = false;
          continue;
        }
        switch (c) {
          /* istanbul ignore next */
          case "/": {
            return false;
          }
          case "\\":
            clearStateChar();
            escaping = true;
            continue;
          // the various stateChar values
          // for the "extglob" stuff.
          case "?":
          case "*":
          case "+":
          case "@":
          case "!":
            this.debug("%s	%s %s %j <-- stateChar", pattern, i, re, c);
            if (inClass) {
              this.debug("  in class");
              if (c === "!" && i === classStart + 1) c = "^";
              re += c;
              continue;
            }
            self.debug("call clearStateChar %j", stateChar);
            clearStateChar();
            stateChar = c;
            if (options.noext) clearStateChar();
            continue;
          case "(":
            if (inClass) {
              re += "(";
              continue;
            }
            if (!stateChar) {
              re += "\\(";
              continue;
            }
            patternListStack.push({
              type: stateChar,
              start: i - 1,
              reStart: re.length,
              open: plTypes[stateChar].open,
              close: plTypes[stateChar].close
            });
            re += stateChar === "!" ? "(?:(?!(?:" : "(?:";
            this.debug("plType %j %j", stateChar, re);
            stateChar = false;
            continue;
          case ")":
            if (inClass || !patternListStack.length) {
              re += "\\)";
              continue;
            }
            clearStateChar();
            hasMagic = true;
            var pl = patternListStack.pop();
            re += pl.close;
            if (pl.type === "!") {
              negativeLists.push(pl);
            }
            pl.reEnd = re.length;
            continue;
          case "|":
            if (inClass || !patternListStack.length || escaping) {
              re += "\\|";
              escaping = false;
              continue;
            }
            clearStateChar();
            re += "|";
            continue;
          // these are mostly the same in regexp and glob
          case "[":
            clearStateChar();
            if (inClass) {
              re += "\\" + c;
              continue;
            }
            inClass = true;
            classStart = i;
            reClassStart = re.length;
            re += c;
            continue;
          case "]":
            if (i === classStart + 1 || !inClass) {
              re += "\\" + c;
              escaping = false;
              continue;
            }
            var cs = pattern.substring(classStart + 1, i);
            try {
              RegExp("[" + cs + "]");
            } catch (er) {
              var sp = this.parse(cs, SUBPARSE);
              re = re.substr(0, reClassStart) + "\\[" + sp[0] + "\\]";
              hasMagic = hasMagic || sp[1];
              inClass = false;
              continue;
            }
            hasMagic = true;
            inClass = false;
            re += c;
            continue;
          default:
            clearStateChar();
            if (escaping) {
              escaping = false;
            } else if (reSpecials[c] && !(c === "^" && inClass)) {
              re += "\\";
            }
            re += c;
        }
      }
      if (inClass) {
        cs = pattern.substr(classStart + 1);
        sp = this.parse(cs, SUBPARSE);
        re = re.substr(0, reClassStart) + "\\[" + sp[0];
        hasMagic = hasMagic || sp[1];
      }
      for (pl = patternListStack.pop(); pl; pl = patternListStack.pop()) {
        var tail = re.slice(pl.reStart + pl.open.length);
        this.debug("setting tail", re, pl);
        tail = tail.replace(/((?:\\{2}){0,64})(\\?)\|/g, function(_, $1, $2) {
          if (!$2) {
            $2 = "\\";
          }
          return $1 + $1 + $2 + "|";
        });
        this.debug("tail=%j\n   %s", tail, tail, pl, re);
        var t = pl.type === "*" ? star : pl.type === "?" ? qmark : "\\" + pl.type;
        hasMagic = true;
        re = re.slice(0, pl.reStart) + t + "\\(" + tail;
      }
      clearStateChar();
      if (escaping) {
        re += "\\\\";
      }
      var addPatternStart = false;
      switch (re.charAt(0)) {
        case "[":
        case ".":
        case "(":
          addPatternStart = true;
      }
      for (var n = negativeLists.length - 1; n > -1; n--) {
        var nl = negativeLists[n];
        var nlBefore = re.slice(0, nl.reStart);
        var nlFirst = re.slice(nl.reStart, nl.reEnd - 8);
        var nlLast = re.slice(nl.reEnd - 8, nl.reEnd);
        var nlAfter = re.slice(nl.reEnd);
        nlLast += nlAfter;
        var openParensBefore = nlBefore.split("(").length - 1;
        var cleanAfter = nlAfter;
        for (i = 0; i < openParensBefore; i++) {
          cleanAfter = cleanAfter.replace(/\)[+*?]?/, "");
        }
        nlAfter = cleanAfter;
        var dollar = "";
        if (nlAfter === "" && isSub !== SUBPARSE) {
          dollar = "$";
        }
        var newRe = nlBefore + nlFirst + nlAfter + dollar + nlLast;
        re = newRe;
      }
      if (re !== "" && hasMagic) {
        re = "(?=.)" + re;
      }
      if (addPatternStart) {
        re = patternStart + re;
      }
      if (isSub === SUBPARSE) {
        return [re, hasMagic];
      }
      if (!hasMagic) {
        return globUnescape(pattern);
      }
      var flags = options.nocase ? "i" : "";
      try {
        var regExp = new RegExp("^" + re + "$", flags);
      } catch (er) {
        return new RegExp("$.");
      }
      regExp._glob = pattern;
      regExp._src = re;
      return regExp;
    }
    minimatch.makeRe = function(pattern, options) {
      return new Minimatch(pattern, options || {}).makeRe();
    };
    Minimatch.prototype.makeRe = makeRe;
    function makeRe() {
      if (this.regexp || this.regexp === false) return this.regexp;
      var set = this.set;
      if (!set.length) {
        this.regexp = false;
        return this.regexp;
      }
      var options = this.options;
      var twoStar = options.noglobstar ? star : options.dot ? twoStarDot : twoStarNoDot;
      var flags = options.nocase ? "i" : "";
      var re = set.map(function(pattern) {
        return pattern.map(function(p) {
          return p === GLOBSTAR ? twoStar : typeof p === "string" ? regExpEscape(p) : p._src;
        }).join("\\/");
      }).join("|");
      re = "^(?:" + re + ")$";
      if (this.negate) re = "^(?!" + re + ").*$";
      try {
        this.regexp = new RegExp(re, flags);
      } catch (ex) {
        this.regexp = false;
      }
      return this.regexp;
    }
    minimatch.match = function(list, pattern, options) {
      options = options || {};
      var mm = new Minimatch(pattern, options);
      list = list.filter(function(f) {
        return mm.match(f);
      });
      if (mm.options.nonull && !list.length) {
        list.push(pattern);
      }
      return list;
    };
    Minimatch.prototype.match = function match(f, partial) {
      if (typeof partial === "undefined") partial = this.partial;
      this.debug("match", f, this.pattern);
      if (this.comment) return false;
      if (this.empty) return f === "";
      if (f === "/" && partial) return true;
      var options = this.options;
      if (path2.sep !== "/") {
        f = f.split(path2.sep).join("/");
      }
      f = f.split(slashSplit);
      this.debug(this.pattern, "split", f);
      var set = this.set;
      this.debug(this.pattern, "set", set);
      var filename;
      var i;
      for (i = f.length - 1; i >= 0; i--) {
        filename = f[i];
        if (filename) break;
      }
      for (i = 0; i < set.length; i++) {
        var pattern = set[i];
        var file = f;
        if (options.matchBase && pattern.length === 1) {
          file = [filename];
        }
        var hit = this.matchOne(file, pattern, partial);
        if (hit) {
          if (options.flipNegate) return true;
          return !this.negate;
        }
      }
      if (options.flipNegate) return false;
      return this.negate;
    };
    Minimatch.prototype.matchOne = function(file, pattern, partial) {
      var options = this.options;
      this.debug(
        "matchOne",
        { "this": this, file, pattern }
      );
      this.debug("matchOne", file.length, pattern.length);
      for (var fi = 0, pi = 0, fl = file.length, pl = pattern.length; fi < fl && pi < pl; fi++, pi++) {
        this.debug("matchOne loop");
        var p = pattern[pi];
        var f = file[fi];
        this.debug(pattern, p, f);
        if (p === false) return false;
        if (p === GLOBSTAR) {
          this.debug("GLOBSTAR", [pattern, p, f]);
          var fr = fi;
          var pr = pi + 1;
          if (pr === pl) {
            this.debug("** at the end");
            for (; fi < fl; fi++) {
              if (file[fi] === "." || file[fi] === ".." || !options.dot && file[fi].charAt(0) === ".") return false;
            }
            return true;
          }
          while (fr < fl) {
            var swallowee = file[fr];
            this.debug("\nglobstar while", file, fr, pattern, pr, swallowee);
            if (this.matchOne(file.slice(fr), pattern.slice(pr), partial)) {
              this.debug("globstar found match!", fr, fl, swallowee);
              return true;
            } else {
              if (swallowee === "." || swallowee === ".." || !options.dot && swallowee.charAt(0) === ".") {
                this.debug("dot detected!", file, fr, pattern, pr);
                break;
              }
              this.debug("globstar swallow a segment, and continue");
              fr++;
            }
          }
          if (partial) {
            this.debug("\n>>> no match, partial?", file, fr, pattern, pr);
            if (fr === fl) return true;
          }
          return false;
        }
        var hit;
        if (typeof p === "string") {
          hit = f === p;
          this.debug("string match", p, f, hit);
        } else {
          hit = f.match(p);
          this.debug("pattern match", p, f, hit);
        }
        if (!hit) return false;
      }
      if (fi === fl && pi === pl) {
        return true;
      } else if (fi === fl) {
        return partial;
      } else if (pi === pl) {
        return fi === fl - 1 && file[fi] === "";
      }
      throw new Error("wtf?");
    };
    function globUnescape(s) {
      return s.replace(/\\(.)/g, "$1");
    }
    function regExpEscape(s) {
      return s.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, "\\$&");
    }
  }
});

// node_modules/vscode-languageclient/lib/common/fileOperations.js
var require_fileOperations = __commonJS({
  "node_modules/vscode-languageclient/lib/common/fileOperations.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.WillDeleteFilesFeature = exports2.WillRenameFilesFeature = exports2.WillCreateFilesFeature = exports2.DidDeleteFilesFeature = exports2.DidRenameFilesFeature = exports2.DidCreateFilesFeature = void 0;
    var code = require("vscode");
    var minimatch = require_minimatch();
    var proto = require_main3();
    var UUID = require_uuid();
    function ensure(target, key) {
      if (target[key] === void 0) {
        target[key] = {};
      }
      return target[key];
    }
    function access(target, key) {
      return target[key];
    }
    function assign(target, key, value) {
      target[key] = value;
    }
    var FileOperationFeature = class _FileOperationFeature {
      constructor(client2, event, registrationType, clientCapability, serverCapability) {
        this._filters = /* @__PURE__ */ new Map();
        this._client = client2;
        this._event = event;
        this._registrationType = registrationType;
        this._clientCapability = clientCapability;
        this._serverCapability = serverCapability;
      }
      get registrationType() {
        return this._registrationType;
      }
      fillClientCapabilities(capabilities) {
        const value = ensure(ensure(capabilities, "workspace"), "fileOperations");
        assign(value, "dynamicRegistration", true);
        assign(value, this._clientCapability, true);
      }
      initialize(capabilities) {
        var _a;
        const options = (_a = capabilities.workspace) === null || _a === void 0 ? void 0 : _a.fileOperations;
        const capability = options !== void 0 ? access(options, this._serverCapability) : void 0;
        if ((capability === null || capability === void 0 ? void 0 : capability.filters) !== void 0) {
          try {
            this.register({
              id: UUID.generateUuid(),
              registerOptions: { filters: capability.filters }
            });
          } catch (e) {
            this._client.warn(`Ignoring invalid glob pattern for ${this._serverCapability} registration: ${e}`);
          }
        }
      }
      register(data) {
        if (!this._listener) {
          this._listener = this._event(this.send, this);
        }
        const minimatchFilter = data.registerOptions.filters.map((filter) => {
          const matcher = new minimatch.Minimatch(filter.pattern.glob, _FileOperationFeature.asMinimatchOptions(filter.pattern.options));
          if (!matcher.makeRe()) {
            throw new Error(`Invalid pattern ${filter.pattern.glob}!`);
          }
          return { scheme: filter.scheme, matcher, kind: filter.pattern.matches };
        });
        this._filters.set(data.id, minimatchFilter);
      }
      unregister(id) {
        this._filters.delete(id);
        if (this._filters.size === 0 && this._listener) {
          this._listener.dispose();
          this._listener = void 0;
        }
      }
      dispose() {
        this._filters.clear();
        if (this._listener) {
          this._listener.dispose();
          this._listener = void 0;
        }
      }
      async filter(event, prop) {
        const fileMatches = await Promise.all(event.files.map(async (item) => {
          const uri = prop(item);
          const path2 = uri.fsPath.replace(/\\/g, "/");
          for (const filters of this._filters.values()) {
            for (const filter of filters) {
              if (filter.scheme !== void 0 && filter.scheme !== uri.scheme) {
                continue;
              }
              if (filter.matcher.match(path2)) {
                if (filter.kind === void 0) {
                  return true;
                }
                const fileType = await _FileOperationFeature.getFileType(uri);
                if (fileType === void 0) {
                  this._client.error(`Failed to determine file type for ${uri.toString()}.`);
                  return true;
                }
                if (fileType === code.FileType.File && filter.kind === proto.FileOperationPatternKind.file || fileType === code.FileType.Directory && filter.kind === proto.FileOperationPatternKind.folder) {
                  return true;
                }
              } else if (filter.kind === proto.FileOperationPatternKind.folder) {
                const fileType = await _FileOperationFeature.getFileType(uri);
                if (fileType === code.FileType.Directory && filter.matcher.match(`${path2}/`)) {
                  return true;
                }
              }
            }
          }
          return false;
        }));
        const files = event.files.filter((_, index) => fileMatches[index]);
        return Object.assign(Object.assign({}, event), { files });
      }
      static async getFileType(uri) {
        try {
          return (await code.workspace.fs.stat(uri)).type;
        } catch (e) {
          return void 0;
        }
      }
      static asMinimatchOptions(options) {
        if (options === void 0) {
          return void 0;
        }
        if (options.ignoreCase === true) {
          return { nocase: true };
        }
        return void 0;
      }
    };
    var NotificationFileOperationFeature = class extends FileOperationFeature {
      constructor(client2, event, notificationType, clientCapability, serverCapability, accessUri, createParams) {
        super(client2, event, notificationType, clientCapability, serverCapability);
        this._notificationType = notificationType;
        this._accessUri = accessUri;
        this._createParams = createParams;
      }
      async send(originalEvent) {
        const filteredEvent = await this.filter(originalEvent, this._accessUri);
        if (filteredEvent.files.length) {
          const next = async (event) => {
            this._client.sendNotification(this._notificationType, this._createParams(event));
          };
          this.doSend(filteredEvent, next);
        }
      }
    };
    var DidCreateFilesFeature = class extends NotificationFileOperationFeature {
      constructor(client2) {
        super(client2, code.workspace.onDidCreateFiles, proto.DidCreateFilesNotification.type, "didCreate", "didCreate", (i) => i, client2.code2ProtocolConverter.asDidCreateFilesParams);
      }
      doSend(event, next) {
        var _a;
        const middleware = (_a = this._client.clientOptions.middleware) === null || _a === void 0 ? void 0 : _a.workspace;
        return (middleware === null || middleware === void 0 ? void 0 : middleware.didCreateFiles) ? middleware.didCreateFiles(event, next) : next(event);
      }
    };
    exports2.DidCreateFilesFeature = DidCreateFilesFeature;
    var DidRenameFilesFeature = class extends NotificationFileOperationFeature {
      constructor(client2) {
        super(client2, code.workspace.onDidRenameFiles, proto.DidRenameFilesNotification.type, "didRename", "didRename", (i) => i.oldUri, client2.code2ProtocolConverter.asDidRenameFilesParams);
      }
      doSend(event, next) {
        var _a;
        const middleware = (_a = this._client.clientOptions.middleware) === null || _a === void 0 ? void 0 : _a.workspace;
        return (middleware === null || middleware === void 0 ? void 0 : middleware.didRenameFiles) ? middleware.didRenameFiles(event, next) : next(event);
      }
    };
    exports2.DidRenameFilesFeature = DidRenameFilesFeature;
    var DidDeleteFilesFeature = class extends NotificationFileOperationFeature {
      constructor(client2) {
        super(client2, code.workspace.onDidDeleteFiles, proto.DidDeleteFilesNotification.type, "didDelete", "didDelete", (i) => i, client2.code2ProtocolConverter.asDidDeleteFilesParams);
      }
      doSend(event, next) {
        var _a;
        const middleware = (_a = this._client.clientOptions.middleware) === null || _a === void 0 ? void 0 : _a.workspace;
        return (middleware === null || middleware === void 0 ? void 0 : middleware.didDeleteFiles) ? middleware.didDeleteFiles(event, next) : next(event);
      }
    };
    exports2.DidDeleteFilesFeature = DidDeleteFilesFeature;
    var RequestFileOperationFeature = class extends FileOperationFeature {
      constructor(client2, event, requestType, clientCapability, serverCapability, accessUri, createParams) {
        super(client2, event, requestType, clientCapability, serverCapability);
        this._requestType = requestType;
        this._accessUri = accessUri;
        this._createParams = createParams;
      }
      async send(originalEvent) {
        const waitUntil = this.waitUntil(originalEvent);
        originalEvent.waitUntil(waitUntil);
      }
      async waitUntil(originalEvent) {
        const filteredEvent = await this.filter(originalEvent, this._accessUri);
        if (filteredEvent.files.length) {
          const next = (event) => {
            return this._client.sendRequest(this._requestType, this._createParams(event)).then(this._client.protocol2CodeConverter.asWorkspaceEdit);
          };
          return this.doSend(filteredEvent, next);
        } else {
          return void 0;
        }
      }
    };
    var WillCreateFilesFeature = class extends RequestFileOperationFeature {
      constructor(client2) {
        super(client2, code.workspace.onWillCreateFiles, proto.WillCreateFilesRequest.type, "willCreate", "willCreate", (i) => i, client2.code2ProtocolConverter.asWillCreateFilesParams);
      }
      doSend(event, next) {
        var _a;
        const middleware = (_a = this._client.clientOptions.middleware) === null || _a === void 0 ? void 0 : _a.workspace;
        return (middleware === null || middleware === void 0 ? void 0 : middleware.willCreateFiles) ? middleware.willCreateFiles(event, next) : next(event);
      }
    };
    exports2.WillCreateFilesFeature = WillCreateFilesFeature;
    var WillRenameFilesFeature = class extends RequestFileOperationFeature {
      constructor(client2) {
        super(client2, code.workspace.onWillRenameFiles, proto.WillRenameFilesRequest.type, "willRename", "willRename", (i) => i.oldUri, client2.code2ProtocolConverter.asWillRenameFilesParams);
      }
      doSend(event, next) {
        var _a;
        const middleware = (_a = this._client.clientOptions.middleware) === null || _a === void 0 ? void 0 : _a.workspace;
        return (middleware === null || middleware === void 0 ? void 0 : middleware.willRenameFiles) ? middleware.willRenameFiles(event, next) : next(event);
      }
    };
    exports2.WillRenameFilesFeature = WillRenameFilesFeature;
    var WillDeleteFilesFeature = class extends RequestFileOperationFeature {
      constructor(client2) {
        super(client2, code.workspace.onWillDeleteFiles, proto.WillDeleteFilesRequest.type, "willDelete", "willDelete", (i) => i, client2.code2ProtocolConverter.asWillDeleteFilesParams);
      }
      doSend(event, next) {
        var _a;
        const middleware = (_a = this._client.clientOptions.middleware) === null || _a === void 0 ? void 0 : _a.workspace;
        return (middleware === null || middleware === void 0 ? void 0 : middleware.willDeleteFiles) ? middleware.willDeleteFiles(event, next) : next(event);
      }
    };
    exports2.WillDeleteFilesFeature = WillDeleteFilesFeature;
  }
});

// node_modules/vscode-languageclient/lib/common/linkedEditingRange.js
var require_linkedEditingRange = __commonJS({
  "node_modules/vscode-languageclient/lib/common/linkedEditingRange.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.LinkedEditingFeature = void 0;
    var code = require("vscode");
    var proto = require_main3();
    var client_1 = require_client();
    function ensure(target, key) {
      if (target[key] === void 0) {
        target[key] = {};
      }
      return target[key];
    }
    var LinkedEditingFeature = class extends client_1.TextDocumentFeature {
      constructor(client2) {
        super(client2, proto.LinkedEditingRangeRequest.type);
      }
      fillClientCapabilities(capabilities) {
        const linkedEditingSupport = ensure(ensure(capabilities, "textDocument"), "linkedEditingRange");
        linkedEditingSupport.dynamicRegistration = true;
      }
      initialize(capabilities, documentSelector) {
        let [id, options] = this.getRegistration(documentSelector, capabilities.linkedEditingRangeProvider);
        if (!id || !options) {
          return;
        }
        this.register({ id, registerOptions: options });
      }
      registerLanguageProvider(options) {
        const provider = {
          provideLinkedEditingRanges: (document, position, token) => {
            const client2 = this._client;
            const provideLinkedEditing = (document2, position2, token2) => {
              return client2.sendRequest(proto.LinkedEditingRangeRequest.type, client2.code2ProtocolConverter.asTextDocumentPositionParams(document2, position2), token2).then(client2.protocol2CodeConverter.asLinkedEditingRanges, (error) => {
                return client2.handleFailedRequest(proto.LinkedEditingRangeRequest.type, error, null);
              });
            };
            const middleware = client2.clientOptions.middleware;
            return middleware.provideLinkedEditingRange ? middleware.provideLinkedEditingRange(document, position, token, provideLinkedEditing) : provideLinkedEditing(document, position, token);
          }
        };
        return [code.languages.registerLinkedEditingRangeProvider(options.documentSelector, provider), provider];
      }
    };
    exports2.LinkedEditingFeature = LinkedEditingFeature;
  }
});

// node_modules/vscode-languageclient/lib/common/commonClient.js
var require_commonClient = __commonJS({
  "node_modules/vscode-languageclient/lib/common/commonClient.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.ProposedFeatures = exports2.CommonLanguageClient = void 0;
    var client_1 = require_client();
    var colorProvider_1 = require_colorProvider();
    var configuration_1 = require_configuration();
    var implementation_1 = require_implementation();
    var typeDefinition_1 = require_typeDefinition();
    var workspaceFolders_1 = require_workspaceFolders();
    var foldingRange_1 = require_foldingRange();
    var declaration_1 = require_declaration();
    var selectionRange_1 = require_selectionRange();
    var progress_1 = require_progress();
    var callHierarchy_1 = require_callHierarchy();
    var semanticTokens_1 = require_semanticTokens();
    var fileOperations_1 = require_fileOperations();
    var linkedEditingRange_1 = require_linkedEditingRange();
    var CommonLanguageClient = class extends client_1.BaseLanguageClient {
      constructor(id, name, clientOptions) {
        super(id, name, clientOptions);
      }
      registerProposedFeatures() {
        this.registerFeatures(ProposedFeatures.createAll(this));
      }
      registerBuiltinFeatures() {
        super.registerBuiltinFeatures();
        this.registerFeature(new configuration_1.ConfigurationFeature(this));
        this.registerFeature(new typeDefinition_1.TypeDefinitionFeature(this));
        this.registerFeature(new implementation_1.ImplementationFeature(this));
        this.registerFeature(new colorProvider_1.ColorProviderFeature(this));
        this.registerFeature(new workspaceFolders_1.WorkspaceFoldersFeature(this));
        this.registerFeature(new foldingRange_1.FoldingRangeFeature(this));
        this.registerFeature(new declaration_1.DeclarationFeature(this));
        this.registerFeature(new selectionRange_1.SelectionRangeFeature(this));
        this.registerFeature(new progress_1.ProgressFeature(this));
        this.registerFeature(new callHierarchy_1.CallHierarchyFeature(this));
        this.registerFeature(new semanticTokens_1.SemanticTokensFeature(this));
        this.registerFeature(new linkedEditingRange_1.LinkedEditingFeature(this));
        this.registerFeature(new fileOperations_1.DidCreateFilesFeature(this));
        this.registerFeature(new fileOperations_1.DidRenameFilesFeature(this));
        this.registerFeature(new fileOperations_1.DidDeleteFilesFeature(this));
        this.registerFeature(new fileOperations_1.WillCreateFilesFeature(this));
        this.registerFeature(new fileOperations_1.WillRenameFilesFeature(this));
        this.registerFeature(new fileOperations_1.WillDeleteFilesFeature(this));
      }
    };
    exports2.CommonLanguageClient = CommonLanguageClient;
    var ProposedFeatures;
    (function(ProposedFeatures2) {
      function createAll(_client) {
        let result = [];
        return result;
      }
      ProposedFeatures2.createAll = createAll;
    })(ProposedFeatures = exports2.ProposedFeatures || (exports2.ProposedFeatures = {}));
  }
});

// node_modules/vscode-languageclient/lib/node/processes.js
var require_processes = __commonJS({
  "node_modules/vscode-languageclient/lib/node/processes.js"(exports2) {
    "use strict";
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.terminate = void 0;
    var cp = require("child_process");
    var path_1 = require("path");
    var isWindows = process.platform === "win32";
    var isMacintosh = process.platform === "darwin";
    var isLinux = process.platform === "linux";
    function terminate(process2, cwd) {
      if (isWindows) {
        try {
          let options = {
            stdio: ["pipe", "pipe", "ignore"]
          };
          if (cwd) {
            options.cwd = cwd;
          }
          cp.execFileSync("taskkill", ["/T", "/F", "/PID", process2.pid.toString()], options);
          return true;
        } catch (err) {
          return false;
        }
      } else if (isLinux || isMacintosh) {
        try {
          var cmd = path_1.join(__dirname, "terminateProcess.sh");
          var result = cp.spawnSync(cmd, [process2.pid.toString()]);
          return result.error ? false : true;
        } catch (err) {
          return false;
        }
      } else {
        process2.kill("SIGKILL");
        return true;
      }
    }
    exports2.terminate = terminate;
  }
});

// node_modules/vscode-languageserver-protocol/node.js
var require_node2 = __commonJS({
  "node_modules/vscode-languageserver-protocol/node.js"(exports2, module2) {
    "use strict";
    module2.exports = require_main3();
  }
});

// node_modules/vscode-languageclient/lib/common/api.js
var require_api3 = __commonJS({
  "node_modules/vscode-languageclient/lib/common/api.js"(exports2) {
    "use strict";
    var __createBinding = exports2 && exports2.__createBinding || (Object.create ? function(o, m, k, k2) {
      if (k2 === void 0) k2 = k;
      Object.defineProperty(o, k2, { enumerable: true, get: function() {
        return m[k];
      } });
    } : function(o, m, k, k2) {
      if (k2 === void 0) k2 = k;
      o[k2] = m[k];
    });
    var __exportStar = exports2 && exports2.__exportStar || function(m, exports3) {
      for (var p in m) if (p !== "default" && !Object.prototype.hasOwnProperty.call(exports3, p)) __createBinding(exports3, m, p);
    };
    Object.defineProperty(exports2, "__esModule", { value: true });
    __exportStar(require_main3(), exports2);
    __exportStar(require_client(), exports2);
    __exportStar(require_commonClient(), exports2);
  }
});

// node_modules/vscode-languageclient/lib/node/main.js
var require_main4 = __commonJS({
  "node_modules/vscode-languageclient/lib/node/main.js"(exports2) {
    "use strict";
    var __createBinding = exports2 && exports2.__createBinding || (Object.create ? function(o, m, k, k2) {
      if (k2 === void 0) k2 = k;
      Object.defineProperty(o, k2, { enumerable: true, get: function() {
        return m[k];
      } });
    } : function(o, m, k, k2) {
      if (k2 === void 0) k2 = k;
      o[k2] = m[k];
    });
    var __exportStar = exports2 && exports2.__exportStar || function(m, exports3) {
      for (var p in m) if (p !== "default" && !Object.prototype.hasOwnProperty.call(exports3, p)) __createBinding(exports3, m, p);
    };
    Object.defineProperty(exports2, "__esModule", { value: true });
    exports2.SettingMonitor = exports2.LanguageClient = exports2.TransportKind = void 0;
    var cp = require("child_process");
    var fs = require("fs");
    var path2 = require("path");
    var SemVer = require_semver2();
    var vscode_1 = require("vscode");
    var Is = require_is();
    var commonClient_1 = require_commonClient();
    var client_1 = require_client();
    var processes_1 = require_processes();
    var node_1 = require_node2();
    __exportStar(require_node2(), exports2);
    __exportStar(require_api3(), exports2);
    var REQUIRED_VSCODE_VERSION = "^1.52.0";
    var Executable;
    (function(Executable2) {
      function is(value) {
        return Is.string(value.command);
      }
      Executable2.is = is;
    })(Executable || (Executable = {}));
    var TransportKind2;
    (function(TransportKind3) {
      TransportKind3[TransportKind3["stdio"] = 0] = "stdio";
      TransportKind3[TransportKind3["ipc"] = 1] = "ipc";
      TransportKind3[TransportKind3["pipe"] = 2] = "pipe";
      TransportKind3[TransportKind3["socket"] = 3] = "socket";
    })(TransportKind2 = exports2.TransportKind || (exports2.TransportKind = {}));
    var Transport;
    (function(Transport2) {
      function isSocket(value) {
        let candidate = value;
        return candidate && candidate.kind === TransportKind2.socket && Is.number(candidate.port);
      }
      Transport2.isSocket = isSocket;
    })(Transport || (Transport = {}));
    var NodeModule;
    (function(NodeModule2) {
      function is(value) {
        return Is.string(value.module);
      }
      NodeModule2.is = is;
    })(NodeModule || (NodeModule = {}));
    var StreamInfo;
    (function(StreamInfo2) {
      function is(value) {
        let candidate = value;
        return candidate && candidate.writer !== void 0 && candidate.reader !== void 0;
      }
      StreamInfo2.is = is;
    })(StreamInfo || (StreamInfo = {}));
    var ChildProcessInfo;
    (function(ChildProcessInfo2) {
      function is(value) {
        let candidate = value;
        return candidate && candidate.process !== void 0 && typeof candidate.detached === "boolean";
      }
      ChildProcessInfo2.is = is;
    })(ChildProcessInfo || (ChildProcessInfo = {}));
    var LanguageClient2 = class extends commonClient_1.CommonLanguageClient {
      constructor(arg1, arg2, arg3, arg4, arg5) {
        let id;
        let name;
        let serverOptions;
        let clientOptions;
        let forceDebug;
        if (Is.string(arg2)) {
          id = arg1;
          name = arg2;
          serverOptions = arg3;
          clientOptions = arg4;
          forceDebug = !!arg5;
        } else {
          id = arg1.toLowerCase();
          name = arg1;
          serverOptions = arg2;
          clientOptions = arg3;
          forceDebug = arg4;
        }
        if (forceDebug === void 0) {
          forceDebug = false;
        }
        super(id, name, clientOptions);
        this._serverOptions = serverOptions;
        this._forceDebug = forceDebug;
        try {
          this.checkVersion();
        } catch (error) {
          if (Is.string(error.message)) {
            this.outputChannel.appendLine(error.message);
          }
          throw error;
        }
      }
      checkVersion() {
        let codeVersion = SemVer.parse(vscode_1.version);
        if (!codeVersion) {
          throw new Error(`No valid VS Code version detected. Version string is: ${vscode_1.version}`);
        }
        if (codeVersion.prerelease && codeVersion.prerelease.length > 0) {
          codeVersion.prerelease = [];
        }
        if (!SemVer.satisfies(codeVersion, REQUIRED_VSCODE_VERSION)) {
          throw new Error(`The language client requires VS Code version ${REQUIRED_VSCODE_VERSION} but received version ${vscode_1.version}`);
        }
      }
      stop() {
        return super.stop().then(() => {
          if (this._serverProcess) {
            let toCheck = this._serverProcess;
            this._serverProcess = void 0;
            if (this._isDetached === void 0 || !this._isDetached) {
              this.checkProcessDied(toCheck);
            }
            this._isDetached = void 0;
          }
        });
      }
      checkProcessDied(childProcess) {
        if (!childProcess) {
          return;
        }
        setTimeout(() => {
          try {
            process.kill(childProcess.pid, 0);
            processes_1.terminate(childProcess);
          } catch (error) {
          }
        }, 2e3);
      }
      handleConnectionClosed() {
        this._serverProcess = void 0;
        super.handleConnectionClosed();
      }
      fillInitializeParams(params) {
        super.fillInitializeParams(params);
        if (params.processId === null) {
          params.processId = process.pid;
        }
      }
      createMessageTransports(encoding) {
        function getEnvironment(env, fork) {
          if (!env && !fork) {
            return void 0;
          }
          let result = /* @__PURE__ */ Object.create(null);
          Object.keys(process.env).forEach((key) => result[key] = process.env[key]);
          if (fork) {
            result["ELECTRON_RUN_AS_NODE"] = "1";
            result["ELECTRON_NO_ASAR"] = "1";
          }
          if (env) {
            Object.keys(env).forEach((key) => result[key] = env[key]);
          }
          return result;
        }
        const debugStartWith = ["--debug=", "--debug-brk=", "--inspect=", "--inspect-brk="];
        const debugEquals = ["--debug", "--debug-brk", "--inspect", "--inspect-brk"];
        function startedInDebugMode() {
          let args = process.execArgv;
          if (args) {
            return args.some((arg) => {
              return debugStartWith.some((value) => arg.startsWith(value)) || debugEquals.some((value) => arg === value);
            });
          }
          return false;
        }
        function assertStdio(process2) {
          if (process2.stdin === null || process2.stdout === null || process2.stderr === null) {
            throw new Error("Process created without stdio streams");
          }
        }
        let server = this._serverOptions;
        if (Is.func(server)) {
          return server().then((result) => {
            if (client_1.MessageTransports.is(result)) {
              this._isDetached = !!result.detached;
              return result;
            } else if (StreamInfo.is(result)) {
              this._isDetached = !!result.detached;
              return { reader: new node_1.StreamMessageReader(result.reader), writer: new node_1.StreamMessageWriter(result.writer) };
            } else {
              let cp2;
              if (ChildProcessInfo.is(result)) {
                cp2 = result.process;
                this._isDetached = result.detached;
              } else {
                cp2 = result;
                this._isDetached = false;
              }
              cp2.stderr.on("data", (data) => this.outputChannel.append(Is.string(data) ? data : data.toString(encoding)));
              return { reader: new node_1.StreamMessageReader(cp2.stdout), writer: new node_1.StreamMessageWriter(cp2.stdin) };
            }
          });
        }
        let json;
        let runDebug = server;
        if (runDebug.run || runDebug.debug) {
          if (this._forceDebug || startedInDebugMode()) {
            json = runDebug.debug;
          } else {
            json = runDebug.run;
          }
        } else {
          json = server;
        }
        return this._getServerWorkingDir(json.options).then((serverWorkingDir) => {
          if (NodeModule.is(json) && json.module) {
            let node = json;
            let transport = node.transport || TransportKind2.stdio;
            if (node.runtime) {
              let args = [];
              let options = node.options || /* @__PURE__ */ Object.create(null);
              if (options.execArgv) {
                options.execArgv.forEach((element) => args.push(element));
              }
              args.push(node.module);
              if (node.args) {
                node.args.forEach((element) => args.push(element));
              }
              const execOptions = /* @__PURE__ */ Object.create(null);
              execOptions.cwd = serverWorkingDir;
              execOptions.env = getEnvironment(options.env, false);
              const runtime = this._getRuntimePath(node.runtime, serverWorkingDir);
              let pipeName = void 0;
              if (transport === TransportKind2.ipc) {
                execOptions.stdio = [null, null, null, "ipc"];
                args.push("--node-ipc");
              } else if (transport === TransportKind2.stdio) {
                args.push("--stdio");
              } else if (transport === TransportKind2.pipe) {
                pipeName = node_1.generateRandomPipeName();
                args.push(`--pipe=${pipeName}`);
              } else if (Transport.isSocket(transport)) {
                args.push(`--socket=${transport.port}`);
              }
              args.push(`--clientProcessId=${process.pid.toString()}`);
              if (transport === TransportKind2.ipc || transport === TransportKind2.stdio) {
                let serverProcess = cp.spawn(runtime, args, execOptions);
                if (!serverProcess || !serverProcess.pid) {
                  return Promise.reject(`Launching server using runtime ${runtime} failed.`);
                }
                this._serverProcess = serverProcess;
                serverProcess.stderr.on("data", (data) => this.outputChannel.append(Is.string(data) ? data : data.toString(encoding)));
                if (transport === TransportKind2.ipc) {
                  serverProcess.stdout.on("data", (data) => this.outputChannel.append(Is.string(data) ? data : data.toString(encoding)));
                  return Promise.resolve({ reader: new node_1.IPCMessageReader(serverProcess), writer: new node_1.IPCMessageWriter(serverProcess) });
                } else {
                  return Promise.resolve({ reader: new node_1.StreamMessageReader(serverProcess.stdout), writer: new node_1.StreamMessageWriter(serverProcess.stdin) });
                }
              } else if (transport === TransportKind2.pipe) {
                return node_1.createClientPipeTransport(pipeName).then((transport2) => {
                  let process2 = cp.spawn(runtime, args, execOptions);
                  if (!process2 || !process2.pid) {
                    return Promise.reject(`Launching server using runtime ${runtime} failed.`);
                  }
                  this._serverProcess = process2;
                  process2.stderr.on("data", (data) => this.outputChannel.append(Is.string(data) ? data : data.toString(encoding)));
                  process2.stdout.on("data", (data) => this.outputChannel.append(Is.string(data) ? data : data.toString(encoding)));
                  return transport2.onConnected().then((protocol) => {
                    return { reader: protocol[0], writer: protocol[1] };
                  });
                });
              } else if (Transport.isSocket(transport)) {
                return node_1.createClientSocketTransport(transport.port).then((transport2) => {
                  let process2 = cp.spawn(runtime, args, execOptions);
                  if (!process2 || !process2.pid) {
                    return Promise.reject(`Launching server using runtime ${runtime} failed.`);
                  }
                  this._serverProcess = process2;
                  process2.stderr.on("data", (data) => this.outputChannel.append(Is.string(data) ? data : data.toString(encoding)));
                  process2.stdout.on("data", (data) => this.outputChannel.append(Is.string(data) ? data : data.toString(encoding)));
                  return transport2.onConnected().then((protocol) => {
                    return { reader: protocol[0], writer: protocol[1] };
                  });
                });
              }
            } else {
              let pipeName = void 0;
              return new Promise((resolve, _reject) => {
                let args = node.args && node.args.slice() || [];
                if (transport === TransportKind2.ipc) {
                  args.push("--node-ipc");
                } else if (transport === TransportKind2.stdio) {
                  args.push("--stdio");
                } else if (transport === TransportKind2.pipe) {
                  pipeName = node_1.generateRandomPipeName();
                  args.push(`--pipe=${pipeName}`);
                } else if (Transport.isSocket(transport)) {
                  args.push(`--socket=${transport.port}`);
                }
                args.push(`--clientProcessId=${process.pid.toString()}`);
                let options = node.options || /* @__PURE__ */ Object.create(null);
                options.env = getEnvironment(options.env, true);
                options.execArgv = options.execArgv || [];
                options.cwd = serverWorkingDir;
                options.silent = true;
                if (transport === TransportKind2.ipc || transport === TransportKind2.stdio) {
                  let sp = cp.fork(node.module, args || [], options);
                  assertStdio(sp);
                  this._serverProcess = sp;
                  sp.stderr.on("data", (data) => this.outputChannel.append(Is.string(data) ? data : data.toString(encoding)));
                  if (transport === TransportKind2.ipc) {
                    sp.stdout.on("data", (data) => this.outputChannel.append(Is.string(data) ? data : data.toString(encoding)));
                    resolve({ reader: new node_1.IPCMessageReader(this._serverProcess), writer: new node_1.IPCMessageWriter(this._serverProcess) });
                  } else {
                    resolve({ reader: new node_1.StreamMessageReader(sp.stdout), writer: new node_1.StreamMessageWriter(sp.stdin) });
                  }
                } else if (transport === TransportKind2.pipe) {
                  node_1.createClientPipeTransport(pipeName).then((transport2) => {
                    let sp = cp.fork(node.module, args || [], options);
                    assertStdio(sp);
                    this._serverProcess = sp;
                    sp.stderr.on("data", (data) => this.outputChannel.append(Is.string(data) ? data : data.toString(encoding)));
                    sp.stdout.on("data", (data) => this.outputChannel.append(Is.string(data) ? data : data.toString(encoding)));
                    transport2.onConnected().then((protocol) => {
                      resolve({ reader: protocol[0], writer: protocol[1] });
                    });
                  });
                } else if (Transport.isSocket(transport)) {
                  node_1.createClientSocketTransport(transport.port).then((transport2) => {
                    let sp = cp.fork(node.module, args || [], options);
                    assertStdio(sp);
                    this._serverProcess = sp;
                    sp.stderr.on("data", (data) => this.outputChannel.append(Is.string(data) ? data : data.toString(encoding)));
                    sp.stdout.on("data", (data) => this.outputChannel.append(Is.string(data) ? data : data.toString(encoding)));
                    transport2.onConnected().then((protocol) => {
                      resolve({ reader: protocol[0], writer: protocol[1] });
                    });
                  });
                }
              });
            }
          } else if (Executable.is(json) && json.command) {
            let command = json;
            let args = command.args || [];
            let options = Object.assign({}, command.options);
            options.cwd = options.cwd || serverWorkingDir;
            let serverProcess = cp.spawn(command.command, args, options);
            if (!serverProcess || !serverProcess.pid) {
              return Promise.reject(`Launching server using command ${command.command} failed.`);
            }
            serverProcess.stderr.on("data", (data) => this.outputChannel.append(Is.string(data) ? data : data.toString(encoding)));
            this._serverProcess = serverProcess;
            this._isDetached = !!options.detached;
            return Promise.resolve({ reader: new node_1.StreamMessageReader(serverProcess.stdout), writer: new node_1.StreamMessageWriter(serverProcess.stdin) });
          }
          return Promise.reject(new Error(`Unsupported server configuration ` + JSON.stringify(server, null, 4)));
        });
      }
      _getRuntimePath(runtime, serverWorkingDirectory) {
        if (path2.isAbsolute(runtime)) {
          return runtime;
        }
        const mainRootPath = this._mainGetRootPath();
        if (mainRootPath !== void 0) {
          const result = path2.join(mainRootPath, runtime);
          if (fs.existsSync(result)) {
            return result;
          }
        }
        if (serverWorkingDirectory !== void 0) {
          const result = path2.join(serverWorkingDirectory, runtime);
          if (fs.existsSync(result)) {
            return result;
          }
        }
        return runtime;
      }
      _mainGetRootPath() {
        let folders = vscode_1.workspace.workspaceFolders;
        if (!folders || folders.length === 0) {
          return void 0;
        }
        let folder = folders[0];
        if (folder.uri.scheme === "file") {
          return folder.uri.fsPath;
        }
        return void 0;
      }
      _getServerWorkingDir(options) {
        let cwd = options && options.cwd;
        if (!cwd) {
          cwd = this.clientOptions.workspaceFolder ? this.clientOptions.workspaceFolder.uri.fsPath : this._mainGetRootPath();
        }
        if (cwd) {
          return new Promise((s) => {
            fs.lstat(cwd, (err, stats) => {
              s(!err && stats.isDirectory() ? cwd : void 0);
            });
          });
        }
        return Promise.resolve(void 0);
      }
      getLocale() {
        const envValue = process.env["VSCODE_NLS_CONFIG"];
        if (envValue === void 0) {
          return "en";
        }
        let config = void 0;
        try {
          config = JSON.parse(envValue);
        } catch (err) {
        }
        if (config === void 0 || typeof config.locale !== "string") {
          return "en";
        }
        return config.locale;
      }
    };
    exports2.LanguageClient = LanguageClient2;
    var SettingMonitor = class {
      constructor(_client, _setting) {
        this._client = _client;
        this._setting = _setting;
        this._listeners = [];
      }
      start() {
        vscode_1.workspace.onDidChangeConfiguration(this.onDidChangeConfiguration, this, this._listeners);
        this.onDidChangeConfiguration();
        return new vscode_1.Disposable(() => {
          if (this._client.needsStop()) {
            this._client.stop();
          }
        });
      }
      onDidChangeConfiguration() {
        let index = this._setting.indexOf(".");
        let primary = index >= 0 ? this._setting.substr(0, index) : this._setting;
        let rest = index >= 0 ? this._setting.substr(index + 1) : void 0;
        let enabled = rest ? vscode_1.workspace.getConfiguration(primary).get(rest, false) : vscode_1.workspace.getConfiguration(primary);
        if (enabled && this._client.needsStart()) {
          this._client.start();
        } else if (!enabled && this._client.needsStop()) {
          this._client.stop();
        }
      }
    };
    exports2.SettingMonitor = SettingMonitor;
  }
});

// node_modules/vscode-languageclient/node.js
var require_node3 = __commonJS({
  "node_modules/vscode-languageclient/node.js"(exports2, module2) {
    "use strict";
    module2.exports = require_main4();
  }
});

// src/extension.ts
var extension_exports = {};
__export(extension_exports, {
  activate: () => activate,
  deactivate: () => deactivate
});
module.exports = __toCommonJS(extension_exports);
var path = __toESM(require("path"));
var import_vscode = require("vscode");
var import_node = __toESM(require_node3());
var client;
function activate(context) {
  console.log("Activating NeuroScript Language Client...");
  const serverModule = context.asAbsolutePath(
    path.join("server", "nslsp_executable")
    // Ensure 'nslsp_executable' is in 'vscode-neuroscript/server/'
    // Adjust if your executable is elsewhere relative to the extension.
  );
  console.log(`LSP Server Executable Path: ${serverModule}`);
  const serverOptions = {
    command: serverModule,
    transport: import_node.TransportKind.stdio
    // Use stdio for communication
  };
  const clientOptions = {
    documentSelector: [{ scheme: "file", language: "neuroscript" }],
    synchronize: {
      // fileEvents: workspace.createFileSystemWatcher('**/.clientrc') // Optional: if server needs to watch other files
    },
    // Capture stderr from the server for logging/debugging in an output channel
    outputChannel: import_vscode.window.createOutputChannel("NeuroScript LSP")
  };
  client = new import_node.LanguageClient(
    "neuroscriptLSP",
    // ID of the language client (must be unique)
    "NeuroScript Language Server",
    // Name of the language client (shows in Output channel dropdown)
    serverOptions,
    clientOptions
  );
  console.log("Starting NeuroScript Language Client...");
  context.subscriptions.push(client.start());
  client.onReady().then(() => {
    console.log("NeuroScript Language Client is ready and connected to the server!");
  }).catch((e) => {
    console.error("NeuroScript Language Client failed to become ready:", e);
    import_vscode.window.showErrorMessage('NeuroScript Server failed to initialize. Check the "NeuroScript LSP" Output channel.');
  });
}
function deactivate() {
  if (!client) {
    return void 0;
  }
  console.log("Deactivating NeuroScript Language Client...");
  return client.stop();
}
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  activate,
  deactivate
});
//# sourceMappingURL=extension.js.map
