# -*- mode: ruby -*-
# vi: set ft=ruby :

box = ENV["box"] ? ENV["box"] : "centos8"
provider = ENV["provider"] ? ENV["provider"] : "virtualbox"

centos7_image = "https://mirrors.ustc.edu.cn/centos-cloud/centos/7/vagrant/x86_64/images/CentOS-7-x86_64-Vagrant-2004_01.VirtualBox.box"
centos8_image = "https://mirrors.ustc.edu.cn/centos-cloud/centos/8/vagrant/x86_64/images/CentOS-8-Vagrant-8.4.2105-20210603.0.x86_64.vagrant-virtualbox.box"
ubuntu_image = "https://mirrors.ustc.edu.cn/ubuntu-cloud-images/vagrant/trusty/current/trusty-server-cloudimg-amd64-vagrant-disk1.box"
guest_iso_path = "https://mirrors.tuna.tsinghua.edu.cn/virtualbox/%{version}/VBoxGuestAdditions_%{version}.iso"

boxes = {
    "centos7" => ["centos/7", "centos/7.2004.01", centos7_image],
    "centos8" => ["roboxes/centos8", "centos/8.4.2105", centos8_image],
    "ubuntu" => ["generic/ubuntu1804", "ubuntu/18.04", ubuntu_image]
}

nodes = {
    "master01" => [2, 4096, 40],
    "slave20" => [2, 4096, 40],
    "slave21" => [2, 4096, 40],
}

Vagrant.configure("2") do |config|

  exsi_boxname, box_name, box_url = boxes[box]

  if box != "ubuntu"
    config.vm.provision "ansible" do |ansible|
      ansible.playbook = "playbook.yml"
    end
  end

  nodes.each do |(name, cfg)|
    numvcpus, memory, storage = cfg

    config.vm.define name do |machine|
      machine.vm.hostname = name

      if provider == "virtualbox"
        machine.vm.box = box_name
        machine.vm.box_url = box_url
        machine.vbguest.iso_path = guest_iso_path
        machine.vbguest.auto_update = false
        config.vm.network "private_network", type: "dhcp"
        machine.vm.provider "virtualbox" do |v|
          v.cpus = numvcpus
          v.memory = memory
        end
      else
        # https://github.com/josenk/vagrant-vmware-esxi
        machine.vm.box = exsi_boxname
        machine.vm.provider :vmware_esxi do |esxi|
          esxi.esxi_hostname = 'api.esxi.do'
          esxi.esxi_username = 'root'
          esxi.esxi_password = 'env:API_EXSI_DO'
          esxi.guest_numvcpus = numvcpus
          esxi.guest_memsize = memory
          esxi.guest_boot_disk_size = storage
          esxi.local_allow_overwrite = 'True'
        end
      end
    end
  end
end
