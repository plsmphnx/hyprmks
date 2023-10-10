# Hyprland Modifier Key Sequence Generator

This is a small tool to generate Hyprland binds which allow modifier keys to be
entered as key sequences, alongside the (typical) key chords. It does so by
generating a submap for each used modifier key combination, alongside the
release binds necessary to enter and exit them via the modifier keys.

## Usage

Simply include binds with modifiers as usual in your `hyprland.conf` and call:

```
hyprmks <hyprland.conf>
```

The tool will generate submap names from the combination of modifier keys;
alternatively, the `#alias=MODS,name` directive can be used to rename the
submap for a particular modifier combination.
