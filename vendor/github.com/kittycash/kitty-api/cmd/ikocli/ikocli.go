package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strconv"

	kcrpc "github.com/kittycash/kitty-api/src/rpc"
	"github.com/kittycash/wallet/src/iko"
	inrpc "github.com/kittycash/wallet/src/rpc"
	"github.com/sirupsen/logrus"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/encoder"
	ikocli "gopkg.in/urfave/cli.v1"
)

const (
	Version = "0.1"
)

var (
	app = ikocli.NewApp()
	log = logrus.New()
)

var (
	ErrInvalidIndexRange = errors.New("invalid index range")
	ErrFileNotExist      = errors.New("input file does not exist")
	ErrInvalidTagName    = errors.New("invalid tag name")
)

type cliKitty struct {
	iko.Kitty
	ToAddress string `json:"to_address"`
}

type outBin struct {
	Entries      []iko.KittyEntry
	Transactions []iko.Transaction
}

func init() {
	app.Name = "ikocli"
	app.Usage = "KittyCash CLI is a help tool for iko-chain and kitty-api"
	app.Description = "KittyCash IKO CLI is a tool to submit kitties to iko-chain and kitty-api"
	app.Version = Version
	commands := ikocli.Commands{
		initCommand(),
		editCommand(),
		finalizeCommand(),
	}
	app.Commands = commands
	app.EnableBashCompletion = true
	app.OnUsageError = func(context *ikocli.Context, err error, isSubcommand bool) error {
		fmt.Fprintf(context.App.Writer, "error: %v\n\n", err)
		ikocli.ShowAppHelp(context)
		return nil
	}
	app.CommandNotFound = func(context *ikocli.Context, command string) {
		tmp := fmt.Sprintf("{{.HelpName}}: '%s' is not a {{.HelpName}} "+
			"command. See '{{.HelpName}} --help'. \n", command)
		ikocli.HelpPrinter(app.Writer, tmp, app)
	}
}

func initCommand() ikocli.Command {
	name := "init"
	return ikocli.Command{
		Name:  name,
		Usage: "Creates a file containing a list of kitties of the given index range",
		Flags: []ikocli.Flag{
			ikocli.StringFlag{
				Name:  "index-range",
				Usage: "Range of kitty indexes to generate",
			},
			ikocli.StringFlag{
				Name:  "file",
				Usage: "Data output to `FILE`",
			},
		},
		OnUsageError: OnCommandUsageError(name),
		Action: func(c *ikocli.Context) error {
			indexRange := c.String("index-range")
			start, end, err := ParseIndexRange(indexRange)
			if err != nil {
				return err
			}

			file := c.String("file")
			if file == "" {
				return errors.New("missing output filename")
			}

			fileHandle, err := os.Create(file)
			if err != nil {
				log.Errorf("unable to open file %v", file)
				return err
			}
			defer fileHandle.Close()
			var kitties []cliKitty
			for start <= end {
				kitties = append(kitties, cliKitty{
					//TODO (therealssj): where does this address come from?
					ToAddress: "",
					//TODO (therealssj): where does the extra data come from?
					Kitty: iko.Kitty{
						ID: iko.KittyID(start),
					},
				})
				start++
			}

			enc := json.NewEncoder(fileHandle)
			enc.SetIndent("", "\t")
			err = enc.Encode(kitties)
			return err
		},
	}
}

