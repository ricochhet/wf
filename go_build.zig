const std = @import("std");

pub fn addExecutable(b: *std.Build, options: GoBuildStep.Options) *GoBuildStep {
    return GoBuildStep.create(b, options);
}

/// Convenience method to create a library.
pub fn addLibrary(b: *std.Build, options: GoBuildStep.Options) *GoBuildStep {
    var library_options = options;
    library_options.make_library = true;
    // A library build requires CGO.
    library_options.cgo_enabled = true;
    return GoBuildStep.create(b, library_options);
}

pub fn build(b: *std.Build) void {
    _ = b;
}

/// Runs `go build` with relevant flags
pub const GoBuildStep = struct {
    step: std.Build.Step,
    generated_bin: ?*std.Build.GeneratedFile,
    opts: Options,

    pub const Options = struct {
        // The name of generated target.
        name: []const u8,
        target: std.Build.ResolvedTarget,
        optimize: std.builtin.OptimizeMode,
        package_path: std.Build.LazyPath,
        cgo_enabled: bool = true,
        make_library: bool = false,
        // TODO(rjk): Add Mac target sysroot support.
        custom: []const u8,
        trim_path: bool = true,
        link_external: bool = true,
        release: bool = true,
    };

    /// Create a GoBuildStep
    pub fn create(b: *std.Build, options: Options) *GoBuildStep {
        const self = b.allocator.create(GoBuildStep) catch unreachable;
        self.* = .{
            .opts = options,
            .generated_bin = null,
            .step = std.Build.Step.init(.{
                .id = .custom,
                .name = "go build",
                .owner = b,
                .makeFn = GoBuildStep.make,
            }),
        };
        return self;
    }

    pub fn make(step: *std.Build.Step, mkoptions: std.Build.Step.MakeOptions) !void {
        const self: *GoBuildStep = @fieldParentPtr("step", step);
        const b = step.owner;
        var go_args = std.ArrayList([]const u8).init(b.allocator);
        defer go_args.deinit();

        try go_args.append("go");
        try go_args.append("build");

        var output_file = try b.cache_root.join(b.allocator, &.{ "go", self.opts.name });
        const target_os = self.opts.target.result.os.tag;
        if (target_os == .windows) {
            output_file = try std.fmt.allocPrint(b.allocator, "{s}.exe", .{output_file});
        }

        try go_args.appendSlice(&.{ "-o", output_file });

        switch (self.opts.optimize) {
            .ReleaseSafe => try go_args.appendSlice(&.{ "-tags", "ReleaseSafe" }),
            .ReleaseFast => try go_args.appendSlice(&.{ "-tags", "ReleaseFast" }),
            .ReleaseSmall => try go_args.appendSlice(&.{ "-tags", "ReleaseFast" }),
            .Debug => try go_args.appendSlice(&.{ "-tags", "Debug" }),
        }

        var env = try std.process.getEnvMap(b.allocator);

        // GOOS and GOARCH fields are not the same as triple fields on some
        // platforms. Adjust them appropriately.
        const target = self.opts.target;
        const goarch = try isa_to_goarch(@tagName(target.result.cpu.arch));
        try env.put("GOARCH", goarch);
        const goos = try ostag_to_goos(@tagName(target.result.os.tag));
        try env.put("GOOS", goos);

        // Cross compilling Go to MacOS requires providing a sysroot. The default
        // sysroot exists when the target is macos, a MacOS SDK is installed and
        // the Zig target is native. Otherwise, a sysroot *might* be present in
        // some fashion. Add support for specifying the system root.

        // Trim path
        if (self.opts.trim_path and self.opts.release) {
            try go_args.appendSlice(&.{"-trimpath"});
        }

        // CGO
        if (self.opts.cgo_enabled) {
            try env.put("CGO_ENABLED", "1");
            // Set zig as the CGO compiler
            const ts = try self.mktargetstring(b);
            const cc = b.fmt(
                "zig cc -target {s}",
                .{ts},
            );
            try env.put("CC", cc);
            const cxx = b.fmt(
                "zig c++ -target {s}",
                .{ts},
            );
            try env.put("CXX", cxx);
        } else {
            try env.put("CGO_ENABLED", "0");
        }

        const customflags = if (self.opts.custom.len != 0) self.opts.custom else "";
        const linkmode = if (self.opts.link_external) "external" else "internal";

        // Tell the linker we are statically linking
        var ldflags_list = std.ArrayList(u8).init(b.allocator);
        defer ldflags_list.deinit();

        try ldflags_list.writer().print(
            "{s} -linkmode={s} -extldflags=-static",
            .{ customflags, linkmode },
        );

        // Strip debug symbols in release
        if (self.opts.release) {
            try ldflags_list.appendSlice(" -w -s");
        }

        try go_args.appendSlice(&.{"-ldflags"});

        const flags_copy = try b.allocator.dupe(u8, ldflags_list.items);
        try go_args.append(flags_copy);

        if (self.opts.make_library) {
            // Where does the .h file go? (next to it)
            try go_args.appendSlice(&.{"-buildmode=c-archive"});
        }

        // Package path always needs to be added last
        try go_args.append(self.opts.package_path.getPath(b));

        const cmd = std.mem.join(b.allocator, " ", go_args.items) catch @panic("OOM");
        const node = mkoptions.progress_node.start(cmd, 1);
        defer node.end();

        // Run the command
        std.debug.print("Running command: {s}\n", .{cmd});
        try self.evalChildProcess(go_args.items, &env);

        if (self.generated_bin == null) {
            const generated_bin = b.allocator.create(std.Build.GeneratedFile) catch unreachable;
            generated_bin.* = .{ .step = step };
            self.generated_bin = generated_bin;
        }
        self.generated_bin.?.path = output_file;
    }

    /// Return the LazyPath of the directory containing the Go-generated include file.
    pub fn getIncludePath(self: *GoBuildStep) std.Build.LazyPath {
        return self.getEmittedBin().dirname();
    }

    /// Return the LazyPath of the generated binary
    pub fn getEmittedBin(self: *GoBuildStep) std.Build.LazyPath {
        if (self.generated_bin) |generated_bin|
            return .{ .generated = .{ .file = generated_bin } };

        const b = self.step.owner;
        const generated_bin = b.allocator.create(std.Build.GeneratedFile) catch unreachable;
        generated_bin.* = .{ .step = &self.step };
        self.generated_bin = generated_bin;
        return .{ .generated = .{ .file = generated_bin } };
    }

    /// Add a run step which depends on the GoBuildStep
    pub fn addRunStep(self: *GoBuildStep) *std.Build.Step.Run {
        const b = self.step.owner;
        const run_step = std.Build.Step.Run.create(b, b.fmt("run {s}", .{self.opts.name}));
        run_step.step.dependOn(&self.step);
        const bin_file = self.getEmittedBin();
        const arg: std.Build.Step.Run.PrefixedLazyPath = .{ .prefix = "", .lazy_path = bin_file };
        run_step.argv.append(b.allocator, .{ .lazy_path = arg }) catch unreachable;
        return run_step;
    }

    /// Add an install step which depends on the GoBuildStep
    pub fn addInstallStep(self: *GoBuildStep) void {
        const b = self.step.owner;
        const bin_file = self.getEmittedBin();

        var output_file = self.opts.name;
        defer b.allocator.free(output_file);

        const target_os = self.opts.target.result.os.tag;

        if (target_os == .windows) {
            output_file = std.fmt.allocPrint(b.allocator, "{s}.exe", .{output_file}) catch unreachable;
        }

        const install_step = b.addInstallBinFile(bin_file, output_file);
        install_step.step.dependOn(&self.step);
        b.getInstallStep().dependOn(&install_step.step);
    }

    fn evalChildProcess(self: *GoBuildStep, argv: []const []const u8, env: *const std.process.EnvMap) !void {
        const s = &self.step;
        const arena = s.owner.allocator;

        try std.Build.Step.handleChildProcUnsupported(s, null, argv);
        try std.Build.Step.handleVerbose(s.owner, null, argv);

        const result = std.process.Child.run(.{
            .allocator = arena,
            .argv = argv,
            .env_map = env,
        }) catch |err| return s.fail("unable to spawn {s}: {s}", .{ argv[0], @errorName(err) });

        if (result.stderr.len > 0) {
            try s.result_error_msgs.append(arena, result.stderr);
        }

        try std.Build.Step.handleChildProcessTerm(s, result.term, null, argv);
    }

    /// Create a target that can build the Go code. Uses "native" on MacOS
    /// so that it has a sysroot.
    // TODO(rjk): Add some kind of completed sysroot support.
    fn mktargetstring(self: *GoBuildStep, b: *std.Build) ![]const u8 {
        const target = self.opts.target;
        if (target.result.os.tag == .ios or target.result.os.tag == .macos) {
            return "native";
        } else {
            return b.fmt(
                "{s}-{s}-{s}",
                .{
                    @tagName(target.result.cpu.arch),
                    @tagName(target.result.os.tag),
                    @tagName(target.result.abi),
                },
            );
        }
        return error.NOTIMPL_CUSTOM_MACOS_SYSROOT;
    }
};

