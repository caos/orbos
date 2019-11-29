package test_test

/*
import (
	"errors"
	"testing"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/core/operator/test"
)

func TestChaniningDesired(t *testing.T) {
	desire, err := test.DesireDefaults(t, "v0.0.0")
	if err != nil {
		t.Fatal(err)
	}

	testValue := "chainingDesiredWorks"
	testKey := "configId"
	desire = desire.Chain(func(des map[string]interface{}) {
		des[testKey] = testValue
	})
	testMap := make(map[string]interface{})
	desire(testMap)
	if property, ok := testMap[testKey]; !ok || property != testValue {
		t.Fatalf("expected %s to be %s", property, testValue)
	}
}

func TestParsingCurrent(t *testing.T) {
	testError := "testerror"
	marshalled := test.MapIteration(&operator.IterationDone{Error: errors.New(testError), Current: []byte(`current:
  kind: orbiter.caos.ch/Orbiter
  state: {}
  version: v1
deps:
  clusters:
    prod:
      current:
        kind: orbiter.caos.ch/KubernetesCluster
        state:
          computes:
            testkubernetesprodgooglebelgium-centos-1c3atmnzs59wlavleny1n7oz:
              metadata:
                group: lowcost
                pool: centos
                provider: googlebelgium
                tier: workers
              software:
                current:
                  kind: nodeagent.caos.ch/NodeAgent
                  state:
                    open:
                      kubelet:
                        port: 10250
                        protocol: tcp
                    ready: true
                    software:
                      containerruntime: docker-ce v18.09.6
                      kubeadm: v1.15.0
                      kubectl: v1.15.0
                      kubelet: v1.15.0
                      swap: disabled
                    version: v0.0.0
                  version: v1`),
		Secrets: []byte(`
googleApplicationCredentialsValue:
  encoding: Base64
  encryption: AES256
  value: iQjq0YN1ik5cV_JevHMqFCvvs9HqNHb1h7mZKd62sc1xxs-Q2tw3n8ITOPZWXpSSKSYi4KNGrkQycD55FbSxQ-fX9PY6tdrlgK1nE2HRym73x0y73dKkLFQTZe56Kr--uldQ9uWwR2bLTZxx3YYc3eLxGUc4XWZYXdIYJQir79fyUdCB47Rj1tUP1W6oCKZWt8pEAv6SVjMXBVxkUCpwm4dMCdn-X2cZALUx-zIYpoFPhGZyW9eYAPILnMt7mGT5c8WgroFMz3zUDOG8gRzVcHgJEqCznNKUv5iI_TJqe5lp01nzNmvNwse0VWACSpDtz-DgsjlNrEjX4EVNKPkL_YIVopc2DXuuXl-9BXGyl3Sz1mjtxNVwevJ-Zrt-tZX_eT75HK3pp15JBRMafccW0wo2XX4cIJqDSfaMHQC4aDDtm94CD7MsTTbFLPDR1UuDqRG1c8MSv8x-8FO2K2voweydl9D4O2SugQQTnX3AZsPKSO0A7kKCoRDaGUPFgUw4eidn6OAZ5sJSi1j3ePsczaxMIrG15oGEQEZfTFTwvvy4wo23viQUNhM2JzMl8Tx6gHvwLwDmUEvfjbKN7MmqiFDM9a2-2MInnalvEPY7CnOeQbiPOupnbNMNM9HRNWTDy_oi5nhtDbx5kHvUglXNWI9oQD7jdb7G1xOVsuhMCyYY-dtmZVsqYqffg3iETwq0iiE8Y0qBY_XUNw8_Yeo8m3RlV3TohU2yztcnsqH0HMMpl4vFin7wGNPkG-TvqLUUvEr6XtWRB7nXsAMx9l4S4pHGmPnD1kdJ1pcIBYfeeDWuT5eERN_gIxGt_oYdYzdhdgxp5mH2unryQPOwDIn33-ldHYw68W2urP3yypOlOJtdd3TY6_rKCeJZPwe7tTi6RS2nS0pfKiTOp2NDzMMfc-ubvfnMZuX1JynkMIfBwbaCo74cJVmqKW5XcFwOeXqqetx4Bt-IWh2t7p8EgyAXqHucs-x8FSuHfITHaFlYJDyxLpsGEthPUsPDAgT3N1SUKkaHvfnd7UM-_0KkY9g0wggcDiVu4QSzhZn0I16qL49GLcAyKAn_QvVZyTwLluj4Bv91hMcjc5dD5FdKy01I9bIJR52D1ZaL-kDBwTTsMaw7eR3HEg48WJ7c9cpbWLPnDHi-8q9O_RJLONRlwSdCdBKvR-PviYqi1GCSojVtlAjwRmn9qnRfeLlsqm1ST4GNmKGPMzeS8apN4y4pyt_69vBX1Ys3WzYxlh479xfO8eoOdhA7Xu_ZsryclkM-fe751GK2jELFLtlX5I5uYDE8sWiKON0Oo8Z-t3yXYi6cTfFSW8lQbCoyjwq1LqnI5l_D9uBkVqcwsB0fgtvjwNPsmAI7p7_kSc-M4ndO07RpoLX2Ixp2gW16tb7gJg83JY6-JXxFNDBH7mqZSRZLmuXzn5VwaAJf4V_1kxVqtC2Efx2XLbzcrvBKUboEOnWvRf0wW0o-4UlRAcUE17o0dzsD6O7ZHEALtZ4kkr0Skf4uPV6jcliYZsLbaVmKcFuI6uRg7E90dnlMQDw5nqvckxMNFnR_Uhbq0D6e0lnmWNFQyFTG9mewCMySplDd0nWKn-K69_PpP69fGof15g-mTRvCM88Dmpbpb8vuyETgzTQ19fKSY-DLia5nsLTPg45TcFRQdcLGge3k0D6nTZGul5dvajDE-alaocc7QIYB3ISKsL14BIQzBAO3zzPK-yjUlTaIuXpg007ZcOfPdIC2omiu_6NhKiwKtgwhMrwLoe1aN8nz_Io-mSPI4LEr7vX03INyDzPNaVmnBB2hdFvpmjyLP4D0PPyd6o8YGGm4nS0M7jcHTd6YPSRhIe7SAtmZahHwPVhYFARxeF9CPZ8-ymAM9NuDlzD2uit_z7Cu3CFdDrt3O4_uEHnRuc9vUlGSgNjnU03sbowuGOb3XEO8NlxPCKkmv-paJvnIKU3lBt4jZeSbcwwLZqrP-I9y8OrW0akBPUK5ZFHGaFDItNlFplgRg_rH2SL4Hn1T3kk1HgfwUV_qi4-4nB85kVuImsFxQjkmZiNRebFKINT1WujP1sD_HCv-diAdIEDZk79N9h1N-dNNiVonhWin_t_QKmnZ5NPTH6MjzeA3nX_bW6Z8g47ytEFjEBw3D2bFVEGCaGc0y7-N2fkn505wzRoUstgfMICKVNWkPZaHLv6tALYXnIHwJqKhNlo-9FzFFR_FIu-3wcC0TvyRljb2mzHI-3nD2_RAPXj6Fi_KjTEyMFa0nGzyMDkiaq5crMVa8asoWOqd8cuYm0Fd7tNL2-8Ufqf6r549XpFMb0kHbA-8rQ_6I4BzkU_Zoyj4e_sH8g76JMH66G8Ta-VOOMInPdMoXcj8G8XxNZgr88Ut48Dnp2pnzoEvhBJILnDyUy4LxUKnr1u73AfMNycdOs21KtQJcmirQLZ_gJ5mIm2nrgeabU6iIh4il1XqZhH_ckG9CDakm6bhddNg1m3qK6Si_AO5ZPi6gQuTBkwinsYcGD65OqswlSp8lz1tyDzRQupbAQ_PAUKoMDVxbYUP0JoyGSPICzUPh7TxaELxC5DdgTeQ8d2bkA0H7pe-e7aP3zVhvMgxyhT-NZcQ-vOW_q2WlaKt8ttzg80XbilsEiLCIuZSMNgL0LAbB-q6mK6_kIrgqri41WjFTmchOgTbNLWKnlWJiw0uJ0LdQocbJyTx0W_eXnF_K22EoyF686sIUbpvPquXMM2AcMLvx1loPz5q5pH4DI3XNWeLioiqCRRG_9JDz94sUhU2Nq3zi8qv7pOozRkVa5CupjYzoT7mwwAfZuR0PTqQ9hfRKIl1W5lZRTGJ68FPvdU7P7S61cTj_YArV7M3kbedCiv6gt64h446rc6SqvDrS0SF-x2Z6VA5eDw7Y5wajsCKy4gLQFw2b2LC7oNABNo8cP1naclQ429-0cKZXxt_gvBWDx7YMZtL-x1fYRYUeGAJ62AkEeBLCfYLs19d6JzC-koW7fjWpF-V2o7gQJeBAoUuGmHKtBesiJcGABq0fCVvr3iLYh7WbLhwUYVTO1fGcU7s3b7FohUuWuJaExy46brmU5zbQO9opy_yqJJ1RpDbvNSbStU5wxTneF0=
instancesSshKey:
  encoding: Base64
  encryption: AES256
  value: OaBG3UzdXXVC32berzDVUP6A9R7ovtFn4Vz3HSXGVYajX_3xzpi7WXCoBI3NjvipOSn404gpx0vmi399PFvrAiRKYga5RTEI6LNxOF4UHIUpqtRi7eQzvB3bgnju_yC8egxbKGZecsTA29npTWpTpxwhTj_Zg2Y9ngJ9DkQ_9MDAv5Tg1jlG_mCnpZPrqICv0MsjZxQtO_R2mPQctDcE6lFXRVywuW_qXViRUgZDHF81m6O8sbydCDPjtLUwGYGmInlyfMKR9LwwuTTCovFaVZWRuk84uvsGDdkJuTgLj0BPeJprRUCEjfObXuQRBr_mB1cDLkcFI6NV0osPPXLfde9Os3xPuO5tknCYM5S7Zn3kvIawBO815Zq2K8cZHf7mtA3RVB8BoFdETxRpEzs2_BoDsEa631w6v2SAJ7v7IklktcxVt-_hfgl-hXvSYizNvjhcH990Nxh-Rb5jLFKzQVFfsGM2AIEJGiR1XOsqU9RkuSmV9hCT0GG3_oYn3-PXPubiVycIbomz7l3RbuLRrwBHl69jptH-yIv2SQAxuymYPKAjHVUzpIwYZzhRBeO9esXylIBRFp_HzV9YeP8yLLlZJ0eFukfmDAoPrbJZZMroKiSJnMEE0OujioDOjGdtXHJKgunOxHsZRodU7iPZMS7FgOyo8r8vXxErIlc7c6zw_h4UJ2jGZHnwO-IleaJwSaqFGblGpMsjXPKNn_v01qx9Pg2yuoWCaKDzlUzlwOpVYcvihHzO1fOxqVHfkGkfuSWbNSTJNbQKEPyYL0F2VnVBWyKX2rVoeGQKAlZboXHKwG0SLoxWBWod5zP1zb1tW17j4aUPXfm1UW_cTvyYotw738VDb8QNQK_Y2yBQ83xnBSH1JODORoNiTF28f8RRspYRS5PTXzZnU3kz2kbTwGfcLTHUubgaLEFTtpb6UwWEEYrywhGvjLgm9Uvk6Au9HEw1PYquJh6GTd3tqSOQamnL8zzEWswhXwXWs01HkD_BrqqbVfhohEqcOSKVQrZmzc5IcxDZ3JpPJn6MW9V-QfFtkHxpthNXDp1cjaIrvqW6xXbYo78opvBV7jzR8jlQHqYyD6oBWeRFZQVjd2nBuZf_lCP9IM3JYbXf4yGD7WmiL6myq1ZlP8JjZ57Rtqx-4R7f7Zccu654TkiqrHph1iEW-49b6ZvkovnQUWKohqIWKeeJBW_7eyyTP5s5QVO0Kf7UOK_lYiABeHWsaCUHryEB1rXKjRShT-Sti6fXoNmwzv1sMsoQVTGR9tMWU6qHiGdGZqdkZVubJehFMrbdAR-fh3I-Dfwu4jIqKpiXtAKS8_tNXFREa5_rJ5a12EboLGSwxi4tjoRZtv0kDdh4rC96jk0s6yFrgJ38NCXRl1fVTdR7tfkbwmi-WyXtA7yNP3h_DamPqerPVkHSeeG_RpxCigY4OHoprpnMPREL4xs2Fh41lOVZmpsD-ObbpAIUArzQiw3xPRcEBUW4-kgNwnWDuHul85dBn-rK9M8Ef76VIIecwVNn4b8s__X1yTovP3P3DtUHDx-UeF26OoJ_SvcaDxwHtBoj6entFzJD32Lp8jB5FxWdVIAJb2o5mwyLDhnsGxN3N-Vpm0SVOoJFPgYTmjty_5qb5siMxUi9AWEVIMphto6f6kPY3rmbguV5PrH-6BsHiGq0xmdSiSATKXDhxu4YayWkB4fQyHbc18KthmO74I9BnSmz6OTmPgbG0lHLz_K4QdWeBFyxtu54eQNMOCzGWbW6wRFXNbOdSxM-C36UADd7pj2PoINhK77Uhn--D6Z21hZKjm1vRSEh8HfErhuZvi8gGj-4Eaxj5e5iGI8ShGnPPBpiAlsRkTX1YK3aZswKBG36G87z4MSSgiof96ZNQ56pSdTYXG_MGU4fHVPIIEpmyp0QU21xpkcQ-oa7Yhi-nBeBfCqgOnCoj0uRV78QgMCQ8Txe8QAhk9Syh74CqRXEqf3eM5wE3p3ZEC5dW_JMzJTaSdtEf5Mv8udMFTmx2R3onGwssz8SwC2CAh04RU7SNvnk6z_ElnlXyVbqxJ-C6fD71DUkLPwq7EVGGdJgjJJa3XGhj26KMNwjLAexQSPFvwYvd3dxcuLh3Bl3SWV5txY31A49iljvUVFBOjaP8VrdgHOm55lEG4S-KGVXlY6bavEf7qmVbP3Z3EakJ1Jm8kFXT3PwST_dT4hMaFKGWPUUL-KxKMu5y8xjxQhAXkZwK0fFhgyGHZ3kxBZ4Aig9964UvvbNqb6kXUasSIqhrd5s0tSJH5bUmoe2tDahAJqmq6D_jE4iyIfzMQt2uUZV_jt7ODRqPP4wvBqze4eFUfmabklO8YteGa19lmjiV9yYbk3UyqbDr-w_lwhO_MaUz5pl53Hf3K_I_KwIzt9LuMKws7TJrSCp7nkMXo9W01pVzZEMLwOEeA_2ytu9xr3zJ-jEwT6wh1MavDjqeUO9emL8B0yRbETpxJLjrzlJh96kZ0WnlfNR3mrPr7QxxQ3ktL5CSPEIvHiJvRfxWraQJkjzJRNmNMWa47i3pXz5ovC6fk_EEXunKSb6us9Q8cI7QA1_cgd6tcIGs36oYro5MdNW91lo4whalz3N6dH1EcYLV0_L8GFzDY7PYIc8QMRqiy9yg8c90RTGp64RVMGzd1HlnPBK2C7gyMb5xTAoQOADziUV6FCzKjKbxpyxZFFfZ5CESFNxqZCjC4y5WN8FE2xzWUJ0N6p3Numq7_PxyK3Z5q5TemX63CbjS3T1kqx_-C3VsMn1w6oW_PBl2EGkJHzd_uAjVcEfliVd3-wypJkhVk-AXZyN5bltHMKDMd5AewGAxP76wdUg989t0fOAUq8RuxVP1U45PecT7O1LhUp44IAy5BuAIny62gOaPJZV0NOxHaERSrRkt0BPmfSfW22qNK9S3Nu69hAGyx_Hf8IpMAFfHJfteiRy-Y73Qq1R2zLnf5HsKyYqpQa1YX5sHutCfUWhNuEXa_j-NJjEXHdAztQQU25H1r3yCS6bWrLc0ExgN1zUsIAZaDSuNm_uSo-zGjzIoKYLoI88yfKT8r5lKuIYlwHEirpGTFJDpSp62eGEK7Z0SdIDUxsusosz2M8x5wCdpiS14TSYTsxZSf2fpo9_Du7OS9zS-5A1_I0W6VTMhAdQ3QtNUqINGDGYML2QuKO4NRGcACe_zfjtZlL75NurlgRFzB3eaGlqsBGWFg7H1U_JprNw_lnq-6FHc_q8cqV55wGhg3av5s4-VBGSZCiUcsa2McqxfReKCxpEZI4pDTF_V3eJ8z6UU7fYSc9_DX-ecux3ZaqXkP63-ABEt3kaTLesYbH6GEB2a3u9Ck4wk4WCHVfSH3bymp2ivrWQzfG8qVyCIVOchnxU66R7NDkzcy6mR2lv7PuY_N78UIO50h_l9rimQ9EunRnqWiXlfjROVLqNkBAyxlpZlRTkQPu8UM6LwGi2vfCdYnJbA06XLajzE9nIFRiPL2WNKjvXqtmB95kQ9BwVOZPzpb-9vTlR4ClNLQHTnakb0uL19rr55MjiyS6gRnDfP-3PPoFmBenl9kCd3r8TJNGKCANM7ToO6i-NNwbLeoQFFu6peSkZ6hH999zQnZbs-E8R2CvtnwaoCh2Oc8JWcJPzn8LFxlA0dSAi6SxMwFfk8PFC3JVZQFJpTJ5LQ35DHS5t5oH5YdeFFVyl8kH2l-7GXD0uILgbt3G2jea13-ZY5ad0aJQYVUkQ5SewvulLFkgaTVXN-fnU4U77n276w3AODVxTXmir0kPX_RHytXr3Q0hgo6E--zrifNfvzzO8PU0vdoKTgBi6NEhQika4Rmc3ah1c_KbC8F60kgBghJucaEP4fT_9VmWmrqkHQSueMgwouci3qWw3Ke69ODvrokkEGrIh-qTwhtuTi1PZ6DcGEsJipbc35859X8rcff8XeO8uzQyHAYPnD9z_sEiINZ0Lhr4RWOILUH0ba6bt43V5gKqit0LnBTLiH5uz2wlsdA2pvy56gf98mDRcf-tiXaoA5tWIUCxFtLwrdeGvJl2ylgW-U1gZoAu-gvjnyY2NoPofTDcP1SIWsk2tiy6EyKD4mq7p0QLpiwIVbXuB00YcLBZhNsG8so9GFyVz148x1SDaQSp67CtiPKA5umrOlAjJIhRtINGVxS0spAThASO2i1--w9ZNZ5qfi0f5LnyYv9mm30vhr-c3kH6xYZa9hnsWM-Uzafv2NhtBx49U4_JErDdMUh4oJ4qVr-L7jfgEGjLSiA13J_dptCsnyOLernMkIbu8bIkDUfQl8HZShGUFk4Nrdy4ThYhPPdYqtjZHzi6WlOaMrMe1j6VoUidT8X7gAtTKrPObDl7tLz8=
instancesSshKeyPub:
  encoding: Base64
  encryption: AES256
  value: Qpoo1n-chZfL_Xljop5CpHbx9DC7s8bqLBX6EcPSWANj4TGsRPvKPRQdZoA9wIXe6bqdB_3BtDa5o5-EBMIJxPGGNV31ZOGdc0tygSK0HW5qMJxTsYALWpbvpmO1qGi0x8Gw1IuIT01bNh57onnhT6GsBUH1qrIHn_qL0r3uYxj8Wco5XZjonpvvjF2WoDluMZXxYkjuQTPo43FHR1mm2sZrF_-LTMM2ZGOoPnJ_BhJ09PvW-NXuVLQVuiAWXws22eMBEAs3tpsCD-xK8LuV5FrWWeGelAqIxKK9ii_q3nn0Aa4j2XoTvKpS6ZFo_H9ksfHkGjMaUGOufDRJcVu35rhnc-cSoQFw_CrldJcWtMwWO0FlTXCblnMq7Ot2lujOs2RsoeVo0tnUa4EF74kZpnVBUW4EKET4BmnLPynCj7pMamYFLCfzqFk9f0PlpE9tRJ6WfrJ9X2x1XBIEhbLze1fSmX8RWVIpSyFwI0Esr39yxcU29WxMCOSMcSXcrVLnVbHgLVEARA1dSNgaQP_wwhocLBAMsOFvuda5yJF2E4Ut9j1fmzxJC3z9tmO9zLNtje5lfWsP9XsC93mpiBZenOZUCfS2I6I7lbK0iPGTSeRjbFSoYUDDUpbKZLtdLNFRqF5RgbUsWXg4bK92vh_ihfsjBFk78iCSGPqpfBjxVJBJrpIy3_Nbr3I9dBRHv2Gw4ckaI_9tG6O5Qpsk2pYP6LQREKfY0j1V7HpLIar30n0mCmUNgYsnZJdcwRR2Bs941tF6VSgF8iPlLEfXlclJqvmXqbNiFuTAAF-ej_VGM3YFdgUlkRcqLJtmjwrdl-K7wVJdvbqYinpg-uEtg0T4D7ZC_VvVd2gLu5Fzax3uuIx5auYXrLz7VHsoUTvEtnvFC0OMp6mgIMrMVjAA0Shy8BLJ_1g-Xro0T2qsegim-CpCGuGqr-liLLd8-uoavGbI46L5DluVHEBB7JqWqsexnxSum1kVcyBR0hIjhhbKGBnU_IezHEBIWFJyS9arklBClaVVWs2ANFI5oYz7gFwzfHzH9JmwVfQ=
kubeadmJoinToken:
  encoding: Base64
  encryption: AES256
  value: Fb9xN4TbDLWPn4h38hVyvXZQU9jQguw34EaD-6m7nefF6B5jzIe1`),
	})
	if marshalled.Error.Error() != testError {
		t.Fatalf("expected error to be %s, but is %s", testError, marshalled.Error.Error())
	}

}
*/
