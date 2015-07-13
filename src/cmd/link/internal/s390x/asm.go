// Inferno utils/5l/asm.c
// http://code.google.com/p/inferno-os/source/browse/utils/5l/asm.c
//
//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
//	Portions Copyright © 2009 The Go Authors.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package s390x

import (
	"cmd/internal/obj"
	"cmd/link/internal/ld"
	"debug/elf"
	"fmt"
	"log"
)

func gentext() {
}

func adddynrela(rel *ld.LSym, s *ld.LSym, r *ld.Reloc) {
	log.Fatalf("adddynrela not implemented")
}

func adddynrel(s *ld.LSym, r *ld.Reloc) {
	targ := r.Sym
	ld.Ctxt.Cursym = s

	switch r.Type {
	default:
		if r.Type >= 256 {
			ld.Diag("unexpected relocation type %d", r.Type)
			return
		}

		// Handle relocations found in ELF object files.
	case 256 + ld.R_390_12,
		256 + ld.R_390_GOT12:
		ld.Diag("s390x 12-bit relocations have not been implemented (relocation type %d)", r.Type-256)
		return

	case 256 + ld.R_390_8,
		256 + ld.R_390_16,
		256 + ld.R_390_32,
		256 + ld.R_390_64:
		if targ.Type == obj.SDYNIMPORT {
			ld.Diag("unexpected R_390_nn relocation for dynamic symbol %s", targ.Name)
		}
		r.Type = obj.R_ADDR
		return

	case 256 + ld.R_390_PC16,
		256 + ld.R_390_PC32,
		256 + ld.R_390_PC64:
		if targ.Type == obj.SDYNIMPORT {
			ld.Diag("unexpected R_390_PCnn relocation for dynamic symbol %s", targ.Name)
		}
		if targ.Type == 0 || targ.Type == obj.SXREF {
			ld.Diag("unknown symbol %s in pcrel", targ.Name)
		}
		r.Type = obj.R_PCREL
		r.Add += int64(r.Siz)
		return

	case 256 + ld.R_390_GOT16,
		256 + ld.R_390_GOT32,
		256 + ld.R_390_GOT64:
		ld.Diag("unimplemented S390x relocation: %v", r.Type-256)
		return

	case 256 + ld.R_390_PLT16DBL,
		256 + ld.R_390_PLT32DBL:
		r.Type = obj.R_PCREL
		r.Variant = ld.RV_390_DBL
		r.Add += int64(r.Siz)
		if targ.Type == obj.SDYNIMPORT {
			addpltsym(ld.Ctxt, targ)
			r.Sym = ld.Linklookup(ld.Ctxt, ".plt", 0)
			r.Add += int64(targ.Plt)
		}
		return

	case 256 + ld.R_390_PLT32,
		256 + ld.R_390_PLT64:
		r.Type = obj.R_PCREL
		r.Add += int64(r.Siz)
		if targ.Type == obj.SDYNIMPORT {
			addpltsym(ld.Ctxt, targ)
			r.Sym = ld.Linklookup(ld.Ctxt, ".plt", 0)
			r.Add += int64(targ.Plt)
		}
		return

	case 256 + ld.R_390_COPY:
		ld.Diag("unimplemented S390x relocation: %v", r.Type-256)

	case 256 + ld.R_390_GLOB_DAT:
		ld.Diag("unimplemented S390x relocation: %v", r.Type-256)

	case 256 + ld.R_390_JMP_SLOT:
		ld.Diag("unimplemented S390x relocation: %v", r.Type-256)

	case 256 + ld.R_390_RELATIVE:
		ld.Diag("unimplemented S390x relocation: %v", r.Type-256)

	case 256 + ld.R_390_GOTOFF:
		if targ.Type == obj.SDYNIMPORT {
			ld.Diag("unexpected R_390_GOTOFF relocation for dynamic symbol %s", targ.Name)
		}
		r.Type = obj.R_GOTOFF
		return

	case 256 + ld.R_390_GOTPC:
		r.Type = obj.R_PCREL
		r.Sym = ld.Linklookup(ld.Ctxt, ".got", 0)
		r.Add += int64(r.Siz)
		return

	case 256 + ld.R_390_PC16DBL,
		256 + ld.R_390_PC32DBL:
		r.Type = obj.R_PCREL
		r.Variant = ld.RV_390_DBL
		r.Add += int64(r.Siz)
		if targ.Type == obj.SDYNIMPORT {
			ld.Diag("unexpected R_390_PCnnDBL relocation for dynamic symbol %s", targ.Name)
		}
		return

	case 256 + ld.R_390_GOTPCDBL:
		r.Type = obj.R_PCREL
		r.Variant = ld.RV_390_DBL
		r.Sym = ld.Linklookup(ld.Ctxt, ".got", 0)
		r.Add += int64(r.Siz)
		return

	case 256 + ld.R_390_GOTENT:
		// if targ.Type != obj.SDYNIMPORT {
		// 	ld.Diag("unexpected R_390_GOTENT relocation for non-dynamic symbol %s", targ.Name)
		// }
		addgotsym(targ)

		r.Type = obj.R_PCREL
		r.Variant = ld.RV_390_DBL
		r.Sym = ld.Linklookup(ld.Ctxt, ".got", 0)
		r.Add += int64(targ.Got)
		r.Add += int64(r.Siz)
		return

	}
	// Handle references to ELF symbols from our own object files.
	if targ.Type != obj.SDYNIMPORT {
		return
	}

	ld.Diag("unsupported relocation for dynamic symbol %s (type=%d stype=%d)", targ.Name, r.Type, targ.Type)
}

