package graphics

/*
#cgo windows LDFLAGS: -lopengl32
#cgo linux   LDFLAGS: -lGL
#cgo darwin  LDFLAGS: -framework OpenGL

#include <stdint.h>

#ifndef APIENTRY
#define APIENTRY
#endif
#ifndef APIENTRYP
#define APIENTRYP APIENTRY*
#endif

typedef unsigned int   GLenum;
typedef int            GLint;
typedef unsigned int   GLuint;
typedef int            GLsizei;
typedef unsigned char  GLubyte;

typedef const GLubyte *(APIENTRYP GPGETSTRING)(GLenum name);
typedef GLenum          (APIENTRYP GPCHECKFRAMEBUFFERSTATUS)(GLenum target);
typedef GLenum          (APIENTRYP GPGETERROR)();
typedef void            (APIENTRYP GPBINDFRAMEBUFFER)(GLenum target, GLuint framebuffer);
typedef void            (APIENTRYP GPBINDRENDERBUFFER)(GLenum target, GLuint renderbuffer);
typedef void            (APIENTRYP GPBINDTEXTURE)(GLenum target, GLuint texture);
typedef void            (APIENTRYP GPDELETEFRAMEBUFFERS)(GLsizei n, const GLuint *framebuffers);
typedef void            (APIENTRYP GPDELETERENDERBUFFERS)(GLsizei n, const GLuint *renderbuffers);
typedef void            (APIENTRYP GPDELETETEXTURES)(GLsizei n, const GLuint* textures);
typedef void            (APIENTRYP GPFRAMEBUFFERRENDERBUFFER)(GLenum target, GLenum attachment, GLenum renderbuffertarget, GLuint renderbuffer);
typedef void            (APIENTRYP GPFRAMEBUFFERTEXTURE2D)(GLenum target, GLenum attachment, GLenum textarget, GLuint texture, GLint level);
typedef void            (APIENTRYP GPGENFRAMEBUFFERS)(GLsizei n, GLuint *fb);
typedef void            (APIENTRYP GPGENRENDERBUFFERS)(GLsizei n, GLuint *rb);
typedef void            (APIENTRYP GPGENTEXTURES)(GLsizei n, GLuint *tex);
typedef void            (APIENTRYP GPREADPIXELS)(GLint x, GLint y, GLsizei w, GLsizei h, GLenum f, GLenum t, void *p);
typedef void            (APIENTRYP GPRENDERBUFFERSTORAGE)(GLenum t, GLenum i, GLsizei w, GLsizei h);
typedef void            (APIENTRYP GPTEXIMAGE2D)(GLenum t, GLint l, GLint i, GLsizei w, GLsizei h, GLint b, GLenum f, GLenum x, const void *px);
typedef void            (APIENTRYP GPTEXPARAMETERI)(GLenum t, GLenum n, GLint p);
typedef void            (APIENTRYP GPPIXELSTOREI)(GLenum n, GLint p);

// Generic C wrappers — one per signature pattern, called from Go via function pointers.
static const GLubyte *gl_get_string(GPGETSTRING f, GLenum a)              { return f(a); }
static GLenum          gl_get_error(GPGETERROR f)                        { return f(); }
static GLenum          gl_check_fb_status(GPCHECKFRAMEBUFFERSTATUS f, GLenum a) { return f(a); }
static void gl_1u32(GPGENTEXTURES f, GLsizei a, GLuint *b)       { f(a, b); }
static void gl_bind_tex(GPBINDTEXTURE f, GLenum a, GLuint b)         { f(a, b); }
static void gl_bind_fb(GPBINDFRAMEBUFFER f, GLenum a, GLuint b)   { f(a, b); }
static void gl_bind_rb(GPBINDRENDERBUFFER f, GLenum a, GLuint b)  { f(a, b); }
static void gl_tex_param(GPTEXPARAMETERI f, GLenum a, GLenum b, GLint c) { f(a, b, c); }
static void gl_tex_image(GPTEXIMAGE2D f, GLenum a, GLint b, GLint c, GLsizei d, GLsizei e, GLint g, GLenum h, GLenum i, const void *j) { f(a, b, c, d, e, g, h, i, j); }
static void gl_fb_tex2d(GPFRAMEBUFFERTEXTURE2D f, GLenum a, GLenum b, GLenum c, GLuint d, GLint e) { f(a, b, c, d, e); }
static void gl_fb_rb(GPFRAMEBUFFERRENDERBUFFER f, GLenum a, GLenum b, GLenum c, GLuint d) { f(a, b, c, d); }
static void gl_rb_storage(GPRENDERBUFFERSTORAGE f, GLenum a, GLenum b, GLsizei c, GLsizei d) { f(a, b, c, d); }
static void gl_delete_fb(GPDELETEFRAMEBUFFERS f, GLsizei a, const GLuint *b) { f(a, b); }
static void gl_delete_rb(GPDELETERENDERBUFFERS f, GLsizei a, const GLuint *b) { f(a, b); }
static void gl_delete_tex(GPDELETETEXTURES f, GLsizei a, const GLuint *b) { f(a, b); }
static void gl_read_pixels(GPREADPIXELS f, GLint a, GLint b, GLsizei c, GLsizei d, GLenum e, GLenum g, void *h) { f(a, b, c, d, e, g, h); }
static void gl_pixel_storei(GPPIXELSTOREI f, GLenum a, GLint b) { f(a, b); }
*/
import "C"
import (
	"errors"
	"fmt"
	"math"
	"unsafe"
)

