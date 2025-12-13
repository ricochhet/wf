// zig build -Dname=app -Dpackage-path=cmd/app -Dflags="$(CUSTOM)"
const std = @import("std");
const go = @import("go_build.zig");

pub fn build(b: *std.Build) void {
    buildWithOptions(b, go.GoBuildStep.Options{
        .name = b.option([]const u8, "name", "Name of executable") orelse "app",
        .target = b.standardTargetOptions(.{}),
        .optimize = b.standardOptimizeOption(.{}),
        .package_path = b.path(b.option([]const u8, "package-path", "Go main entrypoint") orelse "cmd/app"),
        .cgo_enabled = b.option(bool, "cgo-enabled", "Set the CGO_ENABLED flag") orelse true,
        .custom = b.option([]const u8, "flags", "Specifiy custom ldflags") orelse "",
        .trim_path = b.option(bool, "trim-path", "Set the -trimpath flag (release only)") orelse true,
        .link_external = b.option(bool, "link-external", "Set the -linkmode=external flag") orelse true,
        .release = b.option(bool, "release", "Build in release mode") orelse true,
    });
}

pub fn buildWithOptions(b: *std.Build, opts: go.GoBuildStep.Options) void {
    const go_build = go.addExecutable(b, opts);

    const run_cmd = go_build.addRunStep();
    if (b.args) |args| {
        run_cmd.addArgs(args);
    }
    const run_step = b.step("run", "Run the app");
    run_step.dependOn(&run_cmd.step);

    go_build.addInstallStep();
}