func elfreloc1(r *ld.Reloc, sectoff int64) int {
	ld.Thearch.Vput(uint64(sectoff))

	elfsym := r.Xsym.Elfsym
	switch r.Type {
	default:
		return -1

	case obj.R_TLS_LE:
		switch r.Siz {
		default:
			return -1
		case 4:
			ld.Thearch.Vput(ld.R_390_TLS_LE32 | uint64(elfsym)<<32)
		case 8:
			ld.Thearch.Vput(ld.R_390_TLS_LE64 | uint64(elfsym)<<32)
		}

	case obj.R_ADDR:
		switch r.Siz {
		default:
			return -1
		case 4:
			ld.Thearch.Vput(ld.R_390_32 | uint64(elfsym)<<32)
		case 8:
			ld.Thearch.Vput(ld.R_390_64 | uint64(elfsym)<<32)
		}

	case obj.R_PCREL, obj.R_PCRELDBL, obj.R_CALL:
		elfrel := ld.R_390_NONE
		isdbl := r.Variant&ld.RV_TYPE_MASK == ld.RV_390_DBL
		// TODO(mundaym): all DBL style relocations should be signalled using the variant.
		switch r.Type {
		case obj.R_PCRELDBL, obj.R_CALL:
			isdbl = true
		}
		if r.Xsym.Type == obj.SDYNIMPORT && r.Xsym.ElfType == elf.STT_FUNC {
			if isdbl {
				switch r.Siz {
				case 2:
					elfrel = ld.R_390_PLT16DBL
				case 4:
					elfrel = ld.R_390_PLT32DBL
				}
			} else {
				switch r.Siz {
				case 4:
					elfrel = ld.R_390_PLT32
				case 8:
					elfrel = ld.R_390_PLT64
				}
			}
		} else {
			if isdbl {
				switch r.Siz {
				case 2:
					elfrel = ld.R_390_PC16DBL
				case 4:
					elfrel = ld.R_390_PC32DBL
				}
			} else {
				switch r.Siz {
				case 2:
					elfrel = ld.R_390_PC16
				case 4:
					elfrel = ld.R_390_PC32
				case 8:
					elfrel = ld.R_390_PC64
				}
			}
		}
		if elfrel == ld.R_390_NONE {
			return -1 // unsupported size/dbl combination
		}
		ld.Thearch.Vput(uint64(elfrel) | uint64(elfsym)<<32)
	}

	ld.Thearch.Vput(uint64(r.Xadd))
	return 0
}