const (
	gl_VENDOR               = 0x1F00
	gl_VERSION              = 0x1F02
	gl_RENDERER             = 0x1F01
	gl_SHADING_LANG_VERSION = 0x8B8C
	gl_TEXTURE_2D           = 0x0DE1
	gl_RENDERBUFFER         = 0x8D41
	gl_FRAMEBUFFER          = 0x8D40
	gl_TEXTURE_MIN_FILTER   = 0x2801
	gl_TEXTURE_MAG_FILTER   = 0x2800
	gl_NEAREST              = 0x2600
	gl_RGBA                 = 0x1908
	gl_RGB                  = 0x1907
	gl_BGRA                 = 0x80E1
	gl_RGBA8                = 0x8058
	gl_COLOR_ATTACHMENT0    = 0x8CE0
	gl_DEPTH_ATTACHMENT     = 0x8D00
	gl_DEPTH_STENCIL_ATTACH = 0x821A
	gl_DEPTH_COMPONENT24    = 0x81A6
	gl_DEPTH24_STENCIL8     = 0x88F0
	gl_FRAMEBUFFER_COMPLETE = 0x8CD5
	gl_PACK_ALIGNMENT       = 0x0D05
)

// ---- GL function pointers ----

var (
	gpGetString,
	gpGetError,
	gpBindTexture,
	gpBindFramebuffer,
	gpBindRenderbuffer,
	gpTexParameteri,
	gpTexImage2D,
	gpGenTextures,
	gpGenFramebuffers,
	gpGenRenderbuffers,
	gpFramebufferTexture2D,
	gpFramebufferRenderbuffer,
	gpRenderbufferStorage,
	gpCheckFramebufferStatus,
	gpDeleteFramebuffers,
	gpDeleteRenderbuffers,
	gpDeleteTextures,
	gpReadPixels,
	gpPixelStorei unsafe.Pointer
)

type glFunc struct {
	ptr  *unsafe.Pointer
	name string
}

var glFuncs = [...]glFunc{
	{&gpGetString, "glGetString"},
	{&gpGetError, "glGetError"},
	{&gpBindTexture, "glBindTexture"},
	{&gpBindFramebuffer, "glBindFramebuffer"},
	{&gpBindRenderbuffer, "glBindRenderbuffer"},
	{&gpTexParameteri, "glTexParameteri"},
	{&gpTexImage2D, "glTexImage2D"},
	{&gpGenTextures, "glGenTextures"},
	{&gpGenFramebuffers, "glGenFramebuffers"},
	{&gpGenRenderbuffers, "glGenRenderbuffers"},
	{&gpFramebufferTexture2D, "glFramebufferTexture2D"},
	{&gpFramebufferRenderbuffer, "glFramebufferRenderbuffer"},
	{&gpRenderbufferStorage, "glRenderbufferStorage"},
	{&gpCheckFramebufferStatus, "glCheckFramebufferStatus"},
	{&gpDeleteFramebuffers, "glDeleteFramebuffers"},
	{&gpDeleteRenderbuffers, "glDeleteRenderbuffers"},
	{&gpDeleteTextures, "glDeleteTextures"},
	{&gpReadPixels, "glReadPixels"},
	{&gpPixelStorei, "glPixelStorei"},
}

type Context int

const (
	CtxNone Context = iota
	CtxOpenGl
	CtxOpenGlEs2
	CtxOpenGlCore
	CtxOpenGlEs3
	CtxOpenGlEsVersion
	CtxVulkan
	CtxUnknown = math.MaxInt32 - 1
	CtxDummy   = math.MaxInt32
)

