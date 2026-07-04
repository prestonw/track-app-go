#import <Cocoa/Cocoa.h>
#import <QuartzCore/QuartzCore.h>

static BOOL trackappOwnerMatches(NSString *owner) {
    if (owner == nil || owner.length == 0) {
        return NO;
    }
    NSString *lower = owner.lowercaseString;
    return [lower containsString:@"track"] || [lower isEqualToString:@"trackapp"];
}

static NSWindow *trackappFindMainWindow(void) {
    NSWindow *best = nil;
    CGFloat bestArea = 0;
    for (NSWindow *win in [NSApp windows]) {
        if (!win.isVisible) {
            continue;
        }
        NSSize sz = win.frame.size;
        CGFloat area = sz.width * sz.height;
        if (area < 200000.0) {
            continue;
        }
        if (area > bestArea) {
            bestArea = area;
            best = win;
        }
    }
    return best;
}

void trackapp_main_window_borderless(void) {
    NSWindow *win = trackappFindMainWindow();
    if (!win) {
        return;
    }
    win.styleMask = NSWindowStyleMaskBorderless | NSWindowStyleMaskResizable | NSWindowStyleMaskFullSizeContentView;
    win.titlebarAppearsTransparent = YES;
    win.titleVisibility = NSWindowTitleHidden;
    win.backgroundColor = [NSColor clearColor];
    win.opaque = NO;
    win.hasShadow = YES;
}

void trackapp_main_window_hide_animated(void) {
    NSWindow *win = trackappFindMainWindow();
    if (!win) {
        return;
    }
    [NSAnimationContext runAnimationGroup:^(NSAnimationContext *context) {
        context.duration = 0.18;
        context.timingFunction = [CAMediaTimingFunction functionWithName:kCAMediaTimingFunctionEaseIn];
        [[win animator] setAlphaValue:0.0];
    } completionHandler:^{
        [win orderOut:nil];
        [win setAlphaValue:1.0];
    }];
}