func elfsetupplt() {
	plt := ld.Linklookup(ld.Ctxt, ".plt", 0)
	got := ld.Linklookup(ld.Ctxt, ".got", 0)
	if plt.Size == 0 {
		// stg     %r1,56(%r15)
		ld.Adduint8(ld.Ctxt, plt, 0xe3)
		ld.Adduint8(ld.Ctxt, plt, 0x10)
		ld.Adduint8(ld.Ctxt, plt, 0xf0)
		ld.Adduint8(ld.Ctxt, plt, 0x38)
		ld.Adduint8(ld.Ctxt, plt, 0x00)
		ld.Adduint8(ld.Ctxt, plt, 0x24)
		// larl    %r1,_GLOBAL_OFFSET_TABLE_
		ld.Adduint8(ld.Ctxt, plt, 0xc0)
		ld.Adduint8(ld.Ctxt, plt, 0x10)
		ld.Addpcrelplus(ld.Ctxt, plt, got, 6)
		// mvc     48(8,%r15),8(%r1)
		ld.Adduint8(ld.Ctxt, plt, 0xd2)
		ld.Adduint8(ld.Ctxt, plt, 0x07)
		ld.Adduint8(ld.Ctxt, plt, 0xf0)
		ld.Adduint8(ld.Ctxt, plt, 0x30)
		ld.Adduint8(ld.Ctxt, plt, 0x10)
		ld.Adduint8(ld.Ctxt, plt, 0x08)
		// lg      %r1,16(%r1)
		ld.Adduint8(ld.Ctxt, plt, 0xe3)
		ld.Adduint8(ld.Ctxt, plt, 0x10)
		ld.Adduint8(ld.Ctxt, plt, 0x10)
		ld.Adduint8(ld.Ctxt, plt, 0x10)
		ld.Adduint8(ld.Ctxt, plt, 0x00)
		ld.Adduint8(ld.Ctxt, plt, 0x04)
		// br      %r1
		ld.Adduint8(ld.Ctxt, plt, 0x07)
		ld.Adduint8(ld.Ctxt, plt, 0xf1)
		// nopr    %r0
		ld.Adduint8(ld.Ctxt, plt, 0x07)
		ld.Adduint8(ld.Ctxt, plt, 0x00)
		// nopr    %r0
		ld.Adduint8(ld.Ctxt, plt, 0x07)
		ld.Adduint8(ld.Ctxt, plt, 0x00)
		// nopr    %r0
		ld.Adduint8(ld.Ctxt, plt, 0x07)
		ld.Adduint8(ld.Ctxt, plt, 0x00)

		// assume got->size == 0 too
		ld.Addaddrplus(ld.Ctxt, got, ld.Linklookup(ld.Ctxt, ".dynamic", 0), 0)

		ld.Adduint64(ld.Ctxt, got, 0)
		ld.Adduint64(ld.Ctxt, got, 0)
	}
}

func machoreloc1(r *ld.Reloc, sectoff int64) int {
	return -1
}

func archreloc(r *ld.Reloc, s *ld.LSym, val *int64) int {
	if ld.Linkmode == ld.LinkExternal {
		return -1
	}

	switch r.Type {
	case obj.R_CONST:
		*val = r.Add
		return 0

	case obj.R_GOTOFF:
		*val = ld.Symaddr(r.Sym) + r.Add - ld.Symaddr(ld.Linklookup(ld.Ctxt, ".got", 0))
		return 0
	}

	return -1
}

func archrelocvariant(r *ld.Reloc, s *ld.LSym, t int64) int64 {
	switch r.Variant & ld.RV_TYPE_MASK {
	default:
		ld.Diag("unexpected relocation variant %d", r.Variant)
		return t

	case ld.RV_NONE:
		return t

	case ld.RV_390_DBL:
		if (t & 1) != 0 {
			ld.Diag("%s+%v is not 2-byte aligned", r.Sym.Name, r.Sym.Value)
		}
		return t >> 1
	}
}

