#import <Cocoa/Cocoa.h>
#import <QuartzCore/QuartzCore.h>

static NSRect trackappCornerFrame(int corner, NSSize panelSize, NSScreen *screen, CGFloat margin) {
    NSRect vf = screen.visibleFrame;
    CGFloat x = vf.origin.x + margin;
    CGFloat y = vf.origin.y + vf.size.height - panelSize.height - margin;
    switch (corner) {
        case 1:
            x = vf.origin.x + vf.size.width - panelSize.width - margin;
            break;
        case 2:
            y = vf.origin.y + margin;
            break;
        case 3:
            x = vf.origin.x + vf.size.width - panelSize.width - margin;
            y = vf.origin.y + margin;
            break;
    }
    CGFloat grid = 24.0;
    x = round(x / grid) * grid;
    y = round(y / grid) * grid;
    return NSMakeRect(x, y, panelSize.width, panelSize.height);
}

static NSWindow *trackappFindHUDWindow(void) {
    NSWindow *best = nil;
    CGFloat bestArea = 1e12;
    for (NSWindow *win in [NSApp windows]) {
        if (!win.isVisible) {
            continue;
        }
        NSSize sz = win.frame.size;
        CGFloat area = sz.width * sz.height;
        if (area > 120000.0) {
            continue;
        }
        BOOL borderless = (win.styleMask & NSWindowStyleMaskBorderless) != 0;
        NSString *title = win.title ?: @"";
        BOOL hudTitle = [title containsString:@"HUD"] || [title containsString:@"TrackApp"] || title.length == 0;
        if (borderless || (hudTitle && area < 100000.0)) {
            if (area < bestArea) {
                bestArea = area;
                best = win;
            }
        }
    }
    return best;
}

void trackapp_place_hud_window(int corner, int width, int height, double margin, int animate) {
    if (width <= 0 || height <= 0) {
        return;
    }
    NSWindow *win = trackappFindHUDWindow();
    if (!win) {
        return;
    }
    NSScreen *screen = win.screen;
    if (!screen) {
        screen = NSScreen.mainScreen;
    }
    NSSize size = win.frame.size;
    if (size.width <= 0 || size.height <= 0) {
        size = NSMakeSize((CGFloat)width, (CGFloat)height);
    }
    NSRect frame = trackappCornerFrame(corner, size, screen, (CGFloat)margin);
    [win setLevel:NSFloatingWindowLevel];
    win.collectionBehavior = NSWindowCollectionBehaviorCanJoinAllSpaces | NSWindowCollectionBehaviorFullScreenAuxiliary;
    win.hidesOnDeactivate = NO;
    if (animate) {
        [NSAnimationContext runAnimationGroup:^(NSAnimationContext *context) {
            context.duration = 0.15;
            context.timingFunction = [CAMediaTimingFunction functionWithName:kCAMediaTimingFunctionEaseOut];
            [[win animator] setFrame:frame display:YES];
        } completionHandler:nil];
    } else {
        [win setFrame:frame display:YES];
    }
}