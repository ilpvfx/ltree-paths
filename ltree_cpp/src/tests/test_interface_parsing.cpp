#include <catch2/catch_test_macros.hpp>

#include <ltree/parse.hpp>

TEST_CASE("parsing_ValidLabelCodepoint") {
    char32_t should_be_accepted[] = {
        // The two non-alphanumerical codepoints allowed:
        '-', '_',
        // Some examples from latin
        '1', '9', 'a', 'A', 'z', 'Z',
        // Non-ASCII codepoints
        U'\x1D582', // Lu, Mathematical Bold Fraktur Capital W
        U'\xA640',  // Lu, Cyrillic Capital Letter Zemlya
        U'\x05D4',  // Lo, Hebrew Letter He
        U'\x1F88',  // Lt, Greek Capital Letter Alpha With Psili And
                    // Prosgegrammeni
        U'\x216B',  // Nl, Roman Numeral Twelve
        U'\x0D78',  // No, Malayalam Fraction Three Sixteenths
        U'\x1D2F3', // No, Mayan Numeral Nineteen
        U'\xA9F1',  // Nd, Myanmar Tai Laing Digit One
    };

    for (auto c : should_be_accepted) {
        CAPTURE(c);
        CHECK(ltree::ValidLabelCodepoint(c));
    }

    char32_t not_acceptable[] = {
        // ASCII control, punctuation etc.
        '?', '!', '\0',
        // Unicode non-alphanumerics
        U'\x00AB',  // Pi, Left-Pointing Double Angle Quotation Mark
        U'\x12470', // Po, Cuneiform Punctuation Sign Old Assyrian Word Divide
        U'\xFFE5',  // Sc, Fullwidth Yen Sign
        U'\x2003',  // Zs, Em Space
        U'\x20DE',  // Me, Combining Enclosing Square
    };

    for (auto c : not_acceptable) {
        CAPTURE(c);
        CHECK(!ltree::ValidLabelCodepoint(c));
    }
}

// UTF-8 does not encode surrogate codepoints (they are not Unicode scalar
// values), and they should not be accepted. This is tested in particular
// because a lot of software DOES NOT handle this correctly.
TEST_CASE("parsing_Utf8ValidLabel_unicode_surrogates") {
    // For each block we just test one of the characters.
    // U+D800 - U+DB7F High Surrogates Block (BMP)
    // 11101101 10101100 10000001
    char const *db01 = "abc_\xED\xAC\x81_abc";
    auto result1 = ltree::Utf8ValidLabel(db01);
    CHECK(result1.code == ltree::Error::BAD_UNICODE);
    CHECK(result1.error_column == 4);

    // U+DC00 - U+DFFF Low Surrogates Block (BMP)
    // 11101101 10110000 10000001
    char const *dc01 = "abc_\xED\xB0\x81_abc";
    auto result2 = ltree::Utf8ValidLabel(dc01);
    CHECK(result2.code == ltree::Error::BAD_UNICODE);
    CHECK(result2.error_column == 4);

    // U+DB80 - U+DBFF High Private Use Surrogates Block
    // 11101101 10101110 10000001
    char const *db81 = "abc_\xED\xAE\x81_abc";
    auto result3 = ltree::Utf8ValidLabel(db81);
    CHECK(result3.code == ltree::Error::BAD_UNICODE);
    CHECK(result3.error_column == 4);
}

