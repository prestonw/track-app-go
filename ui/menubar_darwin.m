#import <Cocoa/Cocoa.h>
#include "menubar_darwin.h"

extern void trackapp_menu_open(void);
extern void trackapp_menu_today(void);
extern void trackapp_menu_jobs(void);
extern void trackapp_menu_settings(void);
extern void trackapp_menu_toggle_hud(void);
extern void trackapp_menu_quit(void);

@interface TrackAppMenuTarget : NSObject
@end

@implementation TrackAppMenuTarget
- (void)openApp:(id)sender { trackapp_menu_open(); }
- (void)openToday:(id)sender { trackapp_menu_today(); }
- (void)openJobs:(id)sender { trackapp_menu_jobs(); }
- (void)openSettings:(id)sender { trackapp_menu_settings(); }
- (void)toggleHUD:(id)sender { trackapp_menu_toggle_hud(); }
- (void)quitApp:(id)sender { trackapp_menu_quit(); }
@end

static NSStatusItem *gStatusItem = nil;
static NSMenuItem *gStatusLineItem = nil;
static NSMenuItem *gHUDItem = nil;
static TrackAppMenuTarget *gTarget = nil;

static NSMenuItem *trackappAddItem(NSMenu *menu, NSString *title, SEL action, NSString *key) {
    NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:title action:action keyEquivalent:key ?: @""];
    item.target = gTarget;
    [menu addItem:item];
    return item;
}

void trackapp_menubar_install(void) {
    if (gStatusItem != nil) {
        return;
    }
    if ([NSApp respondsToSelector:@selector(setActivationPolicy:)]) {
        [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    }
    gTarget = [[TrackAppMenuTarget alloc] init];
    gStatusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSStatusItemVariableLength];
    gStatusItem.button.toolTip = @"Track App";

    NSMenu *menu = [[NSMenu alloc] initWithTitle:@"Track App"];
    gStatusLineItem = [[NSMenuItem alloc] initWithTitle:@"Track App" action:nil keyEquivalent:@""];
    gStatusLineItem.enabled = NO;
    [menu addItem:gStatusLineItem];
    [menu addItem:[NSMenuItem separatorItem]];

    trackappAddItem(menu, @"Open Track App", @selector(openApp:), @"o");
    trackappAddItem(menu, @"Today", @selector(openToday:), @"t");
    trackappAddItem(menu, @"Job Timers", @selector(openJobs:), @"j");
    trackappAddItem(menu, @"Settings…", @selector(openSettings:), @",");
    gHUDItem = trackappAddItem(menu, @"Show Floating Timer", @selector(toggleHUD:), @"f");
    [menu addItem:[NSMenuItem separatorItem]];
    trackappAddItem(menu, @"Quit", @selector(quitApp:), @"q");

    gStatusItem.menu = menu;
}

void trackapp_menubar_set_icon(const unsigned char *data, int len) {
    if (!gStatusItem || !data || len <= 0) {
        return;
    }
    NSData *blob = [NSData dataWithBytes:data length:(NSUInteger)len];
    NSImage *img = [[NSImage alloc] initWithData:blob];
    if (!img) {
        return;
    }
    img.template = YES;
    img.size = NSMakeSize(18, 18);
    gStatusItem.button.image = img;
    gStatusItem.button.imagePosition = NSImageOnly;
}

void trackapp_menubar_set_status(const char *text) {
    if (!gStatusLineItem) {
        return;
    }
    if (!text) {
        gStatusLineItem.title = @"Track App";
        return;
    }
    gStatusLineItem.title = [NSString stringWithUTF8String:text];
}

void trackapp_menubar_set_hud_label(const char *text) {
    if (!gHUDItem || !text) {
        return;
    }
    gHUDItem.title = [NSString stringWithUTF8String:text];
}