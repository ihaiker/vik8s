# -*- mode: ruby -*-
# vi: set ft=ruby :
#

boxes = {
  "centos7" => "centos/7",
  "centos8" => "roboxes/centos8",
  "ubuntu" => "generic/ubuntu1804"
}
box = ENV["box"] ? ENV["box"] : "centos8"
box_name = boxes[box]

nodes = {
  "master01" => [2, 4096, 40],
  "slave20" => [2, 4096, 40],
  "slave21" => [2, 4096, 40],
}

Vagrant.configure("2") do |config|
  nodes.each do |(name, cfg)|
    vcpus, memory, storage = cfg

    config.vm.define name do |machine|
      machine.vm.hostname = name
      machine.vm.box = box_name

      machine.vm.provider :vmware_esxi do |esxi|
        esxi.esxi_hostname = 'api.esxi.do'
        esxi.esxi_username = 'root'
        esxi.esxi_password = 'env:API_EXSI_DO'
        esxi.guest_numvcpus = vcpus
        esxi.guest_memsize = memory
        esxi.guest_boot_disk_size = storage
        esxi.local_allow_overwrite = 'True'
      end
    end
  end
  if box != "ubuntu"
    config.vm.provision "ansible" do |ansible|
      ansible.playbook = "playbook.yml"
    end
  end
end

# https://github.com/josenk/vagrant-vmware-esxi
