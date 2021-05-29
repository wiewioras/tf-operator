package certificates

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	core "k8s.io/api/core/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tungstenfabric/tf-operator/pkg/k8s"
)

// Certificate object
type Certificate struct {
	client              client.Client
	scheme              *runtime.Scheme
	owner               v1.Object
	sc                  *k8s.Secret
	signer              CertificateSigner
	certificateSubjects []CertificateSubject
}

// NewCertificate creates new cert
func NewCertificate(signer CertificateSigner, cl client.Client, scheme *runtime.Scheme, owner v1.Object, subjects []CertificateSubject, ownerType string) (*Certificate, error) {
	secretName := owner.GetName() + "-secret-certificates"
	kubernetes := k8s.New(cl, scheme)
	return &Certificate{
		client:              cl,
		scheme:              scheme,
		owner:               owner,
		sc:                  kubernetes.Secret(secretName, ownerType, owner),
		signer:              signer,
		certificateSubjects: subjects,
	}, nil
}

// EnsureExistsAndIsSigned ensures cert is signed
func (r *Certificate) EnsureExistsAndIsSigned() error {
	return r.sc.EnsureExists(r)
}

type CertificateSigner interface {
	SignCertificate(secret *core.Secret, certTemplate x509.Certificate, privateKey *rsa.PrivateKey) ([]byte, error)
}

// FillSecret fill secret with data
func (r *Certificate) FillSecret(secret *core.Secret) error {
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}

	for _, subject := range r.certificateSubjects {
		if subject.ip == "" {
			continue
		}
		if err := r.createCertificateForPod(subject, secret); err != nil {
			return err
		}
	}

	for _, subject := range r.certificateSubjects {
		if subject.ip == "" {
			return fmt.Errorf("%s subject IP still no available", subject.name)
		}
	}

	return nil
}

func (r *Certificate) createCertificateForPod(subject CertificateSubject, secret *core.Secret) error {
	if certInSecret(secret, subject.ip) {
		return nil
	}
	certificateTemplate, privateKey, err := subject.generateCertificateTemplate(r.client)
	if err != nil {
		return fmt.Errorf("failed to generate certificate template for %s, %s: %w", subject.hostname, subject.name, err)
	}

	certBytes, err := r.signer.SignCertificate(secret, certificateTemplate, privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign certificate for %s, %s: %w", subject.hostname, subject.name, err)
	}

	certPrivKeyPem, _ := encodeInPemFormat(x509.MarshalPKCS1PrivateKey(privateKey), privateKeyPemType)
	secret.Data[serverPrivateKeyFileName(subject.ip)] = certPrivKeyPem
	secret.Data[serverCertificateFileName(subject.ip)] = certBytes
	secret.Data["status-"+subject.ip] = []byte("Approved")
	return nil
}

func certInSecret(secret *core.Secret, podIP string) bool {
	_, pemOk := secret.Data[serverPrivateKeyFileName(podIP)]
	_, certOk := secret.Data[serverCertificateFileName(podIP)]
	return pemOk && certOk
}

func serverPrivateKeyFileName(ip string) string {
	return fmt.Sprintf("server-key-%s.pem", ip)
}

func serverCertificateFileName(ip string) string {
	return fmt.Sprintf("server-%s.crt", ip)
}
