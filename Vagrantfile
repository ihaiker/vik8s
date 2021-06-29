# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|

  config.vm.box_url = "https://mirrors.ustc.edu.cn/centos-cloud/centos/8/vagrant/x86_64/images/CentOS-8-Vagrant-8.4.2105-20210603.0.x86_64.vagrant-virtualbox.box"
  config.vm.box = "centos/8.4"

  config.vbguest.iso_path = "https://mirrors.tuna.tsinghua.edu.cn/virtualbox/%{version}/VBoxGuestAdditions_%{version}.iso"
  config.vbguest.auto_update = false

  config.vm.synced_folder ".", "/vagrant"

  config.vm.provider "virtualbox" do |v|
    v.cpus = "1"
    v.memory = "2048"
  end

  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "playbook.yml"
  end

  config.vm.define "master0" do |master0|
    master0.vm.hostname = "master0"
    master0.vm.network "private_network", ip: "10.24.0.10"
  end

  # config.vm.define "slave20" do |master0|
  #   master0.vm.hostname = "slave20"
  #   master0.vm.network "private_network", ip: "10.24.0.20"
  # end
  # config.vm.define "slave21" do |master0|
  #   master0.vm.hostname = "slave21"
  #   master0.vm.network "private_network", ip: "10.24.0.21"
  # end

end