func editCommand() ikocli.Command {
	name := "edit"
	return ikocli.Command{
		Name:  name,
		Usage: "Mass edit a file containing kitty data",
		Flags: []ikocli.Flag{
			ikocli.StringFlag{
				Name:  "index-range",
				Usage: "Range of kitty indexes to edit",
			},
			ikocli.StringFlag{
				Name:  "field",
				Usage: "Kitty `FIELD` to modify",
			},
			ikocli.StringFlag{
				Name:  "value",
				Usage: "`VALUE` to set for kitty field",
			},
			ikocli.StringFlag{
				Name:  "file",
				Usage: "Read data from `FILE`",
			},
		},
		OnUsageError: OnCommandUsageError(name),
		Action: func(c *ikocli.Context) error {
			indexRange := c.String("index-range")
			start, end, err := ParseIndexRange(indexRange)
			if err != nil {
				return err
			}

			file := c.String("file")
			if file == "" {
				return errors.New("missing input filename")
			}
			// check if input file exists
			if _, err := os.Stat(file); os.IsNotExist(err) {
				return ErrFileNotExist
			} else if err != nil {
				return err
			}

			field := c.String("field")
			if field == "" {
				return errors.New("missing field to be modified")
			}

			value := c.String("value")
			if value == "" {
				return errors.New("missing field value")
			}

			var kitties []cliKitty
			// open in read-write mode as we need to modify the data
			fileHandle, err := os.OpenFile(file, os.O_RDWR, 0600)
			if err != nil {
				log.Errorf("unable to open file %v", file)
				return err
			}
			defer fileHandle.Close()
			err = json.NewDecoder(fileHandle).Decode(&kitties)
			if err != nil {
				log.Error("unable to decode kitty json")
				return err
			}

			for i := range kitties {
				if uint64(kitties[i].ID) >= start && uint64(kitties[i].ID) <= end {
					SetFieldValue(field, value, &kitties[i])
				}
			}

			kittyData, err := json.MarshalIndent(&kitties, "", "\t")
			if err != nil {
				log.Errorf("unable to marshal kitty json")
				return err
			}
			err = ioutil.WriteFile(file, kittyData, 0600)
			if err != nil {
				log.Errorf("failed to write to file %v", file)
				return err
			}

			return nil
		},
	}
}

func finalizeCommand() ikocli.Command {
	name := "finalize"
	return ikocli.Command{
		Name:  name,
		Usage: "Creates a file containing kitty data to be submitted to kitty-api and iko-chain",
		Flags: []ikocli.Flag{
			ikocli.StringFlag{
				Name:  "secret-key",
				Usage: "Secret key to sign transactions",
			},
			ikocli.StringFlag{
				Name:  "file",
				Usage: "Read data from `FILE`",
			},
			ikocli.StringFlag{
				Name:  "output",
				Usage: "Output to `FILE`",
			},
		},
		OnUsageError: OnCommandUsageError(name),
		Action: func(c *ikocli.Context) error {
			secret := c.String("secret-key")
			if secret == "" {
				return errors.New("missing secret key")
			}

			// create a cipher key from hex string
			sk, err := cipher.SecKeyFromHex(secret)
			if err != nil {
				log.Errorf("invalid secret key %s", secret)
				return err
			}

			file := c.String("file")
			if file == "" {
				return errors.New("missing input filename")
			}
			// check if input file exists
			if _, err := os.Stat(file); os.IsNotExist(err) {
				return ErrFileNotExist
			} else if err != nil {
				return err
			}

			outfile := c.String("output")
			if outfile == "" {
				return errors.New("missing output filename")
			}

			var kitties []cliKitty
			// open input file in read-only mode
			fileHandle, err := os.Open(file)
			if err != nil {
				log.Errorf("unable to open input file %v", file)
				return err
			}
			defer fileHandle.Close()

			// read kitty data from input file
			err = json.NewDecoder(fileHandle).Decode(&kitties)
			if err != nil {
				log.Error("unable to decode kitty json")
				return err
			}

			// create a list of database entries
			entries := make([]iko.KittyEntry, len(kitties))
			// create a list of transactions
			transactions := make([]iko.Transaction, len(kitties))

			for i, kitty := range kitties {
				// database entry
				entries[i] = iko.KittyEntry{
					Reservation: "available",
				}
				entries[i].ID = kitty.ID
				entries[i].Name = kitty.Name
				entries[i].Desc = kitty.Desc
				entries[i].Breed = kitty.Breed
				entries[i].PriceBTC = kitty.PriceSKY
				entries[i].PriceBTC = kitty.PriceBTC
				entries[i].BoxOpen = kitty.BoxOpen
				entries[i].BirthDate = kitty.BirthDate
				entries[i].KittyDNA = kitty.KittyDNA
				entries[i].BoxImgURL = kitty.BoxImgURL
				entries[i].KittyImgURL = kitty.KittyImgURL
				entries[i].Sign(sk)

				// gen transaction
				transactions[i] = *iko.NewGenTx(kitty.ID, sk)
			}

			// write encoded data to outfile
			ioutil.WriteFile(outfile, encoder.Serialize(outBin{
				Entries:      entries,
				Transactions: transactions,
			}), 0600)
			return nil
		},
	}
}