func addpltsym(ctxt *ld.Link, s *ld.LSym) {
	if s.Plt >= 0 {
		return
	}

	ld.Adddynsym(ctxt, s)

	if ld.Iself {
		plt := ld.Linklookup(ctxt, ".plt", 0)
		got := ld.Linklookup(ctxt, ".got", 0)
		rela := ld.Linklookup(ctxt, ".rela.plt", 0)
		if plt.Size == 0 {
			elfsetupplt()
		}
		// larl    %r1,_GLOBAL_OFFSET_TABLE_+index

		ld.Adduint8(ctxt, plt, 0xc0)
		ld.Adduint8(ctxt, plt, 0x10)
		ld.Addpcrelplus(ctxt, plt, got, got.Size+6) // need variant?

		// add to got: pointer to current pos in plt
		ld.Addaddrplus(ctxt, got, plt, plt.Size+8) // weird but correct
		// lg      %r1,0(%r1)
		ld.Adduint8(ctxt, plt, 0xe3)
		ld.Adduint8(ctxt, plt, 0x10)
		ld.Adduint8(ctxt, plt, 0x10)
		ld.Adduint8(ctxt, plt, 0x00)
		ld.Adduint8(ctxt, plt, 0x00)
		ld.Adduint8(ctxt, plt, 0x04)
		// br      %r1
		ld.Adduint8(ctxt, plt, 0x07)
		ld.Adduint8(ctxt, plt, 0xf1)
		// basr    %r1,%r0
		ld.Adduint8(ctxt, plt, 0x0d)
		ld.Adduint8(ctxt, plt, 0x10)
		// lgf     %r1,12(%r1)
		ld.Adduint8(ctxt, plt, 0xe3)
		ld.Adduint8(ctxt, plt, 0x10)
		ld.Adduint8(ctxt, plt, 0x10)
		ld.Adduint8(ctxt, plt, 0x0c)
		ld.Adduint8(ctxt, plt, 0x00)
		ld.Adduint8(ctxt, plt, 0x14)
		// jg .plt
		ld.Adduint8(ctxt, plt, 0xc0)
		ld.Adduint8(ctxt, plt, 0xf4)

		ld.Adduint32(ctxt, plt, uint32(-((plt.Size - 2) >> 1))) // roll-your-own relocation
		//.plt index
		ld.Adduint32(ctxt, plt, uint32(rela.Size)) // rela size before current entry

		// rela
		ld.Addaddrplus(ctxt, rela, got, got.Size-8)

		ld.Adduint64(ctxt, rela, ld.ELF64_R_INFO(uint32(s.Dynid), ld.R_390_JMP_SLOT))
		ld.Adduint64(ctxt, rela, 0)

		s.Plt = int32(plt.Size - 32)

	} else {
		ld.Diag("addpltsym: unsupported binary format")
	}
}

func addgotsym(s *ld.LSym) {
	if s.Got >= 0 {
		return
	}

	ld.Adddynsym(ld.Ctxt, s)
	got := ld.Linklookup(ld.Ctxt, ".got", 0)
	s.Got = int32(got.Size)
	ld.Adduint64(ld.Ctxt, got, 0)

	if ld.Iself {
		rela := ld.Linklookup(ld.Ctxt, ".rela", 0)
		ld.Addaddrplus(ld.Ctxt, rela, got, int64(s.Got))
		ld.Adduint64(ld.Ctxt, rela, ld.ELF64_R_INFO(uint32(s.Dynid), ld.R_390_GLOB_DAT))
		ld.Adduint64(ld.Ctxt, rela, 0)
	} else {
		ld.Diag("addgotsym: unsupported binary format")
	}
}

