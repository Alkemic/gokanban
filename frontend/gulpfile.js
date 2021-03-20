/* global require: false */
// const config = require("./config")
const gulp = require("gulp")
const less = require("gulp-less")
const concat = require("gulp-concat")
const autoprefixer = require("gulp-autoprefixer")
const del = require("del")
const LessPluginCleanCSS = require("less-plugin-clean-css")
const cleancss = new LessPluginCleanCSS({advanced: true})
const ngTemplates = require("gulp-ng-templates")
const sourcemaps = require("gulp-sourcemaps")
const gulpIf = require("gulp-if")
const babel = require("gulp-babel")

const destDir = "../static/"
const nodeDir = "./node_modules/"

const config = {
    styles: {
        src: [
            "styles/**/*.less"
        ],
        file: "kanban.css",
        dest: `${destDir}/styles/`
    },
    vendorStyles: {
        src: [
            `${nodeDir}/bootstrap/dist/css/bootstrap.css`,
            `${nodeDir}/angular-ui-bootstrap/dist/ui-bootstrap-csp.css`
        ],
        file: "vendor.css",
        dest: `${destDir}/styles/`
    },
    scripts: {
        src: [
            "scripts/**/*.js"
        ],
        file: "kanban.js",
        dest: `${destDir}/scripts/`
    },
    vendorScripts: {
        src: [
            `${nodeDir}/jquery/dist/jquery.js`,
            `${nodeDir}/angular/angular.js`,
            `${nodeDir}/angular-drag-and-drop-lists/angular-drag-and-drop-lists.js`,
            `${nodeDir}/angular-ui-bootstrap/dist/ui-bootstrap-tpls.js`,
            `${nodeDir}/angular-sanitize/angular-sanitize.js`
        ],
        file: "vendor.js",
        dest: `${destDir}/scripts/`
    },
    templates: {
        src: "./templates/*.html",
        out: "kanban.templates.js",
        dest: `${destDir}/scripts`,
        moduleName: "kanban.templates",
    },
    files: [
        {src: `${nodeDir}/bootstrap/fonts/glyphicons-halflings-regular.woff2`, dest: `${destDir}/fonts/`},
        {src: `${nodeDir}/bootstrap/fonts/glyphicons-halflings-regular.ttf`, dest: `${destDir}/fonts/`}
    ]
}

const production = typeof process.env.PRODUCTION !== "undefined" && process.env.PRODUCTION === "true"

const clean = () => del([
        config.scripts.dest,
        config.vendorScripts.dest,
        config.styles.dest
    ].concat(
        config.files.map(file => file.dest)
    ).filter((v, i, a)=> a.indexOf(v) === i), {force: true})

const templates = () => gulp
    .src(config.templates.src)
    .pipe(ngTemplates({
        filename: config.templates.out,
        module: config.templates.moduleName,
        path: (path, base) => path.replace(base+"/", ""),
    }))
    .pipe(gulp.dest(config.templates.dest))

const generateStylesProcessor = styles => () => gulp
    .src(styles.src)
    .pipe(gulpIf(!production, sourcemaps.init()))
    .pipe(less({
        plugins: production ? [cleancss] : [],
        paths: config.styles.paths
    }))
    .pipe(autoprefixer(styles.browsers))
    .pipe(concat(styles.file))
    .pipe(gulpIf(!production, sourcemaps.write()))
    .pipe(gulp.dest(styles.dest))

const generateScriptsTask = scripts => () => gulp.src(scripts.src)
    .pipe(gulpIf(!production, sourcemaps.init()))
    .pipe(concat(scripts.file))
    .pipe(gulpIf(production, babel({
        "presets": ["minify", {comments: false}],
        "plugins": ["angularjs-annotate"]
    })))
    .pipe(gulpIf(!production, sourcemaps.write()))
    .pipe(gulp.dest(scripts.dest))

const styles = generateStylesProcessor(config.styles)
styles.displayName = "styles"
const vendorStyles = generateStylesProcessor(config.vendorStyles)
vendorStyles.displayName = "vendor styles"

const scripts = generateScriptsTask(config.scripts)
scripts.displayName = "scripts"
const vendorScripts = generateScriptsTask(config.vendorScripts)
vendorScripts.displayName = "vendor scripts"

const copy = (cb) => {
    config.files.forEach(file => gulp.src(file.src).pipe(gulp.dest(file.dest)))
    cb()
}

const watch = () => {
    gulp.watch(config.styles.src, styles)
    gulp.watch(config.vendorScripts.src, vendorScripts)
    gulp.watch(config.scripts.src, scripts)
    gulp.watch(config.templates.src, templates)
    gulp.watch(config.files.map(el => el.src), copy)
}

const build = gulp.series(clean, gulp.parallel(vendorStyles, styles, vendorScripts, scripts, templates, copy))

exports.clean = clean
exports.styles = styles
exports.vendorStyles = vendorStyles
exports.templates = templates
exports.scripts = scripts
exports.vendorScripts = vendorScripts
exports.copy = copy
exports.watch = watch
exports.build = build
exports.default = build
