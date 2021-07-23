# -*- mode: ruby -*-
# vi: set ft=ruby :

centos7_image = "https://mirrors.ustc.edu.cn/centos-cloud/centos/7/vagrant/x86_64/images/CentOS-7-x86_64-Vagrant-2004_01.VirtualBox.box"
centos8_image = "https://mirrors.ustc.edu.cn/centos-cloud/centos/8/vagrant/x86_64/images/CentOS-8-Vagrant-8.4.2105-20210603.0.x86_64.vagrant-virtualbox.box"
guest_iso_path = "https://mirrors.tuna.tsinghua.edu.cn/virtualbox/%{version}/VBoxGuestAdditions_%{version}.iso"

box = ENV["box"]

Vagrant.configure("2") do |config|

  if box == "ubuntu"
    #config.vm.box_url = "https://mirrors.ustc.edu.cn/ubuntu-cloud-images/vagrant/trusty/current/trusty-server-cloudimg-amd64-vagrant-disk1.box"
    #config.vm.box = "ubuntu/18.04"
    config.vm.box = "hashicorp/bionic64"
  elsif box == "centos7"
    config.vm.box_url = centos7_image
    config.vm.box = "centos/7.2004.01"
  else
    config.vm.box_url = centos8_image
    config.vm.box = "centos/8.4.2105"
  end

  config.vbguest.iso_path = guest_iso_path
  config.vbguest.auto_update = false

  config.vm.define "master0" do |node|
    node.vm.hostname = "master0.vik8s.com"
    node.vm.network "private_network", ip: "10.24.0.10"
    node.vm.provider "virtualbox" do |v|
      v.cpus = "2"
      v.memory = "2048"
    end
    if box != "ubuntu"
      node.vm.provision "ansible" do |ansible|
        ansible.playbook = "playbook.yml"
      end
    end
  end

  # config.vm.define "slave20" do |node|
  #   node.vm.hostname = "slave20.vik8s.com"
  #   node.vm.network "private_network", ip: "10.24.0.20"
  #   node.vm.provider "virtualbox" do |v|
  #     v.cpus = "1"
  #     v.memory = "1024"
  #   end
  # end
  # config.vm.define "slave21" do |node|
  #   node.vm.hostname = "slave21.vik8s.com"
  #   node.vm.network "private_network", ip: "10.24.0.21"
  #   node.vm.provider "virtualbox" do |v|
  #     v.cpus = "1"
  #     v.memory = "1024"
  #   end
  # end

end