func injectCommand() ikocli.Command {
	name := "inject"
	return ikocli.Command{
		Name:  name,
		Usage: "Injects the finalized file data to kitty-api and iko-chain",
		Flags: []ikocli.Flag{
			ikocli.StringFlag{
				Name:  "kitty-api-rpc",
				Usage: "Address in which the kitty-api listens for RPC connections",
			},
			ikocli.StringFlag{
				Name:  "iko-node-rpc",
				Usage: "Address in which the iko-node listens for RPC connections",
			},
			ikocli.StringFlag{
				Name:  "file",
				Usage: "Read data from `FILE`",
			},
		},
		OnUsageError: OnCommandUsageError(name),
		Action: func(c *ikocli.Context) error {

			addrKittyAPI := c.String("kitty-api-rpc")
			if addrKittyAPI == "" {
				return errors.New("missing kitty-api rpc address")
			}

			addrIKONode := c.String("iko-node-rpc")
			if addrIKONode == "" {
				return errors.New("missing iko-node rpc address")
			}

			file := c.String("file")
			if file == "" {
				return errors.New("missing input filename")
			}
			// check if input file exists
			if _, err := os.Stat(file); os.IsNotExist(err) {
				return ErrFileNotExist
			} else if err != nil {
				return err
			}

			var fileData outBin
			// open input file in read-only mode
			fileHandle, err := os.Open(file)
			if err != nil {
				log.Errorf("unable to open input file %v", file)
				return err
			}
			defer fileHandle.Close()

			rawData, err := ioutil.ReadAll(fileHandle)
			if err != nil {
				log.Errorf("unable to read file %v", file)
				return err
			}
			if err := encoder.DeserializeRaw(rawData, &fileData); err != nil {
				log.Errorf("unable to deserialize file data of %v", file)
				return err
			}

			// Inject transactions to iko-node.
			err = func() error {
				client, err := inrpc.NewClient(&inrpc.ClientConfig{
					Address: addrIKONode,
				})
				if err != nil {
					log.Errorf("failed to connect to iko-node via address %s",
						addrIKONode)
					return err
				}
				defer client.Close()

				for i, tx := range fileData.Transactions {
					out, err := client.InjectTx(&inrpc.InjectTxIn{
						Tx: tx,
					})
					if err != nil {
						log.Errorf("failed to inject tx at index %d", i)
						return err
					} else {
						log.Printf("[%5d] OK: %v", out)
					}
				}
				return nil
			}()
			if err != nil {
				return err
			}

			// Inject kitty info to kitty-api.
			err = func() error {
				client, err := kcrpc.NewClient(&kcrpc.ClientConfig{
					Address: addrKittyAPI,
				})
				if err != nil {
					log.Errorf("failed to connect to kitty-api via address %s",
						addrKittyAPI)
					return err
				}
				defer client.Close()

				entryPointers := make([]*iko.KittyEntry, len(fileData.Entries))
				for i, v := range fileData.Entries {
					entryPointers[i] = &v
				}

				return client.AddEntries(&kcrpc.AddEntriesIn{
					Entries: entryPointers,
				})
			}()
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func main() {
	if e := app.Run(os.Args); e != nil {
		log.Println(e)
	}
}

/*
	<< Helpers >>
*/

// ParseIndexRange parses a given index range
// supported format is [2,222] or 5
func ParseIndexRange(indexRange string) (start, end uint64, err error) {
	// regex to match [2,22] or 5 etc.
	r, err := regexp.Compile(`^\[(\d+),(\d+)\]$|^(\d+)$`)
	if err != nil {
		log.Error("failed to compile index range regex")
		return 0, 0, err
	}

	// match indexRange input with above pattern
	match := r.FindStringSubmatch(indexRange)
	if len(match) == 4 {
		if match[0] == match[3] {
			start, err = strconv.ParseUint(match[0], 10, 64)
			if err != nil {
				log.Errorf("unable to convert string %v to uint64", match[1])
				return 0, 0, err
			}
			end = start
		} else {
			start, err = strconv.ParseUint(match[1], 10, 64)
			if err != nil {
				log.Errorf("unable to convert string %v to uint64", match[1])
				return 0, 0, err
			}
			end, err = strconv.ParseUint(match[2], 10, 64)
			if err != nil {
				log.Errorf("unable to convert string %v to uint64", match[2])
				return 0, 0, err
			}
		}
	} else {
		return 0, 0, ErrInvalidIndexRange
	}
	// check that start index is not greater than end index
	if start > end {
		log.Error("start index cannot be greater than end index")
		return 0, 0, ErrInvalidIndexRange
	}

	return start, end, nil
}

/*type Kitty struct {
	ID    KittyID `json:"kitty_id"`    // Identifier for kitty.
	Name  string  `json:"name"`        // Name of kitty.
	Desc  string  `json:"description"` // Description of kitty.
	Breed string  `json:"breed"`       // Kitty breed.

	PriceBTC    int64  `json:"price_btc"`   // Price of kitty in BTC.
	PriceSKY    int64  `json:"price_sky"`   // Price of kitty in SKY.
	Reservation string `json:"reservation"` // Reservation status.

	BoxOpen   bool   `json:"box_open"`    // Whether box is open.
	BirthDate int64  `json:"birth_date"` // Timestamp of box opening.
	KittyDNA  string `json:"kitty_dna"`  // Hex representation of kitty DNA (after box opening).

	BoxImgURL   string `json:"box_image_url"`   // Box image URL.
	KittyImgURL string `json:"kitty_image_url"` // Kitty image URL.
}*/

// SetFieldValue sets value of a kitty field from the tag name
func SetFieldValue(tag string, value interface{}, kitty *cliKitty) error {
	switch tag {
	case "kitty_id":
		if reflect.TypeOf(value) != reflect.TypeOf(kitty.ID) {
			return errors.New("trying to set wrong type of value for Kitty.ID")
		}
		kitty.ID = value.(iko.KittyID)
	case "name":
		if reflect.TypeOf(value) != reflect.TypeOf(kitty.Name) {
			return errors.New("trying to set wrong type of value for Kitty.Name")
		}
		kitty.Name = value.(string)
	case "description":
		if reflect.TypeOf(value) != reflect.TypeOf(kitty.Desc) {
			return errors.New("trying to set wrong type of value for Kitty.Desc")
		}
		kitty.Desc = value.(string)
	case "breed":
		if reflect.TypeOf(value) != reflect.TypeOf(kitty.Breed) {
			return errors.New("trying to set wrong type of value for Kitty.Breed")
		}
		kitty.Breed = value.(string)
	case "price_btc":
		if reflect.TypeOf(value) != reflect.TypeOf(kitty.PriceBTC) {
			return errors.New("trying to set wrong type of value for Kitty.PriceBTC")
		}
		kitty.PriceBTC = value.(int64)
	case "price_sky":
		if reflect.TypeOf(value) != reflect.TypeOf(kitty.PriceSKY) {
			return errors.New("trying to set wrong type of value for Kitty.PriceSKY")
		}
		kitty.PriceSKY = value.(int64)
	case "box_open":
		if reflect.TypeOf(value) != reflect.TypeOf(kitty.BoxOpen) {
			return errors.New("trying to set wrong type of value for Kitty.BoxOpen")
		}
		kitty.BirthDate = value.(int64)
	case "birth_date":
		if reflect.TypeOf(value) != reflect.TypeOf(kitty.BirthDate) {
			return errors.New("trying to set wrong type of value for Kitty.BirthDate")
		}
		kitty.BirthDate = value.(int64)
	case "kitty_dna":
		if reflect.TypeOf(value) != reflect.TypeOf(kitty.KittyDNA) {
			return errors.New("trying to set wrong type of value for Kitty.KittyDNA")
		}
		kitty.KittyDNA = value.(string)
	case "box_image_url":
		if reflect.TypeOf(value) != reflect.TypeOf(kitty.BoxImgURL) {
			return errors.New("trying to set wrong type of value for Kitty.BoxImgURL")
		}
		kitty.BoxImgURL = value.(string)
	case "kitty_image_url":
		if reflect.TypeOf(value) != reflect.TypeOf(kitty.KittyImgURL) {
			return errors.New("trying to set wrong type of value for Kitty.KittyImgURL")
		}
		kitty.KittyImgURL = value.(string)
	case "to_address":
		if reflect.TypeOf(value) != reflect.TypeOf(kitty.ToAddress) {
			return errors.New("trying to set wrong type of value for Kitty.ToAddress")
		}
		kitty.ToAddress = value.(string)
	default:
		return ErrInvalidTagName
	}

	return nil
}

// OnCommandUsageError shows usage error help text
func OnCommandUsageError(command string) ikocli.OnUsageErrorFunc {
	return func(c *ikocli.Context, err error, isSubcommand bool) error {
		fmt.Fprintf(c.App.Writer, "Error: %v\n\n", err)
		ikocli.ShowCommandHelp(c, command)
		return nil
	}
}
