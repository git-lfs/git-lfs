require_relative "../lib/distro"

def test_map
  {
    "centos/8" => {
      name: "RPM RHEL 8/Rocky Linux 8",
      component: "el/8",
      image: "centos_8",
      package_type: "rpm",
      package_tag: "-1.el8",
      equivalent: [
        "el/8",
      ],
    },
    "rocky/9" => {
      name: "RPM RHEL 9/Rocky Linux 9",
      component: "el/9",
      image: "rocky_9",
      package_type: "rpm",
      package_tag: "-1.el9",
      equivalent: [
        "el/9",
        "fedora/41",
        "fedora/42",
        "opensuse/15.6",
        "sles/15.6",
      ],
    },
    "debian/12" => {
      name: "Debian 12",
      component: "debian/bookworm",
      image: "debian_12",
      package_type: "deb",
      package_tag: "",
      equivalent: [
        "debian/bookworm",
        "debian/trixie",
      ]
    },
  }
end

context DistroMapProgram do
  it "should print image names" do
    stdout = StringIO.new
    stderr = StringIO.new
    expect(DistroMapProgram.new(stdout, stderr, test_map).run(["--image-names"])).to eq 0
    expect(stderr.string).to be_empty
    expect(stdout.string).to eq "centos_8 rocky_9 debian_12\n"
  end

  it "should print distro markdown" do
    stdout = StringIO.new
    stderr = StringIO.new
    expect(DistroMapProgram.new(stdout, stderr, test_map).run(["--distro-markdown"])).to eq 0
    expect(stderr.string).to be_empty
    expected = <<~EOM
    [RPM RHEL 8/Rocky Linux 8](https://packagecloud.io/github/git-lfs/packages/el/8/git-lfs-VERSION-1.el8.x86_64.rpm/download)
    [RPM RHEL 9/Rocky Linux 9](https://packagecloud.io/github/git-lfs/packages/el/9/git-lfs-VERSION-1.el9.x86_64.rpm/download)
    [Debian 12](https://packagecloud.io/github/git-lfs/packages/debian/bookworm/git-lfs_VERSION_amd64.deb/download)
    EOM
    expect(stdout.string).to eq expected
  end

  it "should whine when no options were given" do
    stdout = StringIO.new
    stderr = StringIO.new
    expect(DistroMapProgram.new(stdout, stderr, test_map).run([])).to eq 2
    expect(stdout.string).to be_empty
    expect(stderr.string).to eq "A mode option is required\n"
  end
end

context DistroMap do
  it "should produce the correct distro names" do
    map = {
      "centos/8" => [
        "el/8",
      ],
      "rocky/9" => [
        "el/9",
        "fedora/41",
        "fedora/42",
        "opensuse/15.6",
        "sles/15.6",
      ],
      "debian/12" => [
        "debian/bookworm",
        "debian/trixie",
      ],
    }
    expect(DistroMap.new(test_map).distro_name_map).to eq map
  end

  it "should produce the correct image names" do
    expect(DistroMap.new(test_map).image_names).to eq %w[centos_8 rocky_9 debian_12]
  end
end
