package graphics

/*
#cgo windows LDFLAGS: -lgdi32 -lopengl32
#cgo linux   LDFLAGS: -lX11 -lGL
#cgo darwin  LDFLAGS: -framework OpenGL

// ==========================================================================
// rgfw_nano — minimal window + GL context for cloud-game
// Derived from RGFW 2.0.0-dev (libpng license)
// Copyright (C) 2022-26 Riley Mabb (@ColleagueRiley)
// Altered: stripped to ~540 lines from ~16,400
// Original: https://github.com/ColleagueRiley/RGFW
// ==========================================================================

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// ---- platform detection ----
#if defined(_WIN32)
  #define RGFW_WINDOWS
  #include <windows.h>
#elif defined(__linux__) || defined(__FreeBSD__) || defined(__OpenBSD__)
  #define RGFW_X11
  #include <X11/Xlib.h>
  #include <X11/Xutil.h>
  #include <GL/glx.h>
#elif defined(__APPLE__)
  #define RGFW_MACOS
  #include <dlfcn.h>
  #include <OpenGL/OpenGL.h>
  #include <OpenGL/CGLTypes.h>
#endif

// ---- basic types ----
typedef unsigned char      u8;
typedef   signed char      i8;
typedef unsigned short     u16;
typedef   signed short     i16;
typedef unsigned int       u32;
typedef   signed int       i32;
typedef unsigned long long u64;
typedef   signed long long i64;
typedef u8                 RGFW_bool;
#define RGFW_TRUE  1
#define RGFW_FALSE 0

#define RGFW_ALLOC malloc
#define RGFW_FREE  free

// ---- window flags ----
typedef u32 RGFW_windowFlags;
#define RGFW_windowHide   (1u << 9)

// ---- GL types ----
typedef void (*RGFW_proc)(void);

typedef struct RGFW_glContext {
#ifdef RGFW_WINDOWS
    HGLRC ctx;
#elif defined(RGFW_MACOS)
    void* ctx;
#else
    GLXContext ctx;
#endif
} RGFW_glContext;

typedef enum { RGFW_glCore = 0, RGFW_glForwardCompatibility, RGFW_glCompatibility, RGFW_glES } RGFW_glProfile;
typedef enum { RGFW_glAccelerated = 0, RGFW_glSoftware } RGFW_glRenderer;
typedef enum { RGFW_glReleaseNone = 0 } RGFW_glReleaseBehavior;

typedef struct RGFW_glHints {
    i32 stencil, samples, stereo, auxBuffers, doubleBuffer;
    i32 red, green, blue, alpha, depth;
    i32 accumRed, accumGreen, accumBlue, accumAlpha;
    RGFW_bool sRGB, robustness, debug, noError;
    RGFW_glReleaseBehavior releaseBehavior;
    RGFW_glProfile profile;
    i32 major, minor;
    RGFW_glContext* share;
    RGFW_glRenderer renderer;
} RGFW_glHints;

typedef u32 RGFW_gfxContextType;
#define RGFW_gfxNativeOpenGL  (1u << 0)
#define RGFW_gfxOwnedByRGFW   (1u << 2)

// ---- window struct ----
#ifdef RGFW_WINDOWS
typedef struct RGFW_window_src {
    HWND     window;
    HDC      hdc;
    RGFW_glContext* ctx_native;
    RGFW_gfxContextType gfxType;
} RGFW_window_src;
#elif defined(RGFW_MACOS)
typedef struct RGFW_window_src {
    void*    window;
    void*    view;
    RGFW_glContext* ctx_native;
    RGFW_gfxContextType gfxType;
} RGFW_window_src;
#else
typedef struct RGFW_window_src {
    Window   window;
    Display* display;
    Colormap cmap;
    GLXFBConfig fbconfig;
    RGFW_glContext* ctx_native;
    RGFW_gfxContextType gfxType;
} RGFW_window_src;
#endif

typedef struct RGFW_window {
    RGFW_window_src src;
    i32 x, y, w, h;
} RGFW_window;

// ---- global hints ----
static RGFW_glHints RGFW_nano_hints = {0,0,0,0,1, 8,8,8,8,24, 0,0,0,0, 0,0,0,0, RGFW_glReleaseNone, RGFW_glCore, 1,0, NULL, RGFW_glAccelerated};

static void RGFW_resetGlobalHints_OpenGL(void) {
    RGFW_glHints d = {0,0,0,0,1, 8,8,8,8,24, 0,0,0,0, 0,0,0,0, RGFW_glReleaseNone, RGFW_glCore, 1,0, NULL, RGFW_glAccelerated};
    RGFW_nano_hints = d;
}
static void RGFW_setGlobalHints_OpenGL(RGFW_glHints* h) { RGFW_nano_hints = *h; }
static RGFW_glHints* RGFW_getGlobalHints_OpenGL(void) { return &RGFW_nano_hints; }

// ---- API prototypes ----
static i32               RGFW_init(const char* name, u8 flags);
static void              RGFW_deinit(void);
static RGFW_window*      RGFW_createWindow(const char* name, i32 x, i32 y, i32 w, i32 h, RGFW_windowFlags flags);
static void              RGFW_window_close(RGFW_window* win);
static void              RGFW_window_setName(RGFW_window* win, const char* name);
static void              RGFW_window_makeCurrentContext_OpenGL(RGFW_window* win);
static RGFW_proc         RGFW_getProcAddress_OpenGL(const char* name);
static RGFW_bool         RGFW_window_createContextPtr_OpenGL(RGFW_window* win, RGFW_glContext* ctx, RGFW_glHints* hints);
static RGFW_glContext*   RGFW_window_createContext_OpenGL(RGFW_window* win, RGFW_glHints* hints);

// ==========================================================================
// Platform implementations
// ==========================================================================

#ifdef RGFW_WINDOWS
// -------------------------------------------------------------------
//  Windows (Win32 + WGL)
// -------------------------------------------------------------------

static LRESULT CALLBACK RGFW_nano_WndProc(HWND hwnd, UINT msg, WPARAM wp, LPARAM lp) {
    return DefWindowProcW(hwnd, msg, wp, lp);
}

static i32 RGFW_init(const char* name, u8 flags) { (void)name; (void)flags; return 0; }
static void RGFW_deinit(void) {}

static RGFW_window* RGFW_createWindow(const char* name, i32 x, i32 y, i32 w, i32 h, RGFW_windowFlags flags) {
    (void)x; (void)y;
    RGFW_window* win = (RGFW_window*)RGFW_ALLOC(sizeof(RGFW_window));
    if (!win) return NULL;
    memset(win, 0, sizeof(*win));
    win->w = w; win->h = h;

    HINSTANCE hi = GetModuleHandleW(NULL);
    WNDCLASSW wc = {0};
    wc.lpfnWndProc   = RGFW_nano_WndProc;
    wc.hInstance     = hi;
    wc.lpszClassName = L"rgfw_nano_wc";
    RegisterClassW(&wc);

    DWORD style = WS_CLIPSIBLINGS | WS_CLIPCHILDREN;
    RECT rect = {0, 0, (LONG)w, (LONG)h};
    AdjustWindowRect(&rect, style, FALSE);

    win->src.window = CreateWindowExW(0, L"rgfw_nano_wc", L"",
        style, 0, 0, rect.right - rect.left, rect.bottom - rect.top,
        NULL, NULL, hi, NULL);
    if (!win->src.window) { RGFW_FREE(win); return NULL; }

    win->src.hdc = GetDC(win->src.window);
    return win;
}

static void RGFW_window_close(RGFW_window* win) {
    if (!win) return;
    if (win->src.ctx_native && (win->src.gfxType & RGFW_gfxOwnedByRGFW))
        RGFW_FREE(win->src.ctx_native);
    if (win->src.hdc) { ReleaseDC(win->src.window, win->src.hdc); win->src.hdc = NULL; }
    if (win->src.window) { DestroyWindow(win->src.window); win->src.window = NULL; }
    RGFW_FREE(win);
}

static void RGFW_window_setName(RGFW_window* win, const char* name) { (void)win; (void)name; }

static void RGFW_window_makeCurrentContext_OpenGL(RGFW_window* win) {
    if (win && win->src.hdc && win->src.ctx_native && win->src.ctx_native->ctx)
        wglMakeCurrent(win->src.hdc, win->src.ctx_native->ctx);
}

static RGFW_proc RGFW_getProcAddress_OpenGL(const char* name) {
    static HMODULE glModule;
    if (!glModule) glModule = LoadLibraryA("opengl32.dll");
    if (glModule) { RGFW_proc p = (RGFW_proc)GetProcAddress(glModule, name); if (p) return p; }
    return (RGFW_proc)wglGetProcAddress(name);
}

static RGFW_bool RGFW_window_createContextPtr_OpenGL(RGFW_window* win, RGFW_glContext* ctx, RGFW_glHints* hints) {
    (void)hints;
    if (!win || !win->src.hdc || !ctx) return RGFW_FALSE;

    PIXELFORMATDESCRIPTOR pfd = {0};
    pfd.nSize = sizeof(PIXELFORMATDESCRIPTOR);
    pfd.nVersion = 1;
    pfd.dwFlags = PFD_DRAW_TO_WINDOW | PFD_SUPPORT_OPENGL | PFD_DOUBLEBUFFER | PFD_GENERIC_FORMAT;
    pfd.iPixelType = PFD_TYPE_RGBA;
    pfd.cColorBits = 32;
    pfd.cDepthBits = 24;
    pfd.cStencilBits = 8;
    pfd.iLayerType = PFD_MAIN_PLANE;

    int pf = ChoosePixelFormat(win->src.hdc, &pfd);
    if (!pf || !SetPixelFormat(win->src.hdc, pf, &pfd)) return RGFW_FALSE;

    ctx->ctx = wglCreateContext(win->src.hdc);
    if (!ctx->ctx) return RGFW_FALSE;

    win->src.ctx_native = ctx;
    win->src.gfxType = RGFW_gfxNativeOpenGL;
    return RGFW_TRUE;
}

static RGFW_glContext* RGFW_window_createContext_OpenGL(RGFW_window* win, RGFW_glHints* hints) {
    RGFW_glContext* ctx = (RGFW_glContext*)RGFW_ALLOC(sizeof(RGFW_glContext));
    if (!ctx) return NULL;
    if (!RGFW_window_createContextPtr_OpenGL(win, ctx, hints)) { RGFW_FREE(ctx); return NULL; }
    win->src.gfxType |= RGFW_gfxOwnedByRGFW;
    return ctx;
}

// -------------------------------------------------------------------
//  macOS (CGL, offscreen)
// -------------------------------------------------------------------
#elif defined(RGFW_MACOS)

static i32 RGFW_init(const char* name, u8 flags) { (void)name; (void)flags; return 0; }
static void RGFW_deinit(void) {}

static RGFW_window* RGFW_createWindow(const char* name, i32 x, i32 y, i32 w, i32 h, RGFW_windowFlags flags) {
    (void)x; (void)y; (void)flags; (void)name;
    RGFW_window* win = (RGFW_window*)RGFW_ALLOC(sizeof(RGFW_window));
    if (!win) return NULL;
    memset(win, 0, sizeof(*win));
    win->w = w; win->h = h;
    return win;
}

static void RGFW_window_close(RGFW_window* win) {
    if (!win) return;
    if (win->src.ctx_native && (win->src.gfxType & RGFW_gfxNativeOpenGL)) {
        if (win->src.ctx_native->ctx) CGLDestroyContext((CGLContextObj)win->src.ctx_native->ctx);
        if (win->src.gfxType & RGFW_gfxOwnedByRGFW) RGFW_FREE(win->src.ctx_native);
    }
    RGFW_FREE(win);
}

static void RGFW_window_setName(RGFW_window* win, const char* name) { (void)win; (void)name; }

static void RGFW_window_makeCurrentContext_OpenGL(RGFW_window* win) {
    if (win && win->src.ctx_native && win->src.ctx_native->ctx)
        CGLSetCurrentContext((CGLContextObj)win->src.ctx_native->ctx);
}

static RGFW_proc RGFW_getProcAddress_OpenGL(const char* name) {
    static void* gl_handle;
    if (!gl_handle) gl_handle = dlopen("/System/Library/Frameworks/OpenGL.framework/OpenGL", RTLD_LAZY);
    if (gl_handle) { void* sym = dlsym(gl_handle, name); if (sym) return (RGFW_proc)sym; }
    return NULL;
}

static RGFW_bool RGFW_window_createContextPtr_OpenGL(RGFW_window* win, RGFW_glContext* ctx, RGFW_glHints* hints) {
    (void)win;
    if (!ctx) return RGFW_FALSE;

    CGLPixelFormatAttribute attrs[32];
    int n = 0;
    attrs[n++] = kCGLPFAOpenGLProfile;
    attrs[n++] = (CGLPixelFormatAttribute)kCGLOGLPVersion_Legacy;
    if (hints && hints->renderer == RGFW_glAccelerated) attrs[n++] = kCGLPFAAccelerated;
    attrs[n++] = kCGLPFAColorSize;  attrs[n++] = 32;
    attrs[n++] = kCGLPFADepthSize;  attrs[n++] = hints ? hints->depth : 24;
    attrs[n++] = kCGLPFAStencilSize; attrs[n++] = hints ? hints->stencil : 8;
    if (hints && hints->doubleBuffer) attrs[n++] = kCGLPFADoubleBuffer;
    attrs[n++] = 0;

    CGLPixelFormatObj pix = NULL; GLint npix = 0;
    if (CGLChoosePixelFormat(attrs, &pix, &npix) != kCGLNoError || !pix) return RGFW_FALSE;

    CGLError err = CGLCreateContext(pix, NULL, (CGLContextObj*)&ctx->ctx);
    CGLDestroyPixelFormat(pix);
    if (err != kCGLNoError || !ctx->ctx) return RGFW_FALSE;

    if (CGLSetCurrentContext((CGLContextObj)ctx->ctx) != kCGLNoError) {
        CGLDestroyContext((CGLContextObj)ctx->ctx); ctx->ctx = NULL; return RGFW_FALSE;
    }

    win->src.ctx_native = ctx;
    win->src.gfxType = RGFW_gfxNativeOpenGL;
    return RGFW_TRUE;
}

static RGFW_glContext* RGFW_window_createContext_OpenGL(RGFW_window* win, RGFW_glHints* hints) {
    RGFW_glContext* ctx = (RGFW_glContext*)RGFW_ALLOC(sizeof(RGFW_glContext));
    if (!ctx) return NULL;
    if (!RGFW_window_createContextPtr_OpenGL(win, ctx, hints)) { RGFW_FREE(ctx); return NULL; }
    win->src.gfxType |= RGFW_gfxOwnedByRGFW;
    return ctx;
}

// -------------------------------------------------------------------
//  X11 (GLX)
// -------------------------------------------------------------------
#else

static Display* RGFW_nano_display = NULL;
static Atom     RGFW_nano_WM_DELETE_WINDOW;

static i32 RGFW_init(const char* name, u8 flags) {
    (void)name; (void)flags;
    RGFW_nano_display = XOpenDisplay(NULL);
    if (!RGFW_nano_display) return -1;
    RGFW_nano_WM_DELETE_WINDOW = XInternAtom(RGFW_nano_display, "WM_DELETE_WINDOW", False);
    return 0;
}

static void RGFW_deinit(void) {
    if (RGFW_nano_display) { XCloseDisplay(RGFW_nano_display); RGFW_nano_display = NULL; }
}

static RGFW_window* RGFW_createWindow(const char* name, i32 x, i32 y, i32 w, i32 h, RGFW_windowFlags flags) {
    (void)x; (void)y; (void)flags;
    if (!RGFW_nano_display) return NULL;

    RGFW_window* win = (RGFW_window*)RGFW_ALLOC(sizeof(RGFW_window));
    if (!win) return NULL;
    memset(win, 0, sizeof(*win));
    win->w = w; win->h = h;
    win->src.display = RGFW_nano_display;

    int scr = DefaultScreen(RGFW_nano_display);
    Window root = DefaultRootWindow(RGFW_nano_display);

    XSetWindowAttributes swa;
    swa.event_mask = StructureNotifyMask | ExposureMask;
    swa.colormap = XCreateColormap(RGFW_nano_display, root, DefaultVisual(RGFW_nano_display, scr), AllocNone);
    win->src.cmap = swa.colormap;

    win->src.window = XCreateWindow(RGFW_nano_display, root, 0, 0, w, h, 0,
        DefaultDepth(RGFW_nano_display, scr), InputOutput,
        DefaultVisual(RGFW_nano_display, scr), CWEventMask | CWColormap, &swa);
    if (!win->src.window) { RGFW_FREE(win); return NULL; }

    XSetWMProtocols(RGFW_nano_display, win->src.window, &RGFW_nano_WM_DELETE_WINDOW, 1);
    RGFW_window_setName(win, name);
    XMapWindow(RGFW_nano_display, win->src.window);
    XFlush(RGFW_nano_display);
    return win;
}

static void RGFW_window_close(RGFW_window* win) {
    if (!win) return;
    if (win->src.ctx_native && (win->src.gfxType & RGFW_gfxNativeOpenGL)) {
        glXMakeCurrent(win->src.display, None, NULL);
        if (win->src.ctx_native->ctx) glXDestroyContext(win->src.display, win->src.ctx_native->ctx);
        if (win->src.gfxType & RGFW_gfxOwnedByRGFW) RGFW_FREE(win->src.ctx_native);
    }
    if (win->src.window) { XDestroyWindow(win->src.display, win->src.window); win->src.window = 0; }
    if (win->src.cmap) { XFreeColormap(win->src.display, win->src.cmap); win->src.cmap = 0; }
    RGFW_FREE(win);
}

static void RGFW_window_setName(RGFW_window* win, const char* name) {
    if (win && win->src.display && win->src.window) XStoreName(win->src.display, win->src.window, name);
}

static void RGFW_window_makeCurrentContext_OpenGL(RGFW_window* win) {
    if (win && win->src.display && win->src.ctx_native && win->src.ctx_native->ctx)
        glXMakeCurrent(win->src.display, win->src.window, win->src.ctx_native->ctx);
}

static RGFW_proc RGFW_getProcAddress_OpenGL(const char* name) {
    return (RGFW_proc)glXGetProcAddress((const GLubyte*)name);
}

static int RGFW_nano_glx_attribs(int* attribs, int max, RGFW_glHints* hints) {
    int n = 0;
    (void)max;
    attribs[n++] = GLX_RENDER_TYPE;   attribs[n++] = GLX_RGBA_BIT;
    attribs[n++] = GLX_DRAWABLE_TYPE; attribs[n++] = GLX_WINDOW_BIT;
    attribs[n++] = GLX_DOUBLEBUFFER;  attribs[n++] = hints->doubleBuffer ? True : False;
    attribs[n++] = GLX_RED_SIZE;      attribs[n++] = hints->red;
    attribs[n++] = GLX_GREEN_SIZE;    attribs[n++] = hints->green;
    attribs[n++] = GLX_BLUE_SIZE;     attribs[n++] = hints->blue;
    attribs[n++] = GLX_ALPHA_SIZE;    attribs[n++] = hints->alpha;
    attribs[n++] = GLX_DEPTH_SIZE;    attribs[n++] = hints->depth;
    attribs[n++] = GLX_STENCIL_SIZE;  attribs[n++] = hints->stencil;
    attribs[n++] = None;
    return n;
}

static RGFW_bool RGFW_window_createContextPtr_OpenGL(RGFW_window* win, RGFW_glContext* ctx, RGFW_glHints* hints) {
    if (!win || !win->src.display || !ctx) return RGFW_FALSE;

    int attribs[64];
    RGFW_nano_glx_attribs(attribs, 64, hints);

    int fbcount;
    GLXFBConfig* fbcs = glXChooseFBConfig(win->src.display, DefaultScreen(win->src.display), attribs, &fbcount);
    if (!fbcs || fbcount == 0) return RGFW_FALSE;

    win->src.fbconfig = fbcs[0];
    XFree(fbcs);

    GLXContext share = (hints && hints->share) ? hints->share->ctx : NULL;

    typedef GLXContext (*glXCreateContextAttribsARBProc)(Display*, GLXFBConfig, GLXContext, Bool, const int*);
    glXCreateContextAttribsARBProc glXCreateContextAttribsARB =
        (glXCreateContextAttribsARBProc)glXGetProcAddress((const GLubyte*)"glXCreateContextAttribsARB");

    if (glXCreateContextAttribsARB && hints && (hints->major > 0 || hints->minor > 0)) {
        int ctx_attribs[] = {
            GLX_CONTEXT_MAJOR_VERSION_ARB, hints->major,
            GLX_CONTEXT_MINOR_VERSION_ARB, hints->minor,
            GLX_CONTEXT_PROFILE_MASK_ARB,
            hints->profile == RGFW_glCore ? GLX_CONTEXT_CORE_PROFILE_BIT_ARB :
            hints->profile == RGFW_glCompatibility ? GLX_CONTEXT_COMPATIBILITY_PROFILE_BIT_ARB :
            GLX_CONTEXT_ES_PROFILE_BIT_EXT,
            None
        };
        ctx->ctx = glXCreateContextAttribsARB(win->src.display, win->src.fbconfig, share, True, ctx_attribs);
    }
    if (!ctx->ctx) {
        ctx->ctx = glXCreateNewContext(win->src.display, win->src.fbconfig, GLX_RGBA_TYPE, share, True);
    }
    if (!ctx->ctx) return RGFW_FALSE;

    win->src.ctx_native = ctx;
    win->src.gfxType = RGFW_gfxNativeOpenGL;
    return RGFW_TRUE;
}

static RGFW_glContext* RGFW_window_createContext_OpenGL(RGFW_window* win, RGFW_glHints* hints) {
    RGFW_glContext* ctx = (RGFW_glContext*)RGFW_ALLOC(sizeof(RGFW_glContext));
    if (!ctx) return NULL;
    if (!RGFW_window_createContextPtr_OpenGL(win, ctx, hints)) { RGFW_FREE(ctx); return NULL; }
    win->src.gfxType |= RGFW_gfxOwnedByRGFW;
    return ctx;
}
#endif
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// RGFW is a graphics context backed by the inlined rgfw_nano C code.
type RGFW struct {
	win    *C.RGFW_window
	inited bool
}

// NewRGFWContext creates a hidden window with an OpenGL context and an FBO.
func NewRGFWContext(cfg Config) (*RGFW, error) {
	name := C.CString("cloud-retro")
	defer C.free(unsafe.Pointer(name))

	if ret := C.RGFW_init(name, 0); ret < 0 {
		return nil, fmt.Errorf("rgfw init failed: %d", int(ret))
	}

	if !cfg.GLAutoContext {
		var hints C.RGFW_glHints
		C.RGFW_resetGlobalHints_OpenGL()
		hints.major = C.i32(cfg.GLVersionMajor)
		hints.minor = C.i32(cfg.GLVersionMinor)
		if cfg.GLHasDepth {
			hints.depth = 24
		}
		if cfg.GLHasStencil {
			hints.stencil = 8
		}
		hints.doubleBuffer = 1
		hints.renderer = C.RGFW_glSoftware

		switch cfg.Ctx {
		case CtxOpenGlCore:
			hints.profile = C.RGFW_glCore
		case CtxOpenGlEs2, CtxOpenGlEs3, CtxOpenGlEsVersion:
			hints.profile = C.RGFW_glES
			hints.major = 3
			hints.minor = 0
		case CtxOpenGl:
			hints.profile = C.RGFW_glCompatibility
		default:
			hints.profile = C.RGFW_glCompatibility
		}
		C.RGFW_setGlobalHints_OpenGL(&hints)
	}

	wname := C.CString("cloud-retro")
	defer C.free(unsafe.Pointer(wname))
	win := C.RGFW_createWindow(wname, 0, 0, 1, 1, C.RGFW_windowHide)
	if win == nil {
		C.RGFW_deinit()
		return nil, fmt.Errorf("rgfw window creation failed")
	}

	if C.RGFW_window_createContext_OpenGL(win, C.RGFW_getGlobalHints_OpenGL()) == nil {
		C.RGFW_window_close(win)
		C.RGFW_deinit()
		return nil, fmt.Errorf("rgfw GL context creation failed")
	}

	C.RGFW_window_makeCurrentContext_OpenGL(win)

	initContext(func(name string) unsafe.Pointer {
		cname := C.CString(name)
		defer C.free(unsafe.Pointer(cname))
		return unsafe.Pointer(C.RGFW_getProcAddress_OpenGL(cname))
	})

	if err := initFramebuffer(cfg.W, cfg.H, cfg.GLHasDepth, cfg.GLHasStencil); err != nil {
		C.RGFW_window_close(win)
		C.RGFW_deinit()
		return nil, fmt.Errorf("rgfw fbo: %w", err)
	}

	return &RGFW{win: win, inited: true}, nil
}

func (r *RGFW) Deinit() error {
	if !r.inited {
		return nil
	}
	r.inited = false
	destroyFramebuffer()
	C.RGFW_window_close(r.win)
	r.win = nil
	C.RGFW_deinit()
	return nil
}

func (r *RGFW) BindContext() error {
	if !r.inited {
		return fmt.Errorf("rgfw: context not initialized")
	}
	C.RGFW_window_makeCurrentContext_OpenGL(r.win)
	return nil
}

func RGFWGlProcAddress(proc string) unsafe.Pointer {
	cname := C.CString(proc)
	defer C.free(unsafe.Pointer(cname))
	return unsafe.Pointer(C.RGFW_getProcAddress_OpenGL(cname))
}

func RGFWTryInit() error {
	className := C.CString("cloud-retro-test")
	defer C.free(unsafe.Pointer(className))
	if ret := C.RGFW_init(className, 0); ret < 0 {
		return fmt.Errorf("rgfw init failed: %d", int(ret))
	}
	C.RGFW_deinit()
	return nil
}
