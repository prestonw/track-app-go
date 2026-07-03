#import <Cocoa/Cocoa.h>
#import <QuartzCore/QuartzCore.h>
#import <ApplicationServices/ApplicationServices.h>

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

static BOOL trackappOwnerMatches(NSString *owner) {
    if (owner == nil || owner.length == 0) {
        return NO;
    }
    NSString *lower = owner.lowercaseString;
    return [lower containsString:@"track"] || [lower isEqualToString:@"trackapp"];
}

static NSInteger trackappFindWindowNumber(void) {
    CFArrayRef infoList = CGWindowListCopyWindowInfo(
        kCGWindowListOptionOnScreenOnly | kCGWindowListExcludeDesktopElements,
        kCGNullWindowID);
    if (!infoList) {
        return -1;
    }

    NSInteger bestNumber = -1;
    CGFloat bestArea = 1e12;
    CFIndex count = CFArrayGetCount(infoList);
    for (CFIndex i = 0; i < count; i++) {
        CFDictionaryRef info = (CFDictionaryRef)CFArrayGetValueAtIndex(infoList, i);
        CFStringRef ownerRef = CFDictionaryGetValue(info, kCGWindowOwnerName);
        if (!trackappOwnerMatches((__bridge NSString *)ownerRef)) {
            continue;
        }

        CFNumberRef layerRef = CFDictionaryGetValue(info, kCGWindowLayer);
        int layer = 0;
        if (layerRef) {
            CFNumberGetValue(layerRef, kCFNumberIntType, &layer);
        }
        if (layer != 0) {
            continue;
        }

        CFDictionaryRef boundsRef = CFDictionaryGetValue(info, kCGWindowBounds);
        if (!boundsRef) {
            continue;
        }
        CGRect bounds;
        if (!CGRectMakeWithDictionaryRepresentation(boundsRef, &bounds)) {
            continue;
        }

        CGFloat area = bounds.size.width * bounds.size.height;
        if (area < 3000.0 || area > 180000.0) {
            continue;
        }
        if (area >= bestArea) {
            continue;
        }

        CFNumberRef numRef = CFDictionaryGetValue(info, kCGWindowNumber);
        if (!numRef) {
            continue;
        }
        NSInteger num = 0;
        CFNumberGetValue(numRef, kCFNumberNSIntegerType, &num);
        bestNumber = num;
        bestArea = area;
    }
    CFRelease(infoList);
    return bestNumber;
}

static NSWindow *trackappFindHUDWindow(void) {
    NSInteger target = trackappFindWindowNumber();
    if (target >= 0) {
        for (NSWindow *win in [NSApp windows]) {
            if (win.windowNumber == target) {
                return win;
            }
        }
    }

    NSWindow *best = nil;
    CGFloat bestArea = 1e12;
    for (NSWindow *win in [NSApp windows]) {
        NSSize sz = win.frame.size;
        CGFloat area = sz.width * sz.height;
        if (area < 3000.0 || area > 180000.0) {
            continue;
        }
        BOOL borderless = (win.styleMask & NSWindowStyleMaskBorderless) != 0;
        NSString *title = win.title ?: @"";
        BOOL hudTitle = [title containsString:@"HUD"] || [title containsString:@"TrackApp"];
        if (borderless || hudTitle) {
            if (area < bestArea) {
                bestArea = area;
                best = win;
            }
        }
    }
    return best;
}

static void trackappApplyFrame(NSWindow *win, NSRect frame, int animate) {
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
    trackappApplyFrame(win, frame, animate);
}