// In these cases either the leading byte of a multibyte sequence or a following
// byte is bad. UTF-8 terminology: each byte is a code unit.
TEST_CASE("parsing_Utf8ValidLabel_unicode_invalid_utf8_continuation") {
    // The control character 0x80 is the smallest codepoint that won't fit in
    // one UTF-8 code unit. It encodes to UTF-8 as 0xC2 0x80 (happy
    // coincidence).
    char const correct_sequence[] = {(char)0b11000010, (char)0b10000000};
    auto result1 = ltree::Utf8ValidLabel(correct_sequence);
    // However, it is not alphanumeric, so we must reject it with the right
    // message.
    REQUIRE(result1.code == ltree::Error::INVALID_CHARACTER);

    // Let's destroy it in various ways.
    // Bad second byte: top two bits should be 10.
    char const bad_second_byte[] = {(char)0b11000010, (char)0b00000000};
    // Bad first byte: top two bits should be 11.
    char const bad_first_byte1[] = {(char)0b10000010, (char)0b10000000};
    char const bad_first_byte2[] = {(char)0b10000010, (char)0b10000000};

    // Overlong sequences
    // 0xC0 indicates 6 bits in the trailing byte, but this should be encoded as
    // 1 byte and this byte is disallowed.
    char const *impossible_byte1 = "\xC0";

    // 0xC1 indicates a 7 bit codepoint, (with a MSB of 1), but this should be
    // encoded as 1 byte and this byte is also disallowed.
    char const *impossible_byte2 = "\xC1";

    // Encode 'Ā' (0x100) as an unnecessarily sequence. This should be caught by
    // the above.
    char const *too_long1 = "\xC0\x84\x80";

    char const *bad_cases[] = {
        bad_second_byte,  bad_first_byte1,  bad_first_byte2,
        impossible_byte1, impossible_byte2, too_long1,
    };

    for (char const *sequence : bad_cases) {
        CAPTURE(sequence);
        auto result = ltree::Utf8ValidLabel(sequence);
        CHECK(result.code == ltree::Error::BAD_UNICODE);
    }
}

// The decoded code units point to a "codepoint" that is not a Unicode scalar
// value, and not a surrogate (tested elsewhere), i.e a value above U+10FFFF.
TEST_CASE("parsing_Utf8ValidLabel_unicode_invalid_codepoint") {
    // This is larger than the max Unicode codepoint U+10FFFF. It should be
    // rejected:
    // - because it unassigned
    // - it is not alphanumeric
    char const *too_large =
        (char const[]){(char)0b11110101, (char)0b10001111, (char)0b10111111,
                       (char)0b10111111, 0};
    auto result2 = ltree::Utf8ValidLabel(too_large);
    CHECK(result2.code == ltree::Error::BAD_UNICODE);
}

TEST_CASE("parsing_Utf8ValidLabel") {
    char const *good_cases[] = {
        "abc-def", "abc_def", "ab1_2dc", "-", "_", "123",
    };

    for (auto sequence : good_cases) {
        auto result = ltree::Utf8ValidLabel(sequence);
        CHECK(result.code == ltree::Error::SUCCESS);
    }

    struct BadCase {
        char const *input;
        std::size_t expected_error_column;
        ltree::Error::ValueType expected_code;
    };

    BadCase const bad_cases[] = {
        {"", 0, ltree::Error::EXPECTED_LABEL},
        {"label_with space", 10, ltree::Error::INVALID_CHARACTER},
        {"abc\xEF\xBF\xA5_def", 3,
         ltree::Error::INVALID_CHARACTER}, // Fullwidth Yen Symbol
    };

    for (auto [sequence, expected_column, expected_code] : bad_cases) {
        CAPTURE(sequence);
        auto result = ltree::Utf8ValidLabel(sequence);
        CHECK(result.code == expected_code);
        CHECK(result.error_column == expected_column);
    }
}

TEST_CASE("parsing_Utf8ValidPath") {
    // This test is a bit lightweight, we rely on the comprehensive test suite
    // for Path to deal with many of the special cases. It's the same parsing
    // backend.
    char const *good_cases[] = {
        "",
        "one.path.four.labels",
        "kitchen123-sink_2000",
    };
    for (auto sequence : good_cases) {
        CAPTURE(sequence);

        auto result = ltree::Utf8ValidPath(sequence);
        CAPTURE(result.code, result.error_column, result.label_count);

        CHECK(Ok(result));
    }

    char const *bad_cases[] = {
        "spaces not_allowed",
        "abc123.??wat",
    };
    for (auto sequence : bad_cases) {
        CAPTURE(sequence);

        auto result = ltree::Utf8ValidPath(sequence);
        CAPTURE(result.code, result.error_column, result.label_count);

        CHECK(!Ok(result));
    }
}