const GoEnvError = error{
    UNSUPPORTED_GOARCH,
    UNSUPPORTED_GOOS,
};

/// Convert a Zig arch triple member to the corresponding GOARCH value.
fn isa_to_goarch(isa: [:0]const u8) ![:0]const u8 {
    if (std.mem.eql(u8, isa, "x86_64")) {
        return "amd64";
    } else if (std.mem.eql(u8, isa, "arm")) {
        return "arm";
    } else if (std.mem.eql(u8, isa, "aarch64")) {
        return "arm64";
    }
    // TODO(rjk): Consider adding some less popular processors that might
    // work.

    return error.UNSUPPORTED_GOARCH;
}

/// Convert a Zig os triple member to the corresponding GOOS value.
fn ostag_to_goos(os: [:0]const u8) ![:0]const u8 {
    if (std.mem.startsWith(u8, os, "macos")) {
        return "darwin";
    } else if (std.mem.startsWith(u8, os, "linux")) {
        return "linux";
    } else if (std.mem.startsWith(u8, os, "windows")) {
        return "windows";
    }
    // TODO(rjk): Add less common OS. In particular, support Plan9.
    return error.UNSUPPORTED_GOOS;
}

const MacEnvError = error{
    NOTIMPL_CUSTOM_MACOS_SYSROOT,
};