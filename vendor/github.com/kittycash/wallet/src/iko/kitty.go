package iko

import (
	"sort"
	"strconv"

	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
)

var (
	ErrInvalidKittyName      = errors.New("kitty has invalid name")
	ErrInvalidKittyDesc      = errors.New("kitty has invalid description")
	ErrInvalidKittyBreed     = errors.New("kitty has invalid breed")
	ErrInvalidKittyBirthDate = errors.New("kitty has invalid birth date")
	ErrInvalidKittyDNA       = errors.New("kitty has invalid DNA")
)

/*
	<<< KITTY DETAILS >>>
	>>> Used by multiple services, provides off-chain details for kitties and IKO.
*/

type Kitty struct {
	ID        KittyID `json:"kitty_id"`    // Identifier for kitty.
	Name      string  `json:"name"`        // Name of kitty.
	Desc      string  `json:"description"` // Description of kitty.
	Breed     string  `json:"breed"`       // Kitty breed.
	Legendary bool    `json:"legendary"`   // Whether kitty is legendary.

	PriceBTC int64 `json:"price_btc"` // Price of kitty in BTC.
	PriceSKY int64 `json:"price_sky"` // Price of kitty in SKY.

	BoxOpen bool `json:"box_open"` // Whether box is open.

	Created   int64  `json:"created,omitempty"`    // Timestamp that the kitty box began existing.
	BirthDate int64  `json:"birth_date,omitempty"` // Timestamp of box opening (after box opening).
	KittyDNA  string `json:"kitty_dna,omitempty"`  // Hex representation of kitty DNA (after box opening).

	BoxImgURL   string `json:"box_image_url,omitempty"`   // Box image TransformURL.
	KittyImgURL string `json:"kitty_image_url,omitempty"` // Kitty image TransformURL (after box opening).
}

func (k *Kitty) Json() []byte {
	raw, _ := json.Marshal(k)
	return raw
}

func (k *Kitty) Hash() cipher.SHA256 {
	return cipher.SumSHA256(k.Json())
}

func (k *Kitty) Sign(sk cipher.SecKey) cipher.Sig {
	return cipher.SignHash(k.Hash(), sk)
}

func (k *Kitty) Verify(pk cipher.PubKey, sig cipher.Sig) error {
	return cipher.VerifySignature(pk, sig, k.Hash())
}

// CheckData ensures that the Kitty fields makes sense. It does the following:
// - Ensure Name/Desc/Breed fields are filled.
//		- Ensure Desc, Breed and Name fields are all valid and based on algorithm.
// - When BoxOpen == true :
//		- Ensure BirthDate/KittyDNA/KittyImgURL are all set.
// - When BoxOpen == false :
//		- Ensure BirthDate/KittyDNA/KittyImgURL are not set.
//		- Ensure BoxImgURL is set.
func (k *Kitty) CheckData() error {
	if err := checkName(k.Name, make(map[string]struct{})); err != nil {
		return err
	}
	if err := checkDesc(k.Desc); err != nil {
		return err
	}
	if err := checkBreed(k.Breed, make(map[string]struct{})); err != nil {
		return err
	}
	if k.BoxOpen {
		if k.BirthDate <= 0 {
			return ErrInvalidKittyBirthDate
		}
		if err := checkDNA(k.KittyDNA); err != nil {
			return err
		}
		if k.KittyImgURL != "" {
			return errors.New("kitty image TransformURL should be set as box is open")
		}
	} else {
		switch {
		case k.BirthDate != 0:
			return errors.New("birth date should be unset as box is not open")
		case k.KittyDNA != "":
			return errors.New("kitty DNA should be unset as box is not open")
		case k.KittyImgURL != "":
			return errors.New("kitty image TransformURL should be unset as box is not open")
		case k.BoxImgURL == "":
			return errors.New("kitty box TransformURL should be set as box is not open")
		}
	}
	return nil
}