func asmb() {
	if ld.Debug['v'] != 0 {
		fmt.Fprintf(&ld.Bso, "%5.2f asmb\n", obj.Cputime())
	}
	ld.Bso.Flush()

	if ld.Iself {
		ld.Asmbelfsetup()
	}

	sect := ld.Segtext.Sect
	ld.Cseek(int64(sect.Vaddr - ld.Segtext.Vaddr + ld.Segtext.Fileoff))
	ld.Codeblk(int64(sect.Vaddr), int64(sect.Length))
	for sect = sect.Next; sect != nil; sect = sect.Next {
		ld.Cseek(int64(sect.Vaddr - ld.Segtext.Vaddr + ld.Segtext.Fileoff))
		ld.Datblk(int64(sect.Vaddr), int64(sect.Length))
	}

	if ld.Segrodata.Filelen > 0 {
		if ld.Debug['v'] != 0 {
			fmt.Fprintf(&ld.Bso, "%5.2f rodatblk\n", obj.Cputime())
		}
		ld.Bso.Flush()

		ld.Cseek(int64(ld.Segrodata.Fileoff))
		ld.Datblk(int64(ld.Segrodata.Vaddr), int64(ld.Segrodata.Filelen))
	}

	if ld.Debug['v'] != 0 {
		fmt.Fprintf(&ld.Bso, "%5.2f datblk\n", obj.Cputime())
	}
	ld.Bso.Flush()

	ld.Cseek(int64(ld.Segdata.Fileoff))
	ld.Datblk(int64(ld.Segdata.Vaddr), int64(ld.Segdata.Filelen))

	/* output symbol table */
	ld.Symsize = 0

	ld.Lcsize = 0
	symo := uint32(0)
	if ld.Debug['s'] == 0 {
		// TODO: rationalize
		if ld.Debug['v'] != 0 {
			fmt.Fprintf(&ld.Bso, "%5.2f sym\n", obj.Cputime())
		}
		ld.Bso.Flush()
		switch ld.HEADTYPE {
		default:
			if ld.Iself {
				symo = uint32(ld.Segdata.Fileoff + ld.Segdata.Filelen)
				symo = uint32(ld.Rnd(int64(symo), int64(ld.INITRND)))
			}

		case obj.Hplan9:
			symo = uint32(ld.Segdata.Fileoff + ld.Segdata.Filelen)
		}

		ld.Cseek(int64(symo))
		switch ld.HEADTYPE {
		default:
			if ld.Iself {
				if ld.Debug['v'] != 0 {
					fmt.Fprintf(&ld.Bso, "%5.2f elfsym\n", obj.Cputime())
				}
				ld.Asmelfsym()
				ld.Cflush()
				ld.Cwrite(ld.Elfstrdat)

				if ld.Debug['v'] != 0 {
					fmt.Fprintf(&ld.Bso, "%5.2f dwarf\n", obj.Cputime())
				}
				ld.Dwarfemitdebugsections()

				if ld.Linkmode == ld.LinkExternal {
					ld.Elfemitreloc()
				}
			}

		case obj.Hplan9:
			ld.Asmplan9sym()
			ld.Cflush()

			sym := ld.Linklookup(ld.Ctxt, "pclntab", 0)
			if sym != nil {
				ld.Lcsize = int32(len(sym.P))
				for i := 0; int32(i) < ld.Lcsize; i++ {
					ld.Cput(uint8(sym.P[i]))
				}

				ld.Cflush()
			}
		}
	}

	ld.Ctxt.Cursym = nil
	if ld.Debug['v'] != 0 {
		fmt.Fprintf(&ld.Bso, "%5.2f header\n", obj.Cputime())
	}
	ld.Bso.Flush()
	ld.Cseek(0)
	switch ld.HEADTYPE {
	default:
	case obj.Hplan9: /* plan 9 */
		ld.Thearch.Lput(0x647)                      /* magic */
		ld.Thearch.Lput(uint32(ld.Segtext.Filelen)) /* sizes */
		ld.Thearch.Lput(uint32(ld.Segdata.Filelen))
		ld.Thearch.Lput(uint32(ld.Segdata.Length - ld.Segdata.Filelen))
		ld.Thearch.Lput(uint32(ld.Symsize))      /* nsyms */
		ld.Thearch.Lput(uint32(ld.Entryvalue())) /* va of entry */
		ld.Thearch.Lput(0)
		ld.Thearch.Lput(uint32(ld.Lcsize))

	case obj.Hlinux,
		obj.Hfreebsd,
		obj.Hnetbsd,
		obj.Hopenbsd,
		obj.Hnacl:
		ld.Asmbelf(int64(symo))
	}

	ld.Cflush()
	if ld.Debug['c'] != 0 {
		fmt.Printf("textsize=%d\n", ld.Segtext.Filelen)
		fmt.Printf("datsize=%d\n", ld.Segdata.Filelen)
		fmt.Printf("bsssize=%d\n", ld.Segdata.Length-ld.Segdata.Filelen)
		fmt.Printf("symsize=%d\n", ld.Symsize)
		fmt.Printf("lcsize=%d\n", ld.Lcsize)
		fmt.Printf("total=%d\n", ld.Segtext.Filelen+ld.Segdata.Length+uint64(ld.Symsize)+uint64(ld.Lcsize))
	}
}