type PixelFormat int

const (
	UnsignedShort5551 PixelFormat = iota
	UnsignedShort565
	UnsignedInt8888Rev
)

type Config struct {
	Ctx            Context
	W, H           int
	GLAutoContext  bool
	GLVersionMajor uint
	GLVersionMinor uint
	GLHasDepth     bool
	GLHasStencil   bool
}

func initProcAddr(getProcAddr func(name string) unsafe.Pointer) error {
	for _, f := range glFuncs {
		p := getProcAddr(f.name)
		if p == nil {
			return errors.New(f.name)
		}
		*f.ptr = p
	}
	return nil
}

func goStr(s *uint8) string { return C.GoString((*C.char)(unsafe.Pointer(s))) }

func getString(name uint32) *uint8 {
	return (*uint8)(C.gl_get_string((C.GPGETSTRING)(gpGetString), C.GLenum(name)))
}
func genTextures(n int32, textures *uint32) {
	C.gl_1u32((C.GPGENTEXTURES)(gpGenTextures), C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(textures)))
}
func bindTexture(target, texture uint32) {
	C.gl_bind_tex((C.GPBINDTEXTURE)(gpBindTexture), C.GLenum(target), C.GLuint(texture))
}
func bindFramebuffer(target, framebuffer uint32) {
	C.gl_bind_fb((C.GPBINDFRAMEBUFFER)(gpBindFramebuffer), C.GLenum(target), C.GLuint(framebuffer))
}
func bindRenderbuffer(target, renderbuffer uint32) {
	C.gl_bind_rb((C.GPBINDRENDERBUFFER)(gpBindRenderbuffer), C.GLenum(target), C.GLuint(renderbuffer))
}
func texParameteri(target, pname uint32, param int32) {
	C.gl_tex_param((C.GPTEXPARAMETERI)(gpTexParameteri), C.GLenum(target), C.GLenum(pname), C.GLint(param))
}
func texImage2D(target uint32, level, internalformat, width, height, border int32, format, xtype uint32, pixels unsafe.Pointer) {
	C.gl_tex_image((C.GPTEXIMAGE2D)(gpTexImage2D), C.GLenum(target), C.GLint(level), C.GLint(internalformat), C.GLsizei(width), C.GLsizei(height), C.GLint(border), C.GLenum(format), C.GLenum(xtype), pixels)
}
func genFramebuffers(n int32, framebuffers *uint32) {
	C.gl_1u32((C.GPGENTEXTURES)(gpGenFramebuffers), C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(framebuffers)))
}
func framebufferTexture2D(target, attachment, texTarget, texture uint32, level int32) {
	C.gl_fb_tex2d((C.GPFRAMEBUFFERTEXTURE2D)(gpFramebufferTexture2D), C.GLenum(target), C.GLenum(attachment), C.GLenum(texTarget), C.GLuint(texture), C.GLint(level))
}
func genRenderbuffers(n int32, renderbuffers *uint32) {
	C.gl_1u32((C.GPGENTEXTURES)(gpGenRenderbuffers), C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(renderbuffers)))
}
func renderbufferStorage(target, internalformat uint32, width, height int32) {
	C.gl_rb_storage((C.GPRENDERBUFFERSTORAGE)(gpRenderbufferStorage), C.GLenum(target), C.GLenum(internalformat), C.GLsizei(width), C.GLsizei(height))
}
func framebufferRenderbuffer(target, attachment, rbtarget, rb uint32) {
	C.gl_fb_rb((C.GPFRAMEBUFFERRENDERBUFFER)(gpFramebufferRenderbuffer), C.GLenum(target), C.GLenum(attachment), C.GLenum(rbtarget), C.GLuint(rb))
}
func checkFramebufferStatus(target uint32) uint32 {
	return uint32(C.gl_check_fb_status((C.GPCHECKFRAMEBUFFERSTATUS)(gpCheckFramebufferStatus), C.GLenum(target)))
}
func deleteRenderbuffers(n int32, renderbuffers *uint32) {
	C.gl_delete_rb((C.GPDELETERENDERBUFFERS)(gpDeleteRenderbuffers), C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(renderbuffers)))
}
func deleteFramebuffers(n int32, framebuffers *uint32) {
	C.gl_delete_fb((C.GPDELETEFRAMEBUFFERS)(gpDeleteFramebuffers), C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(framebuffers)))
}
func deleteTextures(n int32, textures *uint32) {
	C.gl_delete_tex((C.GPDELETETEXTURES)(gpDeleteTextures), C.GLsizei(n), (*C.GLuint)(unsafe.Pointer(textures)))
}
func readPixels(x, y, width, height int32, format, xtype uint32, pixels unsafe.Pointer) {
	C.gl_read_pixels((C.GPREADPIXELS)(gpReadPixels), C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height), C.GLenum(format), C.GLenum(xtype), pixels)
}
func pixelStorei(pname uint32, param int32) {
	C.gl_pixel_storei((C.GPPIXELSTOREI)(gpPixelStorei), C.GLenum(pname), C.GLint(param))
}

