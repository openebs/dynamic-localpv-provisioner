v3.5.0 / 2023-12-12
===================
* fix: quota is not calculated correctly ([#161](https://github.com/openebs/dynamic-localpv-provisioner/pull/161),[@MingZhang-YBPS](https://github.com/MingZhang-YBPS))
* feat(usage): update ua to ga4 analytics ([#166](https://github.com/openebs/dynamic-localpv-provisioner/pull/166),[@niladrih](https://github.com/niladrih))

v3.4.0 / 2022-09-30
===================
* chore: allow resolution of templating values ([#162](https://github.com/openebs/dynamic-localpv-provisioner/pull/162),[@Abhinandan-Purkait](https://github.com/Abhinandan-Purkait))

v3.3.0 / 2022-07-13
===================
* feat(hostpath): enforce quotas for hostpath with an ext4 filesystem ([#137](https://github.com/openebs/dynamic-localpv-provisioner/pull/137),[@hickersonj](https://github.com/hickersonj))

v3.2.0 / 2022-04-19
===================
* fix bug where klog logging flags are not parsed ([#127](https://github.com/openebs/dynamic-localpv-provisioner/pull/127), [@niladrih](https://github.com/niladrih))
* fix bug where XFS-Quota does not work with LVM ([#130](https://github.com/openebs/dynamic-localpv-provisioner/pull/130), [@csschwe](https://github.com/csschwe))


v3.1.0 / 2022-01-06
========================
* add support for multiple Node Affinity Labels for both hostpath and device volumes. ([#102](https://github.com/openebs/dynamic-localpv-provisioner/pull/102),[@Ab-hishek](https://https://github.com/Ab-hishek))
* add support for BlockDevice label selectors with device volumes. ([#106](https://github.com/openebs/dynamic-localpv-provisioner/pull/106),[@Ab-hishek](https://https://github.com/Ab-hishek))


v3.0.0 / 2021-09-22
========================
* add support for enabling XFS project quota in hostpath volumes. ([#78](https://github.com/openebs/dynamic-localpv-provisioner/pull/78),[@almas33](https://github.com/almas33))


v2.8.0 / 2021-04-14
========================
* fix provisioner crashing when old PVs are not cleaned up. ([#39](https://github.com/openebs/dynamic-localpv-provisioner/pull/39),[@niladrih](https://github.com/niladrih))


v2.8.0-RC1 / 2021-04-07
========================
* fix provisioner crashing when old PVs are not cleaned up. ([#39](https://github.com/openebs/dynamic-localpv-provisioner/pull/39),[@niladrih](https://github.com/niladrih))



v2.7.0 / 2021-03-11
========================
* add support to push multiarch images to multiple registries and remove travis from repository ([#32](https://github.com/openebs/dynamic-localpv-provisioner/pull/32),[@akhilerm](https://github.com/akhilerm))


v2.7.0-RC2 / 2021-03-10
========================
* add support to push multiarch images to multiple registries and remove travis from repository ([#32](https://github.com/openebs/dynamic-localpv-provisioner/pull/32),[@akhilerm](https://github.com/akhilerm))


v2.7.0-RC1 / 2021-03-08
========================
No changes since v2.6.0



v2.6.0 / 2021-02-13
========================
No changes since v2.5.0



v2.5.0 / 2021-01-13
========================
* add openebs localpv helm charts ([#14](https://github.com/openebs/dynamic-localpv-provisioner/pull/14),[@prateekpandey14](https://github.com/prateekpandey14))
* support passing image pull secrets when creating helper pod by localpv provisioner ([#22](https://github.com/openebs/dynamic-localpv-provisioner/pull/22),[@allenhaozi](https://github.com/allenhaozi))


v2.5.0-RC1 / 2021-01-08
========================
* add openebs localpv helm charts ([#14](https://github.com/openebs/dynamic-localpv-provisioner/pull/14),[@prateekpandey14](https://github.com/prateekpandey14))
* support passing image pull secrets when creating helper pod by localpv provisioner ([#22](https://github.com/openebs/dynamic-localpv-provisioner/pull/22),[@allenhaozi](https://github.com/allenhaozi))



v2.4.0 / 2020-12-13
========================
* allow custom node affinity label in place of hostnames for localpv hostpath provisioner ([#15](https://github.com/openebs/dynamic-localpv-provisioner/pull/15),[@kmova](https://github.com/kmova))


v2.4.0-RC1 / 2020-12-11
========================
* allow custom node affinity label in place of hostnames for localpv hostpath provisioner ([#15](https://github.com/openebs/dynamic-localpv-provisioner/pull/15),[@kmova](https://github.com/kmova))



v2.3.0 / 2020-11-14
========================
* add support for multiarch builds to localpv provisioner ([#2](https://github.com/openebs/dynamic-localpv-provisioner/pull/2),[@akhilerm](https://github.com/akhilerm))


v2.3.0-RC1 / 2020-11-11
========================
* add support for multiarch builds to localpv provisioner ([#2](https://github.com/openebs/dynamic-localpv-provisioner/pull/2),[@akhilerm](https://github.com/akhilerm))



# Changelog


v2.2.0 / 2020-10-14
========================

The Changelog for v2.2.0 and prior releases were maintaind under https://github.com/openebs/maya

