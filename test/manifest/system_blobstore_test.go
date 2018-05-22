package manifest_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("System Blobstore", func() {
	Describe("s3 compatible", func() {
		var (
			inputProperties map[string]interface{}
			buckets         = []string{"buildpacks", "droplets", "packages", "resources"}
		)

		BeforeEach(func() {
			inputProperties = map[string]interface{}{
				".properties.system_blobstore":                            "external",
				".properties.system_blobstore.external.buildpacks_bucket": "some-buildpacks-bucket",
				".properties.system_blobstore.external.droplets_bucket":   "some-droplets-bucket",
				".properties.system_blobstore.external.packages_bucket":   "some-packages-bucket",
				".properties.system_blobstore.external.resources_bucket":  "some-resources-bucket",
			}
		})

		Context("when the user enables versioning", func() {
			It("enables the s3-versioned-blobstore-backup-restorer, and disables the s3-unversioned-blobstore-backup-restorer", func() {
				inputProperties[".properties.system_blobstore.external.versioning"] = true

				manifest, err := product.RenderService.RenderManifest(inputProperties)
				Expect(err).NotTo(HaveOccurred())

				job, err := manifest.FindInstanceGroupJob("backup-prepare", "s3-versioned-blobstore-backup-restorer")
				Expect(err).NotTo(HaveOccurred())

				jobEnabled, err := job.Property("enabled")
				Expect(err).NotTo(HaveOccurred())
				Expect(jobEnabled).To(BeTrue())

				job, err = manifest.FindInstanceGroupJob("backup-prepare", "s3-unversioned-blobstore-backup-restorer")
				Expect(err).NotTo(HaveOccurred())

				jobEnabled, err = job.Property("enabled")
				Expect(err).NotTo(HaveOccurred())
				Expect(jobEnabled).To(BeFalse())
			})

			Context("and IAM instance profiles are disabled", func() {

				BeforeEach(func() {
					inputProperties[".properties.system_blobstore.external.iam_instance_profile_authentication"] = false
					inputProperties[".properties.system_blobstore.external.access_key"] = "some-access-key-id"
					inputProperties[".properties.system_blobstore.external.secret_key"] = map[string]string{
						"secret": "some-secret-access-key",
					}
				})

				It("specifies that backups use the provided access key", func() {
					manifest, err := product.RenderService.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob("backup-prepare", "s3-versioned-blobstore-backup-restorer")
					Expect(err).NotTo(HaveOccurred())

					for _, bucket := range buckets {
						bucketProperties, err := job.Property(fmt.Sprintf("buckets/%s", bucket))
						Expect(err).NotTo(HaveOccurred())
						Expect(bucketProperties).NotTo(HaveKey("use_iam_profile"))
						Expect(bucketProperties).To(HaveKeyWithValue("aws_access_key_id", "some-access-key-id"))
						Expect(bucketProperties).To(HaveKeyWithValue("aws_secret_access_key", "some-secret-access-key"))
					}
				})

			})

			Context("and IAM instance profiles are enabled", func() {

				BeforeEach(func() {
					inputProperties[".properties.system_blobstore.external.iam_instance_profile_authentication"] = true
				})

				It("specifies that backups should use it", func() {
					manifest, err := product.RenderService.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob("backup-prepare", "s3-versioned-blobstore-backup-restorer")
					Expect(err).NotTo(HaveOccurred())

					for _, bucket := range buckets {
						iamInstanceProfileAuthentication, err := job.Property(fmt.Sprintf("buckets/%s/use_iam_profile", bucket))
						Expect(err).NotTo(HaveOccurred())
						Expect(iamInstanceProfileAuthentication).To(BeTrue())

						bucketProperties, err := job.Property(fmt.Sprintf("buckets/%s", bucket))
						Expect(err).NotTo(HaveOccurred())
						Expect(bucketProperties).NotTo(HaveKey("aws_access_key_id"))
						Expect(bucketProperties).NotTo(HaveKey("aws_access_secret_key"))
					}
				})

			})

		})

		Context("when the user disables versioning", func() {

			BeforeEach(func() {
				inputProperties[".properties.system_blobstore.external.backup_region"] = "some-backup-region"
				inputProperties[".properties.system_blobstore.external.buildpacks_backup_bucket"] = "some-buildpacks-bucket"
				inputProperties[".properties.system_blobstore.external.droplets_backup_bucket"] = "some-droplets-bucket"
				inputProperties[".properties.system_blobstore.external.packages_backup_bucket"] = "some-packages-bucket"
				inputProperties[".properties.system_blobstore.external.resources_backup_bucket"] = "some-resources-bucket"
			})

			It("disables the s3-versioned-blobstore-backup-restorer and enables the s3-unversioned-blobstore-backup-restorer", func() {
				manifest, err := product.RenderService.RenderManifest(inputProperties)
				Expect(err).NotTo(HaveOccurred())

				job, err := manifest.FindInstanceGroupJob("backup-prepare", "s3-versioned-blobstore-backup-restorer")
				Expect(err).NotTo(HaveOccurred())

				jobEnabled, err := job.Property("enabled")
				Expect(err).NotTo(HaveOccurred())
				Expect(jobEnabled).To(BeFalse())

				job, err = manifest.FindInstanceGroupJob("backup-prepare", "s3-unversioned-blobstore-backup-restorer")
				Expect(err).NotTo(HaveOccurred())

				jobEnabled, err = job.Property("enabled")
				Expect(err).NotTo(HaveOccurred())
				Expect(jobEnabled).To(BeTrue())

				buildpacksBackupRegion, err := job.Property("buckets/buildpacks/backup/region")
				Expect(err).NotTo(HaveOccurred())
				Expect(buildpacksBackupRegion).To(Equal("some-backup-region"))

				buildpacksBackupName, err := job.Property("buckets/buildpacks/backup/name")
				Expect(err).NotTo(HaveOccurred())
				Expect(buildpacksBackupName).To(Equal("some-buildpacks-bucket"))

				dropletsBackupRegion, err := job.Property("buckets/droplets/backup/region")
				Expect(err).NotTo(HaveOccurred())
				Expect(dropletsBackupRegion).To(Equal("some-backup-region"))

				dropletsBackupName, err := job.Property("buckets/droplets/backup/name")
				Expect(err).NotTo(HaveOccurred())
				Expect(dropletsBackupName).To(Equal("some-droplets-bucket"))

				packagesBackupRegion, err := job.Property("buckets/packages/backup/region")
				Expect(err).NotTo(HaveOccurred())
				Expect(packagesBackupRegion).To(Equal("some-backup-region"))

				packagesBackupName, err := job.Property("buckets/packages/backup/name")
				Expect(err).NotTo(HaveOccurred())
				Expect(packagesBackupName).To(Equal("some-packages-bucket"))

				resourcesBackupRegion, err := job.Property("buckets/resources/backup/region")
				Expect(err).NotTo(HaveOccurred())
				Expect(resourcesBackupRegion).To(Equal("some-backup-region"))

				resourcesBackupName, err := job.Property("buckets/resources/backup/name")
				Expect(err).NotTo(HaveOccurred())
				Expect(resourcesBackupName).To(Equal("some-resources-bucket"))
			})

			Context("and IAM instance profiles are disabled", func() {

				BeforeEach(func() {
					inputProperties[".properties.system_blobstore.external.iam_instance_profile_authentication"] = false
					inputProperties[".properties.system_blobstore.external.access_key"] = "some-access-key-id"
					inputProperties[".properties.system_blobstore.external.secret_key"] = map[string]string{
						"secret": "some-secret-access-key",
					}
				})

				It("specifies that backups use the provided access key", func() {
					manifest, err := product.RenderService.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob("backup-prepare", "s3-unversioned-blobstore-backup-restorer")
					Expect(err).NotTo(HaveOccurred())

					for _, bucket := range buckets {
						bucketProperties, err := job.Property(fmt.Sprintf("buckets/%s", bucket))
						Expect(err).NotTo(HaveOccurred())
						Expect(bucketProperties).NotTo(HaveKey("use_iam_profile"))
						Expect(bucketProperties).To(HaveKeyWithValue("aws_access_key_id", "some-access-key-id"))
						Expect(bucketProperties).To(HaveKeyWithValue("aws_secret_access_key", "some-secret-access-key"))
					}
				})

			})

			Context("and IAM instance profiles are enabled", func() {

				BeforeEach(func() {
					inputProperties[".properties.system_blobstore.external.iam_instance_profile_authentication"] = true
				})

				It("specifies that backups should use it", func() {
					manifest, err := product.RenderService.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob("backup-prepare", "s3-unversioned-blobstore-backup-restorer")
					Expect(err).NotTo(HaveOccurred())

					for _, bucket := range buckets {
						iamInstanceProfileAuthentication, err := job.Property(fmt.Sprintf("buckets/%s/use_iam_profile", bucket))
						Expect(err).NotTo(HaveOccurred())
						Expect(iamInstanceProfileAuthentication).To(BeTrue())

						bucketProperties, err := job.Property(fmt.Sprintf("buckets/%s", bucket))
						Expect(err).NotTo(HaveOccurred())
						Expect(bucketProperties).NotTo(HaveKey("aws_access_key_id"))
						Expect(bucketProperties).NotTo(HaveKey("aws_access_secret_key"))
					}
				})

			})

		})
	})
})