var (
	fbo, tex, rbo      uint32
	hasDepth           bool
	pixType, pixFormat uint32
	buf                []byte
	bufPtr             unsafe.Pointer
)

func initContext(getProcAddr func(name string) unsafe.Pointer) {
	if err := initProcAddr(getProcAddr); err != nil {
		panic(err)
	}
	pixelStorei(gl_PACK_ALIGNMENT, 1)
}

func initFramebuffer(width, height int, depth, stencil bool) error {
	w, h := int32(width), int32(height)
	hasDepth = depth

	genTextures(1, &tex)
	bindTexture(gl_TEXTURE_2D, tex)
	texParameteri(gl_TEXTURE_2D, gl_TEXTURE_MIN_FILTER, gl_NEAREST)
	texParameteri(gl_TEXTURE_2D, gl_TEXTURE_MAG_FILTER, gl_NEAREST)
	texImage2D(gl_TEXTURE_2D, 0, gl_RGBA8, w, h, 0, pixType, pixFormat, nil)
	bindTexture(gl_TEXTURE_2D, 0)

	genFramebuffers(1, &fbo)
	bindFramebuffer(gl_FRAMEBUFFER, fbo)
	framebufferTexture2D(gl_FRAMEBUFFER, gl_COLOR_ATTACHMENT0, gl_TEXTURE_2D, tex, 0)

	if depth {
		genRenderbuffers(1, &rbo)
		bindRenderbuffer(gl_RENDERBUFFER, rbo)
		format, attachment := uint32(gl_DEPTH_COMPONENT24), uint32(gl_DEPTH_ATTACHMENT)
		if stencil {
			format, attachment = gl_DEPTH24_STENCIL8, gl_DEPTH_STENCIL_ATTACH
		}
		renderbufferStorage(gl_RENDERBUFFER, format, w, h)
		framebufferRenderbuffer(gl_FRAMEBUFFER, attachment, gl_RENDERBUFFER, rbo)
		bindRenderbuffer(gl_RENDERBUFFER, 0)
	}

	if status := checkFramebufferStatus(gl_FRAMEBUFFER); status != gl_FRAMEBUFFER_COMPLETE {
		return fmt.Errorf("framebuffer incomplete: 0x%X", status)
	}
	return nil
}

func destroyFramebuffer() {
	if hasDepth {
		deleteRenderbuffers(1, &rbo)
	}
	deleteFramebuffers(1, &fbo)
	deleteTextures(1, &tex)
}

// ReadFramebuffer reads the current FBO color attachment into the internal buffer.
// Call SetBuffer first to size the buffer (w * h * 4).
func ReadFramebuffer(size, w, h uint) []byte {
	bindFramebuffer(gl_FRAMEBUFFER, fbo)
	readPixels(0, 0, int32(w), int32(h), pixType, pixFormat, bufPtr)
	return buf[:size]
}

func SetBuffer(size int) {
	buf = make([]byte, size)
	bufPtr = unsafe.Pointer(&buf[0])
}

func SetPixelFormat(format PixelFormat) error {
	switch format {
	case UnsignedShort5551:
		pixFormat, pixType = 0x8034, gl_BGRA // GL_UNSIGNED_SHORT_5_5_5_1
	case UnsignedShort565:
		pixFormat, pixType = 0x8363, gl_RGB // GL_UNSIGNED_SHORT_5_6_5
	case UnsignedInt8888Rev:
		pixFormat, pixType = 0x8367, gl_BGRA // GL_UNSIGNED_INT_8_8_8_8_REV
	default:
		return errors.New("unknown pixel format")
	}
	return nil
}

func GLInfo() (version, vendor, renderer, glsl string) {
	return goStr(getString(gl_VERSION)),
		goStr(getString(gl_VENDOR)),
		goStr(getString(gl_RENDERER)),
		goStr(getString(gl_SHADING_LANG_VERSION))
}

func GlFbo() uint32 { return fbo }