func checkName(name string, _ map[string]struct{}) error {
	// TODO: Determine nameList algorithm and mechanism.
	if name == "" {
		return errors.WithMessage(ErrInvalidKittyName,
			"empty name not allowed")
	}
	//if _, ok := validList[name]; !ok {
	//	return errors.WithMessage(ErrInvalidKittyName,
	//		fmt.Sprintf("'%s' is not allowed", name))
	//}
	return nil
}

func checkDesc(desc string) error {
	// TODO: Determine algorithm for checking kitty description.
	if desc == "" {
		return errors.WithMessage(ErrInvalidKittyDesc,
			"an empty description is not allowed")
	}
	return nil
}

func checkBreed(breed string, _ map[string]struct{}) error {
	// TODO: Determine breedList algorithm and mechanism.
	if breed == "" {
		return errors.WithMessage(ErrInvalidKittyBreed,
			"an empty breed is not allowed")
	}
	//if _, ok := validList[breed]; !ok {
	//	return errors.WithMessage(ErrInvalidKittyBreed,
	//		fmt.Sprintf("'%s' is not allowed", breed))
	//}
	return nil
}

func checkDNA(dna string) error {
	// TODO: Determine algorithm for checking kitty DNA.
	if dna == "" {
		return errors.WithMessage(ErrInvalidKittyDNA,
			fmt.Sprintf("'%s' is not allowed", dna))
	}
	return nil
}

/*
	<<< KITTY ID >>>
	>>> For IKO, kitties are indexed with IDs, not DNA.
*/

type KittyID uint64

func KittyIDFromString(idStr string) (KittyID, error) {
	id, e := strconv.ParseUint(idStr, 10, 64)
	return KittyID(id), e
}

type KittyIDs []KittyID

func (ids KittyIDs) Sort() {
	sort.Slice(ids, func(i, j int) bool {
		return (ids)[i] < (ids)[j]
	})
}

func (ids *KittyIDs) Add(id KittyID) {
	*ids = append(*ids, id)
	ids.Sort()
}

func (ids *KittyIDs) Remove(id KittyID) {
	for i, v := range *ids {
		if v == id {
			*ids = append((*ids)[:i], (*ids)[i+1:]...)
			return
		}
	}
}

/*
	<<< KITTY ENTRY >>>
	>>> The way a kitty is entered in the kitty-api.
*/

type KittyEntry struct {
	Kitty
	Sig         string `json:"sig"`               // Signature should be verified with
	Reservation string `json:"reservation"`       // Whether kitty is reserved or not.
	Address     string `json:"address,omitempty"` // The address in which the kitty resides in.
}

func KittyEntryFromJson(raw []byte) (*KittyEntry, error) {
	out := new(KittyEntry)
	err := json.Unmarshal(raw, out)
	return out, err
}

func (e *KittyEntry) Json() []byte {
	raw, _ := json.Marshal(e)
	return raw
}

func (e *KittyEntry) Sign(sk cipher.SecKey) {
	e.Sig = e.Kitty.Sign(sk).Hex()
}

func (e *KittyEntry) Verify(pk cipher.PubKey) error {
	sig, err := cipher.SigFromHex(e.Sig)
	if err != nil {
		return err
	}
	return e.Kitty.Verify(pk, sig)
}

/*
	<<< KITTY STATE >>>
	>>> The state of a kitty as represented when the IKO Chain is compiled.
*/

type KittyState struct {
	Address      cipher.Address
	Transactions TxHashes
}

func (s KittyState) Serialize() []byte {
	return encoder.Serialize(s)
}

/*
	<<< ADDRESS STATE >>>
	>>> The state of an address as represented when the IKO Chain is compiled.
*/

type AddressState struct {
	Kitties      KittyIDs
	Transactions TxHashes
}

func NewAddressState() *AddressState {
	return &AddressState{
		Kitties:      make(KittyIDs, 0),
		Transactions: make(TxHashes, 0),
	}
}

func (a AddressState) Serialize() []byte {
	return encoder.Serialize(a)
}
