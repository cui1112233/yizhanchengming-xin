/* REA ev_684a0100194ec31e6b8494e8e69aaaa339cea0a178c463f433a1c4aa3d4aedbd; original SHA b5b22086...; constant comments are analyst annotations. */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

longlong FUN_1800b9330(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  int *piVar1;
  char cVar2;
  int iVar3;
  uint uVar4;
  longlong *plVar5;
  longlong *plVar6;
  longlong lVar7;
  longlong *plVar8;
  undefined8 uVar9;
  longlong lVar10;
  longlong lVar11;
  undefined8 uVar12;
  code *pcVar13;
  longlong *plVar14;
  longlong *plVar15;
  longlong *plVar16;
  longlong **pplVar17;
  undefined4 uVar18;
  undefined4 uVar19;
  code *pcVar20;
  longlong *plVar21;
  bool bVar22;
  undefined1 auStack_4c8 [32];
  longlong *local_4a8;
  longlong *local_4a0;
  longlong *local_498;
  longlong *local_490;
  longlong *local_488;
  longlong **local_480;
  longlong *local_478;
  longlong *local_470;
  longlong *local_468;
  longlong *local_460;
  longlong *local_458;
  longlong *local_450;
  longlong *local_448;
  longlong *local_440;
  longlong *local_438;
  longlong *local_430;
  longlong *local_428;
  longlong *local_420;
  longlong *local_418;
  longlong *local_410;
  longlong *local_408;
  longlong *local_400;
  code *local_3f8;
  longlong *local_3f0;
  longlong *local_3e8;
  longlong *local_3e0;
  longlong *local_3d8;
  longlong *local_3d0;
  longlong *local_3c8;
  longlong *local_3c0;
  longlong *local_3b8;
  code *local_3b0;
  longlong local_3a8;
  longlong *local_3a0;
  longlong *local_398;
  longlong local_390;
  longlong *local_388;
  longlong *local_380;
  longlong *local_378;
  longlong *local_370;
  undefined4 local_368;
  longlong *local_358;
  longlong *local_348;
  longlong *local_338;
  longlong *local_330;
  longlong *local_328;
  longlong *local_320;
  longlong *local_318;
  longlong *local_310;
  code *local_308;
  longlong *local_300;
  longlong *local_2f8;
  longlong *local_2f0;
  longlong *local_2e8;
  code *local_2e0;
  longlong *local_2d8;
  longlong *local_2d0;
  longlong *local_2c8;
  longlong *local_2c0;
  longlong *local_2b8;
  longlong *local_2b0;
  longlong *local_2a8;
  code *local_2a0;
  longlong *local_298;
  longlong *local_290;
  longlong *local_288;
  longlong *local_280;
  longlong *local_278;
  longlong local_270;
  longlong *local_268;
  longlong *local_260;
  longlong *local_258;
  longlong *local_250;
  longlong *local_248;
  longlong *local_240;
  longlong *local_238;
  longlong *local_230;
  longlong *local_228;
  longlong *local_220;
  longlong *local_218;
  longlong *local_210;
  undefined8 local_208;
  longlong *local_200;
  longlong *local_1f8;
  longlong *local_1f0;
  undefined8 local_1e8;
  longlong *local_1e0;
  longlong *local_1d8;
  longlong *local_1d0;
  longlong *local_1c8;
  longlong *local_1c0;
  longlong *local_1b8;
  longlong *local_1b0;
  code *local_1a8;
  longlong *local_1a0;
  longlong *local_198;
  longlong *local_190;
  longlong *local_188;
  longlong *local_180;
  longlong *local_178;
  longlong *local_170;
  longlong *local_168;
  longlong *local_160;
  longlong *local_158;
  longlong *local_150;
  longlong *local_148;
  longlong *local_140;
  longlong *local_138;
  longlong *local_130;
  longlong *local_128;
  longlong *local_120;
  longlong *local_118;
  longlong *local_110;
  longlong *local_108;
  code *local_100;
  longlong *local_f8;
  longlong *local_f0;
  longlong *local_e8;
  longlong *local_e0;
  longlong *local_d8;
  longlong *local_d0;
  code *local_c8;
  longlong *local_c0;
  longlong *local_b8;
  longlong *local_b0;
  longlong *local_a8;
  longlong *local_98;
  longlong *local_90;
  code *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  code *local_60;
  longlong *local_58;
  ulonglong local_48;
  
  local_48 = DAT_1803a6040 ^ (ulonglong)auStack_4c8;
  local_290 = (longlong *)*param_3;
  local_220 = (longlong *)param_3[1];
  local_390 = param_1;
  local_260 = (longlong *)FUN_18036d7e0(param_3[2]);
  local_258 = (longlong *)FUN_18036d7e0(param_3[3]);
  local_2e8 = (longlong *)param_3[4];
  local_250 = (longlong *)param_3[5];
  local_2d8 = (longlong *)param_3[6];
  plVar5 = (longlong *)FUN_18036d7e0(param_3[7]);
  local_288 = (longlong *)param_3[8];
  local_218 = plVar5;
  local_240 = (longlong *)FUN_18036d7e0(param_3[9]);
  local_238 = (longlong *)FUN_18036d7e0(param_3[10]);
  local_230 = (longlong *)FUN_18036d7e0(param_3[0xb]);
  local_280 = (longlong *)param_3[0xc];
  local_278 = (longlong *)param_3[0xd];
  local_318 = (longlong *)0x0;
  plVar16 = (longlong *)0x0;
  local_380 = (longlong *)0x0;
  plVar6 = (longlong *)FUN_18036d6d0();
  plVar14 = (longlong *)0x0;
  local_2a8 = (longlong *)0x0;
  local_298 = (longlong *)0x0;
  plVar21 = (longlong *)0x0;
  local_268 = (longlong *)0x0;
  local_398 = (longlong *)0x0;
  local_310 = (longlong *)0x0;
  local_300 = (longlong *)0x0;
  local_2a0 = (code *)0x0;
  local_2d0 = (longlong *)0x0;
  local_2c8 = (longlong *)0x0;
  local_2c0 = (longlong *)0x0;
  local_2b8 = (longlong *)0x0;
  local_2f8 = (longlong *)0x0;
  local_2b0 = (longlong *)0x0;
  local_330 = (longlong *)0x0;
  local_2f0 = (longlong *)0x0;
  local_2e0 = (code *)0x0;
  local_210 = (longlong *)0x0;
  local_338 = (longlong *)0x0;
  local_3a8 = 0;
  local_328 = plVar6;
  local_498 = (longlong *)FUN_18036fb80(param_1);
  iVar3 = (int)*plVar5 + 1;
  if (iVar3 != 0) {
    *(int *)plVar5 = iVar3;
  }
  local_480 = &local_200;
  local_478 = (longlong *)0x1;
  local_488 = (longlong *)0x0;
  local_490 = (longlong *)DAT_1803ca7c0;
  local_4a0 = (longlong *)0x0;
  local_4a8 = (longlong *)0x0;
  local_200 = plVar5;
/* novels.pipeline.chapter_review_cycle#41="log" ; novels.pipeline.chapter_review_cycle#42="run_chapter_review_cycle.<locals>.log" */
  lVar7 = FUN_1803609a0(FUN_1800bed60,DAT_1803b20a8,DAT_1803b20b0,DAT_1803b1ed8);
  plVar5 = DAT_1803b1f08;
  plVar6[2] = lVar7;
  local_388 = plVar5;
  if (plVar5 == (longlong *)0x0) {
LAB_1800b955f:
    plVar5 = (longlong *)FUN_18036cd10(param_1,DAT_1803b1e98,DAT_1803ca7c0,0x110);
    DAT_1803b1f08 = plVar5;
  }
  else if ((1 < *plVar5) || (plVar5[2] != 0)) {
    if ((plVar5 != (longlong *)0x0) &&
       ((-1 < (int)*plVar5 && (*plVar5 = *plVar5 + -1, *plVar5 == 0)))) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
    goto LAB_1800b955f;
  }
  local_388 = plVar5;
  FUN_180001d00(param_1,plVar5);
  plVar6 = local_290;
  iVar3 = FUN_180001510(local_290);
  if (iVar3 == -1) {
    lVar7 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar16 = (longlong *)0x0;
    plVar21 = plVar14;
    plVar8 = plVar14;
    uVar18 = 0x92;
  }
  else {
/* novels.pipeline.chapter_review_cycle#18="" */
    plVar8 = DAT_1803b1ff0;
    if (iVar3 == 1) {
      plVar8 = plVar6;
    }
/* novels.pipeline.chapter_review_cycle#13="strip" */
    plVar6 = (longlong *)FUN_18037f0c0(param_1,plVar8,DAT_1803b1fc8);
    uVar18 = 0x92;
    if (plVar6 == (longlong *)0x0) {
      lVar7 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      plVar8 = (longlong *)0x0;
    }
    else {
      *(undefined4 *)(plVar5 + 5) = 0x92;
      plVar8 = (longlong *)FUN_180376210(param_1,plVar6);
      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
        (**(code **)(plVar6[1] + 0x30))(plVar6);
      }
      if (plVar8 == (longlong *)0x0) {
        lVar7 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar16 = (longlong *)0x0;
        plVar5 = local_388;
        plVar21 = plVar14;
        plVar8 = plVar14;
      }
      else {
        local_3a0 = plVar8;
        iVar3 = FUN_180001510(plVar8);
        plVar5 = local_388;
/* novels.pipeline.chapter_review_cycle#43="章节内容为空，无法审稿" */
        uVar9 = DAT_1803b20b8;
        pcVar20 = PyExc_ValueError_exref;
        if (iVar3 == -1) {
          lVar7 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          uVar18 = 0x93;
        }
        else if (iVar3 == 0) {
          *(undefined4 *)(local_388 + 5) = 0x94;
          local_3a8 = FUN_180376770(param_1,*(undefined8 *)pcVar20,uVar9);
          FUN_18000d1d0(param_1,local_3a8);
          uVar18 = 0x94;
          lVar7 = local_3a8;
        }
        else {
/* novels.pipeline.chapter_review_cycle#44="max" */
          uVar9 = FUN_18000d570(DAT_1803b20c0);
          plVar16 = DAT_1803ab250;
          plVar6 = (longlong *)PyNumber_Long(local_250);
          plVar5 = local_388;
          param_1 = local_390;
          uVar18 = 0x96;
          if (plVar6 == (longlong *)0x0) {
            plVar16 = (longlong *)0x0;
            lVar7 = *(longlong *)(local_390 + 0x60);
            *(undefined8 *)(local_390 + 0x60) = 0;
          }
          else {
            *(undefined4 *)(local_388 + 5) = 0x96;
            local_1f8 = plVar16;
            local_1f0 = plVar6;
            plVar16 = (longlong *)FUN_180377750(local_390,uVar9,&local_1f8);
            local_370 = plVar16;
            if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
              (**(code **)(plVar6[1] + 0x30))(plVar6);
            }
            if (plVar16 == (longlong *)0x0) {
              FUN_18000ce30(param_1,&local_3a8);
              plVar16 = (longlong *)0x0;
              plVar8 = local_3a0;
              lVar7 = local_3a8;
            }
            else {
              local_2a8 = plVar16;
              local_498 = (longlong *)FUN_18036fb80(param_1,DAT_1803b20c8);
              plVar16 = local_328;
              local_140 = local_238;
              iVar3 = (int)*local_238 + 1;
              if (iVar3 != 0) {
                *(int *)local_238 = iVar3;
              }
              local_138 = local_258;
              iVar3 = (int)*local_258 + 1;
              if (iVar3 != 0) {
                *(int *)local_258 = iVar3;
              }
              local_130 = local_328;
              iVar3 = (int)*local_328 + 1;
              if (iVar3 != 0) {
                *(int *)local_328 = iVar3;
              }
              local_128 = local_240;
              iVar3 = (int)*local_240 + 1;
              if (iVar3 != 0) {
                *(int *)local_240 = iVar3;
              }
              local_120 = local_230;
              iVar3 = (int)*local_230 + 1;
              if (iVar3 != 0) {
                *(int *)local_230 = iVar3;
              }
              local_118 = local_260;
              iVar3 = (int)*local_260 + 1;
              if (iVar3 != 0) {
                *(int *)local_260 = iVar3;
              }
              local_480 = &local_140;
              local_478 = (longlong *)0x6;
              local_488 = (longlong *)0x0;
              local_490 = (longlong *)DAT_1803ca7c0;
              local_4a0 = (longlong *)0x0;
              local_4a8 = (longlong *)0x0;
              plVar6 = (longlong *)
/* novels.pipeline.chapter_review_cycle#46="assess" ; novels.pipeline.chapter_review_cycle#47="run_chapter_review_cycle.<locals>.assess" */
                       FUN_1803609a0(FUN_1800bf180,DAT_1803b20d0,DAT_1803b20d8,DAT_1803b1ee8);
              lVar7 = plVar16[2];
              *(undefined4 *)(plVar5 + 5) = 0xb1;
              local_298 = plVar6;
              plVar16 = (longlong *)FUN_1803770d0(param_1,lVar7,DAT_1803b20e0);
              if (plVar16 == (longlong *)0x0) {
                FUN_18000ce30(param_1,&local_3a8);
                plVar16 = (longlong *)0x0;
                plVar8 = local_3a0;
                uVar18 = 0xb1;
                lVar7 = local_3a8;
              }
              else {
                if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                  (**(code **)(plVar16[1] + 0x30))();
                }
                *(undefined4 *)(plVar5 + 5) = 0xb2;
                plVar21 = (longlong *)FUN_180376770(param_1,plVar6,plVar8);
                local_228 = plVar21;
                if (plVar21 == (longlong *)0x0) {
                  FUN_18000ce30(param_1,&local_3a8);
                  plVar16 = local_398;
                  plVar21 = local_398;
                  plVar8 = local_3a0;
                  uVar18 = 0xb2;
                  lVar7 = local_3a8;
                }
                else {
                  local_268 = plVar21;
                  local_318 = (longlong *)FUN_180372160(param_1,1);
                  iVar3 = (int)*plVar21 + 1;
                  if (iVar3 != 0) {
                    *(int *)plVar21 = iVar3;
                  }
                  FUN_1800b46f0(local_318,0,plVar21);
                  iVar3 = (int)*plVar21 + 1;
                  if (iVar3 != 0) {
                    *(int *)plVar21 = iVar3;
                  }
                  local_398 = plVar21;
/* novels.pipeline.chapter_review_cycle#49="audit" */
                  plVar16 = (longlong *)FUN_18037f0c0(param_1,plVar21,DAT_1803b20e8);
                  if (plVar16 == (longlong *)0x0) {
LAB_1800b994b:
                    FUN_18000ce30(param_1,&local_3a8);
                  }
                  else {
/* novels.pipeline.chapter_review_cycle#7="get" */
                    plVar6 = (longlong *)FUN_18037f0c0(param_1,plVar16,DAT_1803b1f98);
                    if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                      lVar7 = FUN_18000cc40(plVar16);
                      (**(code **)(lVar7 + 0x30))(plVar16);
                    }
                    if (plVar6 == (longlong *)0x0) goto LAB_1800b994b;
                    *(undefined4 *)(plVar5 + 5) = 0xb6;
                    plVar16 = (longlong *)FUN_1803770d0(param_1,plVar6,DAT_1803b20f0);
                    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                      lVar7 = FUN_18000cc40(plVar6);
                      (**(code **)(lVar7 + 0x30))(plVar6);
                    }
                    uVar18 = 0xb6;
                    if (plVar16 == (longlong *)0x0) {
                      FUN_18000ce30(param_1,&local_3a8);
                      plVar16 = local_380;
                      plVar8 = local_3a0;
                      lVar7 = local_3a8;
                      goto LAB_1800bcc4b;
                    }
                    iVar3 = FUN_180001510(plVar16);
                    if (iVar3 != -1) {
                      if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                        lVar7 = FUN_18000cc40(plVar16);
                        (**(code **)(lVar7 + 0x30))(plVar16);
                      }
                      if (iVar3 == 0) {
                        lVar7 = FUN_1800b4fc0(param_1);
                        param_1 = local_390;
                        if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#57="_is_passed" */
                          FUN_180373340(local_390,&local_3a8,DAT_1803b2128);
                          plVar16 = local_380;
                          plVar8 = local_3a0;
                          uVar18 = 0xc2;
                          lVar7 = local_3a8;
                        }
                        else {
/* novels.pipeline.chapter_review_cycle#49="audit" */
                          plVar16 = (longlong *)FUN_18037f0c0(local_390,plVar21,DAT_1803b20e8);
                          if (plVar16 == (longlong *)0x0) {
                            FUN_18000ce30(param_1,&local_3a8);
                            plVar16 = local_380;
                            plVar5 = local_388;
                            plVar8 = local_3a0;
                            uVar18 = 0xc2;
                            lVar7 = local_3a8;
                          }
                          else {
/* novels.pipeline.chapter_review_cycle#54="score" */
                            plVar5 = (longlong *)FUN_18037f0c0(param_1,plVar21,DAT_1803b2110);
                            if (plVar5 == (longlong *)0x0) {
                              FUN_18000ce30(param_1,&local_3a8);
                              if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                              {
                                lVar7 = FUN_18000cc40(plVar16);
                                (**(code **)(lVar7 + 0x30))(plVar16);
                              }
                              plVar16 = local_380;
                              plVar5 = local_388;
                              plVar8 = local_3a0;
                              uVar18 = 0xc2;
                              lVar7 = local_3a8;
                            }
                            else {
/* novels.pipeline.chapter_review_cycle#55="length_in_range" */
                              plVar21 = (longlong *)FUN_18037f0c0(param_1,plVar21,DAT_1803b2118);
                              if (plVar21 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_3a8);
                                if ((-1 < (int)*plVar16) &&
                                   (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                  lVar7 = FUN_18000cc40(plVar16);
                                  (**(code **)(lVar7 + 0x30))(plVar16);
                                }
                                if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                  lVar7 = FUN_18000cc40(plVar5);
                                  (**(code **)(lVar7 + 0x30))(plVar5);
                                }
                                plVar16 = local_380;
                                plVar5 = local_388;
                                plVar21 = local_398;
                                plVar8 = local_3a0;
                                uVar18 = 0xc2;
                                lVar7 = local_3a8;
                              }
                              else {
                                uVar18 = 0xc2;
                                *(undefined4 *)(local_388 + 5) = 0xc2;
                                local_188 = local_2d8;
                                local_1a0 = plVar16;
                                local_198 = plVar5;
                                local_190 = plVar21;
                                plVar6 = (longlong *)FUN_180379a80(param_1,lVar7,&local_1a0);
                                if ((-1 < (int)*plVar16) &&
                                   (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                  lVar7 = FUN_18000cc40(plVar16);
                                  (**(code **)(lVar7 + 0x30))(plVar16);
                                }
                                if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                  lVar7 = FUN_18000cc40(plVar5);
                                  (**(code **)(lVar7 + 0x30))(plVar5);
                                }
                                if ((-1 < (int)*plVar21) &&
                                   (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                  lVar7 = FUN_18000cc40(plVar21);
                                  (**(code **)(lVar7 + 0x30))(plVar21);
                                }
                                if (plVar6 == (longlong *)0x0) {
                                  FUN_18000ce30(param_1,&local_3a8);
                                  plVar16 = local_380;
                                  plVar5 = local_388;
                                  plVar21 = local_398;
                                  plVar8 = local_3a0;
                                  lVar7 = local_3a8;
                                }
                                else {
                                  iVar3 = FUN_180001510(plVar6);
                                  if (iVar3 == -1) {
                                    FUN_18000ce30(param_1,&local_3a8);
                                    plVar16 = local_380;
                                    plVar5 = local_388;
                                    plVar21 = local_398;
                                    plVar8 = local_3a0;
                                    lVar7 = local_3a8;
                                    if ((-1 < (int)*plVar6) &&
                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                      lVar7 = FUN_18000cc40(plVar6);
                                      (**(code **)(lVar7 + 0x30))(plVar6);
                                      plVar16 = local_380;
                                      plVar5 = local_388;
                                      plVar21 = local_398;
                                      plVar8 = local_3a0;
                                      lVar7 = local_3a8;
                                    }
                                  }
                                  else {
                                    if ((-1 < (int)*plVar6) &&
                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                      lVar7 = FUN_18000cc40(plVar6);
                                      (**(code **)(lVar7 + 0x30))(plVar6);
                                    }
/* novels.pipeline.chapter_review_cycle#58="审稿通过（" */
                                    piVar1 = DAT_1803b2130;
                                    if (iVar3 == 0) {
                                      local_310 = DAT_1803ab250;
                                      iVar3 = (int)*DAT_1803ab250 + 1;
                                      if (iVar3 != 0) {
                                        *(int *)DAT_1803ab250 = iVar3;
                                      }
                                      plVar16 = (longlong *)FUN_180370ec0(param_1);
                                      local_380 = plVar16;
                                      plVar21 = (longlong *)FUN_180375130(param_1,local_2a8);
                                      if (plVar21 == (longlong *)0x0) {
                                        FUN_18000ce30(param_1,&local_3a8);
                                        plVar5 = local_388;
                                        plVar21 = local_398;
                                        plVar8 = local_3a0;
                                        uVar18 = 0xd0;
                                        lVar7 = local_3a8;
                                      }
                                      else {
                                        local_348 = (longlong *)FUN_18000d280(param_1,plVar21);
                                        if ((-1 < (int)*plVar21) &&
                                           (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                          lVar7 = FUN_18000cc40(plVar21);
                                          (**(code **)(lVar7 + 0x30))(plVar21);
                                        }
                                        if (local_348 == (longlong *)0x0) {
                                          FUN_18000ce30(param_1,&local_3a8);
                                          plVar5 = local_388;
                                          plVar21 = local_398;
                                          plVar8 = local_3a0;
                                          uVar18 = 0xd0;
                                          lVar7 = local_3a8;
                                        }
                                        else {
                                          plVar21 = (longlong *)FUN_180038d80();
                                          plVar5 = local_388;
                                          plVar16 = local_398;
                                          if (plVar21 != (longlong *)0x0) {
                                            local_368 = 0xd1;
                                            do {
                                              plVar8 = local_210;
                                              pcVar20 = local_2a0;
                                              plVar14 = local_2d0;
                                              plVar6 = local_300;
                                              plVar5 = local_310;
                                              uVar18 = local_368;
                                              local_248 = local_3a0;
                                              local_378 = plVar16;
                                              plVar16 = plVar21;
                                              if (((local_210 != (longlong *)0x0) &&
                                                  (-1 < (int)*local_210)) &&
                                                 (*local_210 = *local_210 + -1, *local_210 == 0)) {
                                                local_210 = plVar21;
                                                lVar7 = FUN_18000cc40();
                                                (**(code **)(lVar7 + 0x30))(plVar8);
                                                plVar16 = local_210;
                                              }
                                              local_210 = plVar16;
                                              iVar3 = (int)*plVar21 + 1;
                                              if (iVar3 != 0) {
                                                *(int *)plVar21 = iVar3;
                                              }
                                              local_300 = plVar21;
                                              if (((plVar6 != (longlong *)0x0) &&
                                                  (-1 < (int)*plVar6)) &&
                                                 (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar6);
                                                (**(code **)(lVar7 + 0x30))(plVar6);
                                              }
/* novels.pipeline.chapter_review_cycle#61="callable" */
                                              uVar9 = FUN_18000d570(DAT_1803b2148);
                                              param_1 = local_390;
                                              if (local_288 == (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#62="should_cancel" */
                                                FUN_180370fe0(local_390,&local_3a8,DAT_1803b2150);
                                                FUN_18000d1d0(param_1,local_3a8);
                                                plVar5 = local_388;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = local_368;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              *(undefined4 *)(local_388 + 5) = uVar18;
                                              plVar16 = (longlong *)
                                                        FUN_180376770(local_390,uVar9,local_288);
                                              if (plVar16 == (longlong *)0x0) {
LAB_1800bb80d:
                                                FUN_18000ce30(param_1,&local_3a8);
                                                plVar5 = local_388;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = local_368;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              iVar3 = FUN_180001510(plVar16);
                                              if (iVar3 == -1) {
LAB_1800bb7d4:
                                                FUN_18000ce30(param_1,&local_3a8);
                                                plVar5 = local_388;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = local_368;
                                                lVar7 = local_3a8;
                                                if ((-1 < (int)*plVar16) &&
                                                   (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar16);
                                                  (**(code **)(lVar7 + 0x30))(plVar16);
                                                  plVar5 = local_388;
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                  uVar18 = local_368;
                                                  lVar7 = local_3a8;
                                                }
                                                goto LAB_1800bc74b;
                                              }
                                              bVar22 = iVar3 != 0;
                                              if ((-1 < (int)*plVar16) &&
                                                 (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar16);
                                                (**(code **)(lVar7 + 0x30))(plVar16);
                                              }
                                              param_1 = local_390;
                                              if (iVar3 != 0) {
                                                *(undefined4 *)(local_388 + 5) = 0xd1;
                                                plVar16 = (longlong *)
                                                          FUN_180376210(local_390,local_288);
                                                if (plVar16 == (longlong *)0x0) goto LAB_1800bb80d;
                                                iVar3 = FUN_180001510(plVar16);
                                                if (iVar3 == -1) goto LAB_1800bb7d4;
                                                bVar22 = iVar3 != 0;
                                                if ((-1 < (int)*plVar16) &&
                                                   (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar16);
                                                  (**(code **)(lVar7 + 0x30))(plVar16);
                                                }
                                              }
                                              param_1 = local_390;
                                              plVar16 = local_398;
                                              plVar6 = local_398;
                                              if (bVar22) goto LAB_1800bc7b2;
                                              local_358 = (longlong *)
                                                          FUN_180380b80(plVar21,DAT_1803ab258);
                                              param_1 = local_390;
                                              if (local_358 == (longlong *)0x0) {
                                                FUN_18000ce30(local_390,&local_3a8);
                                                plVar5 = local_388;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd3;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              local_310 = local_358;
                                              if (((plVar5 != (longlong *)0x0) &&
                                                  (-1 < (int)*plVar5)) &&
                                                 (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar5);
                                                (**(code **)(lVar7 + 0x30))(plVar5);
                                              }
                                              plVar16 = local_280;
                                              param_1 = local_390;
                                              if (local_280 == (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#63="temperature_step" */
                                                FUN_180370fe0(local_390,&local_3a8,DAT_1803b2158);
                                                FUN_18000d1d0(param_1,local_3a8);
                                                plVar5 = local_388;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd4;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              iVar3 = FUN_1803858f0(local_280,DAT_1803ab250);
                                              pcVar13 = _Py_NoneStruct_exref;
                                              param_1 = local_390;
                                              if (iVar3 == -1) {
LAB_1800ba6bb:
                                                FUN_18000ce30(param_1,&local_3a8);
                                                uVar18 = 0xd4;
                                                plVar5 = local_388;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              if (iVar3 == 1) {
/* novels.pipeline.chapter_review_cycle#64="min" */
                                                uVar12 = FUN_18000d570(DAT_1803b2160);
                                                plVar5 = local_278;
                                                param_1 = local_390;
                                                uVar9 = DAT_1803b2168;
                                                if (local_278 != (longlong *)0x0) {
                                                  plVar21 = (longlong *)
                                                            FUN_180381a40(plVar21,plVar16);
                                                  param_1 = local_390;
                                                  if (plVar21 != (longlong *)0x0) {
                                                    plVar16 = (longlong *)
                                                              FUN_180381110(plVar5,plVar21);
                                                    if ((-1 < (int)*plVar21) &&
                                                       (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                      lVar7 = FUN_18000cc40(plVar21);
                                                      (**(code **)(lVar7 + 0x30))(plVar21);
                                                    }
                                                    param_1 = local_390;
                                                    if (plVar16 != (longlong *)0x0) {
                                                      *(undefined4 *)(local_388 + 5) = 0xd4;
                                                      local_1e8 = uVar9;
                                                      local_1e0 = plVar16;
                                                      pcVar13 = (code *)FUN_180377750(local_390,
                                                                                      uVar12,&
                                                  local_1e8);
                                                  local_308 = pcVar13;
                                                  if ((-1 < (int)*plVar16) &&
                                                     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                    lVar7 = FUN_18000cc40(plVar16);
                                                    (**(code **)(lVar7 + 0x30))(plVar16);
                                                  }
                                                  if (pcVar13 != (code *)0x0) goto LAB_1800ba6f4;
                                                  }
                                                  }
                                                  goto LAB_1800ba6bb;
                                                }
/* novels.pipeline.chapter_review_cycle#66="base_temperature" */
                                                FUN_180370fe0(local_390,&local_3a8,DAT_1803b2170);
                                                FUN_18000d1d0(param_1,local_3a8);
                                                plVar5 = local_388;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd4;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              local_308 = _Py_NoneStruct_exref;
                                              if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
                                                *(int *)_Py_NoneStruct_exref =
                                                     *(int *)_Py_NoneStruct_exref + 1;
                                              }
LAB_1800ba6f4:
                                              local_2a0 = pcVar13;
                                              if (((pcVar20 != (code *)0x0) &&
                                                  (-1 < *(int *)pcVar20)) &&
                                                 (*(longlong *)pcVar20 = *(longlong *)pcVar20 + -1,
                                                 *(longlong *)pcVar20 == 0)) {
                                                lVar7 = FUN_18000cc40(pcVar20);
                                                (**(code **)(lVar7 + 0x30))(pcVar20);
                                              }
                                              iVar3 = FUN_180001510(pcVar13);
/* novels.pipeline.chapter_review_cycle#67="，温度 " */
                                              piVar1 = DAT_1803b2178;
/* novels.pipeline.chapter_review_cycle#18="" */
                                              plVar21 = DAT_1803b1ff0;
                                              if (iVar3 == -1) {
LAB_1800ba7c7:
                                                FUN_18000ce30(param_1,&local_3a8);
                                                plVar5 = local_388;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd5;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              if (iVar3 == 0) {
/* novels.pipeline.chapter_review_cycle#18="" */
                                                iVar3 = (int)*DAT_1803b1ff0 + 1;
                                                if (iVar3 != 0) {
/* novels.pipeline.chapter_review_cycle#18="" */
                                                  *(int *)DAT_1803b1ff0 = iVar3;
                                                }
                                              }
                                              else {
                                                plVar16 = (longlong *)FUN_180374a40(param_1,2);
                                                if (*piVar1 + 1 != 0) {
                                                  *piVar1 = *piVar1 + 1;
                                                }
                                                FUN_18000cc80(plVar16,0,piVar1);
/* novels.pipeline.chapter_review_cycle#68=".2f" */
                                                lVar7 = FUN_180375520(param_1,pcVar13,DAT_1803b2180)
                                                ;
                                                if (lVar7 == 0) {
                                                  FUN_18000ce30(param_1,&local_3a8);
                                                  uVar18 = 0xd5;
                                                  plVar5 = local_388;
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                  lVar7 = local_3a8;
                                                  if ((-1 < (int)*plVar16) &&
                                                     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                    lVar7 = FUN_18000cc40(plVar16);
                                                    (**(code **)(lVar7 + 0x30))(plVar16);
                                                    plVar5 = local_388;
                                                    plVar21 = local_398;
                                                    plVar8 = local_3a0;
                                                    lVar7 = local_3a8;
                                                  }
                                                  goto LAB_1800bc74b;
                                                }
                                                FUN_18000cc80(plVar16,1,lVar7);
                                                plVar21 = (longlong *)
/* novels.pipeline.chapter_review_cycle#18="" */
                                                          PyUnicode_Join(DAT_1803b1ff0,plVar16);
                                                if ((-1 < (int)*plVar16) &&
                                                   (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar16);
                                                  (**(code **)(lVar7 + 0x30))(plVar16);
                                                }
                                                if (plVar21 == (longlong *)0x0) goto LAB_1800ba7c7;
                                              }
                                              local_2d0 = plVar21;
                                              if (((plVar14 != (longlong *)0x0) &&
                                                  (-1 < (int)*plVar14)) &&
                                                 (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar14);
                                                (**(code **)(lVar7 + 0x30))(plVar14);
                                              }
/* novels.pipeline.chapter_review_cycle#69="修复轮次 " */
                                              piVar1 = DAT_1803b2188;
                                              lVar7 = local_328[2];
                                              if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#41="log" */
                                                FUN_180370fe0(param_1,&local_3a8,DAT_1803b20a8);
                                                FUN_18000d1d0(param_1,local_3a8);
                                                plVar5 = local_388;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd6;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              plVar16 = (longlong *)FUN_180374a40(param_1,9);
                                              if (*piVar1 + 1 != 0) {
                                                *piVar1 = *piVar1 + 1;
                                              }
                                              FUN_18000cc80(plVar16,0,piVar1);
                                              param_1 = local_390;
                                              lVar10 = FUN_180375520(local_390,local_358,
/* novels.pipeline.chapter_review_cycle#18="" */
                                                                     DAT_1803b1ff0);
                                              if (lVar10 == 0) {
LAB_1800bc631:
                                                FUN_18000ce30(param_1,&local_3a8);
LAB_1800bc63a:
                                                uVar18 = 0xd6;
                                                plVar5 = local_388;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                lVar7 = local_3a8;
                                                if ((-1 < (int)*plVar16) &&
                                                   (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar16);
                                                  (**(code **)(lVar7 + 0x30))(plVar16);
                                                  plVar5 = local_388;
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                  lVar7 = local_3a8;
                                                }
                                                goto LAB_1800bc74b;
                                              }
                                              FUN_18000cc80(plVar16,1,lVar10);
/* novels.pipeline.chapter_review_cycle#70="/" */
                                              if (*DAT_1803b2190 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#70="/" */
                                                *DAT_1803b2190 = *DAT_1803b2190 + 1;
                                              }
                                              FUN_18000cc80(plVar16,2);
/* novels.pipeline.chapter_review_cycle#18="" */
                                              lVar10 = FUN_180375520(param_1,local_2a8,DAT_1803b1ff0
                                                                    );
                                              if (lVar10 == 0) goto LAB_1800bc631;
                                              FUN_18000cc80(plVar16,3,lVar10);
/* novels.pipeline.chapter_review_cycle#71="（当前 " */
                                              if (*DAT_1803b2198 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#71="（当前 " */
                                                *DAT_1803b2198 = *DAT_1803b2198 + 1;
                                              }
                                              FUN_18000cc80(plVar16,4);
                                              plVar6 = local_398;
                                              if (local_398 == (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#72="current" */
                                                FUN_180370fe0(param_1,&local_3a8,DAT_1803b21a0);
                                                FUN_18000d1d0(param_1,local_3a8);
                                                goto LAB_1800bc63a;
                                              }
                                              plVar5 = (longlong *)
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                       FUN_18037f0c0(param_1,local_398,DAT_1803b2110
                                                                    );
                                              if (plVar5 == (longlong *)0x0) goto LAB_1800bc631;
/* novels.pipeline.chapter_review_cycle#59=".0f" */
                                              lVar10 = FUN_180375520(param_1,plVar5,DAT_1803b2138);
                                              if ((-1 < (int)*plVar5) &&
                                                 (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                lVar11 = FUN_18000cc40(plVar5);
                                                (**(code **)(lVar11 + 0x30))(plVar5);
                                              }
                                              if (lVar10 == 0) goto LAB_1800bc631;
                                              FUN_18000cc80(plVar16,5,lVar10);
/* novels.pipeline.chapter_review_cycle#73=" 分" */
                                              if (*DAT_1803b21a8 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#73=" 分" */
                                                *DAT_1803b21a8 = *DAT_1803b21a8 + 1;
                                              }
                                              FUN_18000cc80(plVar16,6);
                                              iVar3 = (int)*plVar21 + 1;
                                              if (iVar3 != 0) {
                                                *(int *)plVar21 = iVar3;
                                              }
                                              FUN_18000cc80(plVar16,7,plVar21);
/* novels.pipeline.chapter_review_cycle#74="）" */
                                              if (*DAT_1803b21b0 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#74="）" */
                                                *DAT_1803b21b0 = *DAT_1803b21b0 + 1;
                                              }
                                              FUN_18000cc80(plVar16,8);
                                              plVar21 = (longlong *)
/* novels.pipeline.chapter_review_cycle#18="" */
                                                        PyUnicode_Join(DAT_1803b1ff0,plVar16);
                                              if ((-1 < (int)*plVar16) &&
                                                 (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                lVar10 = FUN_18000cc40(plVar16);
                                                (**(code **)(lVar10 + 0x30))(plVar16);
                                              }
                                              plVar5 = local_388;
                                              lVar10 = local_390;
                                              if (plVar21 == (longlong *)0x0) {
                                                FUN_18000ce30(param_1,&local_3a8);
                                                plVar5 = local_388;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd6;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              *(undefined4 *)(local_388 + 5) = 0xd6;
                                              plVar16 = (longlong *)
                                                        FUN_180376770(local_390,lVar7,plVar21);
                                              if ((-1 < (int)*plVar21) &&
                                                 (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar21);
                                                (**(code **)(lVar7 + 0x30))(plVar21);
                                              }
                                              param_1 = lVar10;
                                              if (plVar16 == (longlong *)0x0) {
                                                FUN_18000ce30(lVar10,&local_3a8);
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd6;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              if ((-1 < (int)*plVar16) &&
                                                 (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar16);
                                                (**(code **)(lVar7 + 0x30))(plVar16);
                                              }
                                              lVar7 = FUN_1800b4eb0(lVar10);
                                              if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#75="_format_issues_feedback" */
                                                FUN_180373340(lVar10,&local_3a8,DAT_1803b21b8);
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd7;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              plVar21 = (longlong *)
/* novels.pipeline.chapter_review_cycle#49="audit" */
                                                        FUN_18037f0c0(lVar10,plVar6,DAT_1803b20e8);
                                              if (plVar21 == (longlong *)0x0) {
LAB_1800bc59c:
                                                FUN_18000ce30(lVar10,&local_3a8);
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd7;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              plVar16 = (longlong *)
/* novels.pipeline.chapter_review_cycle#7="get" */
                                                        FUN_18037f0c0(lVar10,plVar21,DAT_1803b1f98);
                                              if ((-1 < (int)*plVar21) &&
                                                 (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                lVar11 = FUN_18000cc40(plVar21);
                                                (**(code **)(lVar11 + 0x30))(plVar21);
                                              }
                                              if (plVar16 == (longlong *)0x0) goto LAB_1800bc59c;
                                              *(undefined4 *)(plVar5 + 5) = 0xd7;
                                              plVar21 = (longlong *)
                                                        FUN_1803770d0(lVar10,plVar16,DAT_1803b21c0);
                                              if ((-1 < (int)*plVar16) &&
                                                 (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                lVar11 = FUN_18000cc40(plVar16);
                                                (**(code **)(lVar11 + 0x30))(plVar16);
                                              }
                                              if (plVar21 == (longlong *)0x0) {
                                                FUN_18000ce30(lVar10,&local_3a8);
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd7;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              iVar3 = FUN_180001510(plVar21);
                                              if (iVar3 == -1) {
LAB_1800bc4e3:
                                                FUN_18000ce30(lVar10,&local_3a8);
                                                if ((-1 < (int)*plVar21) &&
                                                   (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar21);
                                                  (**(code **)(lVar7 + 0x30))(plVar21);
                                                }
LAB_1800bc508:
                                                param_1 = lVar10;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd7;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              if (iVar3 != 1) {
                                                if ((-1 < (int)*plVar21) &&
                                                   (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                  lVar11 = FUN_18000cc40(plVar21);
                                                  (**(code **)(lVar11 + 0x30))(plVar21);
                                                }
                                                plVar21 = (longlong *)FUN_180372160(lVar10,0);
                                              }
                                              plVar16 = (longlong *)
/* novels.pipeline.chapter_review_cycle#49="audit" */
                                                        FUN_18037f0c0(lVar10,plVar6,DAT_1803b20e8);
                                              if (plVar16 == (longlong *)0x0) goto LAB_1800bc4e3;
                                              plVar6 = (longlong *)
/* novels.pipeline.chapter_review_cycle#7="get" */
                                                       FUN_18037f0c0(lVar10,plVar16,DAT_1803b1f98);
                                              if ((-1 < (int)*plVar16) &&
                                                 (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                lVar10 = FUN_18000cc40(plVar16);
                                                (**(code **)(lVar10 + 0x30))(plVar16);
                                              }
                                              lVar10 = local_390;
                                              if (plVar6 == (longlong *)0x0) goto LAB_1800bc4e3;
                                              *(undefined4 *)(plVar5 + 5) = 0xd7;
                                              plVar16 = (longlong *)
                                                        FUN_1803770d0(local_390,plVar6,DAT_1803b1fe0
                                                                     );
                                              if ((-1 < (int)*plVar6) &&
                                                 (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                lVar11 = FUN_18000cc40(plVar6);
                                                (**(code **)(lVar11 + 0x30))(plVar6);
                                              }
                                              if (plVar16 == (longlong *)0x0) goto LAB_1800bc4e3;
                                              iVar3 = FUN_180001510(plVar16);
                                              param_1 = lVar10;
                                              if (iVar3 == -1) {
                                                FUN_18000ce30(lVar10,&local_3a8);
                                                if ((-1 < (int)*plVar21) &&
                                                   (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar21);
                                                  (**(code **)(lVar7 + 0x30))(plVar21);
                                                }
                                                if (((int)*plVar16 < 0) ||
                                                   (*plVar16 = *plVar16 + -1, *plVar16 != 0))
                                                goto LAB_1800bc508;
                                                lVar7 = FUN_18000cc40(plVar16);
                                                (**(code **)(lVar7 + 0x30))(plVar16);
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd7;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              if (iVar3 != 1) {
                                                if ((-1 < (int)*plVar16) &&
                                                   (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                  lVar11 = FUN_18000cc40(plVar16);
                                                  (**(code **)(lVar11 + 0x30))(plVar16);
                                                }
/* novels.pipeline.chapter_review_cycle#18="" */
                                                plVar16 = DAT_1803b1ff0;
/* novels.pipeline.chapter_review_cycle#18="" */
                                                iVar3 = (int)*DAT_1803b1ff0 + 1;
                                                if (iVar3 != 0) {
/* novels.pipeline.chapter_review_cycle#18="" */
                                                  *(int *)DAT_1803b1ff0 = iVar3;
                                                }
                                              }
                                              plVar6 = (longlong *)FUN_1803749a0(plVar16);
                                              if ((-1 < (int)*plVar16) &&
                                                 (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                lVar11 = FUN_18000cc40(plVar16);
                                                (**(code **)(lVar11 + 0x30))(plVar16);
                                              }
                                              if (plVar6 == (longlong *)0x0) goto LAB_1800bc4e3;
                                              *(undefined4 *)(plVar5 + 5) = 0xd7;
                                              local_1d8 = plVar21;
                                              local_1d0 = plVar6;
                                              plVar16 = (longlong *)
                                                        FUN_180377750(lVar10,lVar7,&local_1d8);
                                              if ((-1 < (int)*plVar21) &&
                                                 (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar21);
                                                (**(code **)(lVar7 + 0x30))(plVar21);
                                              }
                                              if ((-1 < (int)*plVar6) &&
                                                 (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar6);
                                                (**(code **)(lVar7 + 0x30))(plVar6);
                                              }
                                              plVar21 = local_2c8;
                                              if (plVar16 == (longlong *)0x0) {
                                                FUN_18000ce30(lVar10,&local_3a8);
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd7;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              plVar6 = plVar16;
                                              if (((local_2c8 != (longlong *)0x0) &&
                                                  (-1 < (int)*local_2c8)) &&
                                                 (*local_2c8 = *local_2c8 + -1, *local_2c8 == 0)) {
                                                local_2c8 = plVar16;
                                                lVar7 = FUN_18000cc40(plVar21);
                                                (**(code **)(lVar7 + 0x30))(plVar21);
                                                plVar6 = local_2c8;
                                              }
                                              local_2c8 = plVar6;
                                              plVar21 = (longlong *)
/* novels.pipeline.chapter_review_cycle#49="audit" */
                                                        FUN_18037f0c0(lVar10,local_398,DAT_1803b20e8
                                                                     );
                                              if (plVar21 == (longlong *)0x0) {
LAB_1800bc4aa:
                                                FUN_18000ce30(lVar10,&local_3a8);
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd8;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              plVar6 = (longlong *)
/* novels.pipeline.chapter_review_cycle#7="get" */
                                                       FUN_18037f0c0(lVar10,plVar21,DAT_1803b1f98);
                                              if ((-1 < (int)*plVar21) &&
                                                 (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar21);
                                                (**(code **)(lVar7 + 0x30))(plVar21);
                                              }
                                              if (plVar6 == (longlong *)0x0) goto LAB_1800bc4aa;
                                              *(undefined4 *)(plVar5 + 5) = 0xd8;
                                              plVar21 = (longlong *)
                                                        FUN_1803770d0(lVar10,plVar6,DAT_1803b21c0);
                                              if ((-1 < (int)*plVar6) &&
                                                 (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar6);
                                                (**(code **)(lVar7 + 0x30))(plVar6);
                                              }
                                              if (plVar21 == (longlong *)0x0) {
                                                FUN_18000ce30(lVar10,&local_3a8);
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd8;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              iVar3 = FUN_180001510(plVar21);
                                              if (iVar3 == -1) {
                                                FUN_18000ce30(lVar10,&local_3a8);
                                                if ((-1 < (int)*plVar21) &&
                                                   (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar21);
                                                  (**(code **)(lVar7 + 0x30))(plVar21);
                                                }
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd8;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              if (iVar3 != 1) {
                                                if ((-1 < (int)*plVar21) &&
                                                   (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar21);
                                                  (**(code **)(lVar7 + 0x30))(plVar21);
                                                }
                                                plVar21 = (longlong *)FUN_180372160(lVar10,0);
                                              }
                                              plVar6 = (longlong *)FUN_18000d280(lVar10,plVar21);
                                              if ((-1 < (int)*plVar21) &&
                                                 (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar21);
                                                (**(code **)(lVar7 + 0x30))(plVar21);
                                              }
                                              if (plVar6 == (longlong *)0x0) {
                                                FUN_18000ce30(lVar10,&local_3a8);
                                                plVar5 = local_388;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xd8;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              plVar21 = (longlong *)FUN_18000d430(plVar6);
                                              plVar5 = (longlong *)0x0;
                                              param_1 = local_390;
                                              plVar14 = local_2c0;
                                              plVar8 = local_2b8;
                                              while (local_390 = param_1, local_2c0 = plVar14,
                                                    local_2b8 = plVar8, plVar21 != (longlong *)0x0)
                                              {
                                                if (((plVar5 != (longlong *)0x0) &&
                                                    (-1 < (int)*plVar5)) &&
                                                   (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar5);
                                                  (**(code **)(lVar7 + 0x30))(plVar5);
                                                }
                                                iVar3 = (int)*plVar21 + 1;
                                                if (iVar3 != 0) {
                                                  *(int *)plVar21 = iVar3;
                                                }
                                                local_2c0 = plVar21;
                                                if (((plVar14 != (longlong *)0x0) &&
                                                    (-1 < (int)*plVar14)) &&
                                                   (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                  lVar7 = FUN_18000cc40();
                                                  (**(code **)(lVar7 + 0x30))(plVar14);
                                                }
                                                param_1 = local_390;
                                                plVar5 = (longlong *)
                                                         FUN_18037f0c0(local_390,plVar21,
/* novels.pipeline.chapter_review_cycle#7="get" */
                                                                       DAT_1803b1f98);
                                                if (plVar5 == (longlong *)0x0) {
                                                  FUN_18000ce30(param_1,&local_3a8);
                                                  uVar18 = 0xd9;
                                                  goto LAB_1800bc3da;
                                                }
                                                *(undefined4 *)(local_388 + 5) = 0xd9;
                                                plVar14 = (longlong *)
                                                          FUN_1803770d0(param_1,plVar5,DAT_1803b21c8
                                                                       );
                                                if ((-1 < (int)*plVar5) &&
                                                   (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar5);
                                                  (**(code **)(lVar7 + 0x30))(plVar5);
                                                }
                                                if (plVar14 == (longlong *)0x0) {
                                                  FUN_18000ce30(param_1,&local_3a8);
                                                  uVar18 = 0xd9;
                                                  goto LAB_1800bc3da;
                                                }
                                                iVar3 = FUN_180001510(plVar14);
                                                if (iVar3 == -1) {
                                                  FUN_18000ce30(param_1,&local_3a8);
                                                  if ((-1 < (int)*plVar14) &&
                                                     (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                    lVar7 = FUN_18000cc40(plVar14);
                                                    (**(code **)(lVar7 + 0x30))(plVar14);
                                                  }
                                                  uVar18 = 0xd9;
                                                  goto LAB_1800bc3da;
                                                }
                                                if (iVar3 != 1) {
                                                  if ((-1 < (int)*plVar14) &&
                                                     (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                    lVar7 = FUN_18000cc40(plVar14);
                                                    (**(code **)(lVar7 + 0x30))(plVar14);
                                                  }
/* novels.pipeline.chapter_review_cycle#78="unknown" */
                                                  plVar14 = DAT_1803b21d0;
/* novels.pipeline.chapter_review_cycle#78="unknown" */
                                                  iVar3 = (int)*DAT_1803b21d0 + 1;
                                                  if (iVar3 != 0) {
/* novels.pipeline.chapter_review_cycle#78="unknown" */
                                                    *(int *)DAT_1803b21d0 = iVar3;
                                                  }
                                                }
                                                plVar5 = (longlong *)FUN_1803749a0(plVar14);
                                                if ((-1 < (int)*plVar14) &&
                                                   (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar14);
                                                  (**(code **)(lVar7 + 0x30))(plVar14);
                                                }
                                                param_1 = local_390;
                                                if (plVar5 == (longlong *)0x0) {
                                                  FUN_18000ce30(local_390,&local_3a8);
                                                  uVar18 = 0xd9;
                                                  goto LAB_1800bc3da;
                                                }
                                                local_2b8 = plVar5;
                                                if (((plVar8 != (longlong *)0x0) &&
                                                    (-1 < (int)*plVar8)) &&
                                                   (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar8);
                                                  (**(code **)(lVar7 + 0x30))(plVar8);
                                                }
                                                param_1 = local_390;
                                                plVar14 = DAT_1803ab250;
                                                if (local_380 == (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#79="failure_dimensions" */
                                                  FUN_180370fe0(local_390,&local_3a8,DAT_1803b21d8);
                                                  FUN_18000d1d0(param_1,local_3a8);
                                                  uVar18 = 0xda;
                                                  goto LAB_1800bc3da;
                                                }
                                                plVar8 = (longlong *)
                                                         FUN_18036f8e0(local_390,local_380,plVar5);
                                                if (plVar8 == (longlong *)0x0) {
                                                  cVar2 = FUN_18005bad0(param_1);
                                                  if (cVar2 == '\0') {
                                                    iVar3 = (int)*plVar14 + 1;
                                                    if (iVar3 != 0) {
                                                      *(int *)plVar14 = iVar3;
                                                    }
                                                    plVar8 = plVar14;
                                                    if (plVar14 != (longlong *)0x0)
                                                    goto LAB_1800bafdd;
                                                  }
LAB_1800bb8ac:
                                                  FUN_18000ce30(param_1,&local_3a8);
                                                  uVar18 = 0xda;
                                                  goto LAB_1800bc3da;
                                                }
LAB_1800bafdd:
                                                plVar14 = (longlong *)
                                                          FUN_180380b80(plVar8,DAT_1803ab258);
                                                if ((-1 < (int)*plVar8) &&
                                                   (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar8);
                                                  (**(code **)(lVar7 + 0x30))(plVar8);
                                                }
                                                if (plVar14 == (longlong *)0x0) goto LAB_1800bb8ac;
                                                iVar3 = PyDict_SetItem(local_380,plVar5,plVar14);
                                                if ((-1 < (int)*plVar14) &&
                                                   (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar14);
                                                  (**(code **)(lVar7 + 0x30))(plVar14);
                                                }
                                                if (iVar3 != 0) {
                                                  FUN_18000ce30(param_1,&local_3a8);
                                                  uVar18 = 0xda;
                                                  goto LAB_1800bc3da;
                                                }
                                                cVar2 = FUN_18000d5b0();
                                                plVar5 = plVar21;
                                                if (cVar2 == '\0') goto LAB_1800bc3cb;
                                                plVar21 = (longlong *)FUN_18000d430(plVar6);
                                                param_1 = local_390;
                                                plVar14 = local_2c0;
                                                plVar8 = local_2b8;
                                              }
                                              cVar2 = FUN_1800b4700(param_1);
                                              if (cVar2 == '\0') {
LAB_1800bc3cb:
                                                FUN_18000ce30(param_1,&local_3a8);
                                                uVar18 = 0xd8;
                                                plVar21 = plVar5;
LAB_1800bc3da:
                                                lVar7 = local_3a8;
                                                FUN_18000ce70(&local_3a8);
                                                if (((plVar21 != (longlong *)0x0) &&
                                                    (-1 < (int)*plVar21)) &&
                                                   (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                  lVar10 = FUN_18000cc40(plVar21);
                                                  (**(code **)(lVar10 + 0x30))(plVar21);
                                                }
                                                plVar5 = local_388;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                if ((-1 < (int)*plVar6) &&
                                                   (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                  lVar10 = FUN_18000cc40(plVar6);
                                                  (**(code **)(lVar10 + 0x30))(plVar6);
                                                  plVar5 = local_388;
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                }
                                                goto LAB_1800bc74b;
                                              }
                                              if (((plVar5 != (longlong *)0x0) &&
                                                  (-1 < (int)*plVar5)) &&
                                                 (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar5);
                                                (**(code **)(lVar7 + 0x30))(plVar5);
                                              }
                                              if ((-1 < (int)*plVar6) &&
                                                 (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar6);
                                                (**(code **)(lVar7 + 0x30))(plVar6);
                                              }
                                              plVar6 = local_2e8;
                                              plVar5 = local_388;
                                              plVar21 = local_3a0;
                                              if (local_2e8 == (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#80="revise_fn" */
                                                FUN_180370fe0(param_1,&local_3a8,DAT_1803b21e0);
                                                FUN_18000d1d0(param_1,local_3a8);
                                                plVar5 = local_388;
                                                plVar21 = local_3a0;
LAB_1800bb5f8:
                                                local_270 = local_3a8;
                                                FUN_18000ce70(&local_3a8);
                                                local_208 = FUN_18000ccc0(param_1);
                                                lVar7 = FUN_18000ce80(&local_270);
                                                if (lVar7 == 0) {
                                                  uVar9 = FUN_18038e2a0(plVar5,0xdc);
                                                }
                                                else {
                                                  uVar9 = FUN_18000cc90(lVar7,plVar5,0xdc);
                                                }
                                                FUN_180001300(&local_270,uVar9);
                                                FUN_18000d050(param_1,&local_270);
                                                uVar9 = FUN_18000cc40(*(undefined8 *)
                                                                       (param_1 + 0xf8));
                                                iVar3 = FUN_18000ce90(param_1,uVar9,
                                                                      *(undefined8 *)
                                                                       PyExc_TypeError_exref);
                                                if (iVar3 == 0) {
                                                  cVar2 = FUN_180373290(param_1,&local_3a8);
                                                  uVar18 = 0xdb;
                                                  if (cVar2 != '\0') {
                                                    uVar18 = 0;
                                                  }
                                                  lVar7 = FUN_18000ce80(&local_3a8);
                                                  if ((lVar7 != 0) &&
                                                     (*(longlong **)(lVar7 + 0x18) == plVar5)) {
                                                    *(undefined4 *)(plVar5 + 5) =
                                                         *(undefined4 *)(lVar7 + 0x24);
                                                  }
                                                }
                                                else {
/* novels.pipeline.chapter_review_cycle#80="revise_fn" */
                                                  uVar9 = DAT_1803b21e0;
                                                  if ((plVar6 == (longlong *)0x0) ||
/* novels.pipeline.chapter_review_cycle#81="content" */
                                                     (uVar9 = DAT_1803b21e8,
                                                     plVar21 == (longlong *)0x0)) {
                                                    FUN_180370fe0(param_1,&local_3a8,uVar9);
                                                    FUN_18000d1d0(param_1,local_3a8);
LAB_1800bc362:
                                                    uVar18 = 0xde;
                                                  }
                                                  else {
                                                    *(undefined4 *)(plVar5 + 5) = 0xde;
                                                    local_1c8 = plVar21;
                                                    local_1c0 = plVar16;
                                                    plVar16 = (longlong *)
                                                              FUN_180377750(param_1,plVar6,
                                                                            &local_1c8);
                                                    if (plVar16 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_3a8);
                                                      uVar18 = 0xde;
                                                    }
                                                    else {
                                                      iVar3 = FUN_180001510(plVar16);
                                                      if (iVar3 == -1) {
                                                        FUN_18000ce30(param_1,&local_3a8);
                                                        if (((int)*plVar16 < 0) ||
                                                           (*plVar16 = *plVar16 + -1, *plVar16 != 0)
                                                           ) goto LAB_1800bc362;
                                                        lVar7 = FUN_18000cc40(plVar16);
                                                        (**(code **)(lVar7 + 0x30))(plVar16);
                                                        uVar18 = 0xde;
                                                      }
                                                      else {
                                                        if (iVar3 != 1) {
                                                          if ((-1 < (int)*plVar16) &&
                                                             (*plVar16 = *plVar16 + -1,
                                                             *plVar16 == 0)) {
                                                            lVar7 = FUN_18000cc40(plVar16);
                                                            (**(code **)(lVar7 + 0x30))(plVar16);
                                                          }
/* novels.pipeline.chapter_review_cycle#18="" */
                                                          plVar16 = DAT_1803b1ff0;
/* novels.pipeline.chapter_review_cycle#18="" */
                                                          iVar3 = (int)*DAT_1803b1ff0 + 1;
                                                          if (iVar3 != 0) {
/* novels.pipeline.chapter_review_cycle#18="" */
                                                            *(int *)DAT_1803b1ff0 = iVar3;
                                                          }
                                                        }
                                                        plVar6 = (longlong *)
                                                                 FUN_18037f0c0(param_1,plVar16,
/* novels.pipeline.chapter_review_cycle#13="strip" */
                                                                               DAT_1803b1fc8);
                                                        if ((-1 < (int)*plVar16) &&
                                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)
                                                           ) {
                                                          lVar7 = FUN_18000cc40(plVar16);
                                                          (**(code **)(lVar7 + 0x30))(plVar16);
                                                        }
                                                        if (plVar6 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_3a8);
                                                          uVar18 = 0xde;
                                                        }
                                                        else {
                                                          *(undefined4 *)(plVar5 + 5) = 0xde;
                                                          plVar8 = (longlong *)
                                                                   FUN_180376210(param_1,plVar6);
                                                          if ((-1 < (int)*plVar6) &&
                                                             (*plVar6 = *plVar6 + -1, *plVar6 == 0))
                                                          {
                                                            lVar7 = FUN_18000cc40(plVar6);
                                                            (**(code **)(lVar7 + 0x30))(plVar6);
                                                          }
                                                          plVar16 = local_2f8;
                                                          param_1 = local_390;
                                                          if (plVar8 != (longlong *)0x0) {
                                                            plVar6 = plVar8;
                                                            if (((local_2f8 != (longlong *)0x0) &&
                                                                (-1 < (int)*local_2f8)) &&
                                                               (*local_2f8 = *local_2f8 + -1,
                                                               *local_2f8 == 0)) {
                                                              local_2f8 = plVar8;
                                                              lVar7 = FUN_18000cc40(plVar16);
                                                              (**(code **)(lVar7 + 0x30))(plVar16);
                                                              plVar6 = local_2f8;
                                                            }
                                                            local_2f8 = plVar6;
                                                            lVar7 = local_390;
                                                            FUN_18000cce0(local_390,&local_208);
                                                            plVar16 = local_2f8;
                                                            goto LAB_1800bb281;
                                                          }
                                                          FUN_18000ce30(local_390,&local_3a8);
                                                          uVar18 = 0xde;
                                                        }
                                                      }
                                                    }
                                                  }
                                                }
                                                lVar7 = local_3a8;
                                                FUN_18000ce70(&local_3a8);
                                                FUN_18000cce0(param_1,&local_208);
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                goto LAB_1800bc74b;
                                              }
                                              if (local_3a0 == (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#81="content" */
                                                FUN_180370fe0(param_1,&local_3a8,DAT_1803b21e8);
                                                FUN_18000d1d0(param_1,local_3a8);
                                                plVar5 = local_388;
                                                goto LAB_1800bb5f8;
                                              }
                                              *(undefined4 *)(local_388 + 5) = 0xdc;
                                              local_1b8 = local_3a0;
                                              local_1a8 = local_308;
                                              local_1b0 = plVar16;
                                              plVar14 = (longlong *)
                                                        FUN_180378890(param_1,local_2e8,&local_1b8);
                                              if (plVar14 == (longlong *)0x0) goto LAB_1800bb5e8;
                                              iVar3 = FUN_180001510(plVar14);
                                              if (iVar3 == -1) {
                                                FUN_18000ce30(param_1,&local_3a8);
                                                plVar6 = local_2e8;
                                                if ((-1 < (int)*plVar14) &&
                                                   (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar14);
                                                  (**(code **)(lVar7 + 0x30))(plVar14);
                                                  plVar6 = local_2e8;
                                                }
                                                goto LAB_1800bb5f8;
                                              }
                                              if (iVar3 != 1) {
                                                if ((-1 < (int)*plVar14) &&
                                                   (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar14);
                                                  (**(code **)(lVar7 + 0x30))(plVar14);
                                                }
/* novels.pipeline.chapter_review_cycle#18="" */
                                                plVar14 = DAT_1803b1ff0;
/* novels.pipeline.chapter_review_cycle#18="" */
                                                iVar3 = (int)*DAT_1803b1ff0 + 1;
                                                if (iVar3 != 0) {
/* novels.pipeline.chapter_review_cycle#18="" */
                                                  *(int *)DAT_1803b1ff0 = iVar3;
                                                }
                                              }
                                              plVar6 = (longlong *)
/* novels.pipeline.chapter_review_cycle#13="strip" */
                                                       FUN_18037f0c0(param_1,plVar14,DAT_1803b1fc8);
                                              if ((-1 < (int)*plVar14) &&
                                                 (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar14);
                                                (**(code **)(lVar7 + 0x30))(plVar14);
                                              }
                                              if (plVar6 == (longlong *)0x0) {
LAB_1800bb5e8:
                                                FUN_18000ce30(param_1,&local_3a8);
                                                plVar6 = local_2e8;
                                                goto LAB_1800bb5f8;
                                              }
                                              *(undefined4 *)(plVar5 + 5) = 0xdc;
                                              plVar8 = (longlong *)FUN_180376210(param_1,plVar6);
                                              if ((-1 < (int)*plVar6) &&
                                                 (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar6);
                                                (**(code **)(lVar7 + 0x30))(plVar6);
                                              }
                                              plVar6 = local_2f8;
                                              param_1 = local_390;
                                              if (plVar8 == (longlong *)0x0) goto LAB_1800bb5e8;
                                              lVar7 = local_390;
                                              plVar16 = plVar8;
                                              if (((local_2f8 != (longlong *)0x0) &&
                                                  (-1 < (int)*local_2f8)) &&
                                                 (*local_2f8 = *local_2f8 + -1, *local_2f8 == 0)) {
                                                local_2f8 = plVar8;
                                                lVar7 = FUN_18000cc40(plVar6);
                                                (**(code **)(lVar7 + 0x30))(plVar6);
                                                lVar7 = local_390;
                                                plVar16 = local_2f8;
                                              }
LAB_1800bb281:
                                              local_2f8 = plVar16;
                                              iVar3 = FUN_180001510(plVar8);
                                              if ((iVar3 == -1) ||
                                                 ((uVar4 = (uint)(iVar3 == 0), iVar3 != 0 &&
                                                  (uVar4 = FUN_1803832e0(plVar8,plVar21),
                                                  uVar4 == 0xffffffff)))) {
                                                FUN_18000ce30(lVar7,&local_3a8);
                                                param_1 = lVar7;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xdf;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              param_1 = local_390;
/* novels.pipeline.chapter_review_cycle#69="修复轮次 " */
                                              piVar1 = DAT_1803b2188;
                                              if (uVar4 == 1) {
                                                lVar7 = local_328[2];
                                                if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#41="log" */
                                                  FUN_180370fe0(local_390,&local_3a8,DAT_1803b20a8);
                                                  FUN_18000d1d0(param_1,local_3a8);
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                  uVar18 = 0xe0;
                                                  lVar7 = local_3a8;
                                                  goto LAB_1800bc74b;
                                                }
                                                plVar6 = (longlong *)FUN_180374a40(local_390,3);
                                                if (*piVar1 + 1 != 0) {
                                                  *piVar1 = *piVar1 + 1;
                                                }
                                                FUN_18000cc80(plVar6,0,piVar1);
                                                lVar10 = FUN_180375520(param_1,local_358,
/* novels.pipeline.chapter_review_cycle#18="" */
                                                                       DAT_1803b1ff0);
                                                if (lVar10 == 0) {
                                                  FUN_18000ce30(param_1,&local_3a8);
                                                  uVar18 = 0xe0;
LAB_1800bbb2f:
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                  lVar7 = local_3a8;
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar7 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar7 + 0x30))(plVar6);
                                                    plVar21 = local_398;
                                                    plVar8 = local_3a0;
                                                    lVar7 = local_3a8;
                                                  }
                                                  goto LAB_1800bc74b;
                                                }
                                                FUN_18000cc80(plVar6,1,lVar10);
/* novels.pipeline.chapter_review_cycle#82=" 未产出新内容，退出循环" */
                                                if (*DAT_1803b21f0 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#82=" 未产出新内容，退出循环" */
                                                  *DAT_1803b21f0 = *DAT_1803b21f0 + 1;
                                                }
                                                FUN_18000cc80(plVar6,2);
                                                plVar21 = (longlong *)
/* novels.pipeline.chapter_review_cycle#18="" */
                                                          PyUnicode_Join(DAT_1803b1ff0,plVar6);
                                                if ((-1 < (int)*plVar6) &&
                                                   (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                  lVar10 = FUN_18000cc40(plVar6);
                                                  (**(code **)(lVar10 + 0x30))(plVar6);
                                                }
                                                if (plVar21 == (longlong *)0x0) {
                                                  FUN_18000ce30(param_1,&local_3a8);
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                  uVar18 = 0xe0;
                                                  lVar7 = local_3a8;
                                                  goto LAB_1800bc74b;
                                                }
                                                *(undefined4 *)(plVar5 + 5) = 0xe0;
                                                plVar14 = (longlong *)
                                                          FUN_180376770(param_1,lVar7,plVar21);
                                                if ((-1 < (int)*plVar21) &&
                                                   (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar21);
                                                  (**(code **)(lVar7 + 0x30))(plVar21);
                                                }
                                                if (plVar14 == (longlong *)0x0) {
                                                  FUN_18000ce30(param_1,&local_3a8);
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                  uVar18 = 0xe0;
                                                  lVar7 = local_3a8;
                                                  goto LAB_1800bc74b;
                                                }
LAB_1800bbc2c:
                                                plVar16 = local_398;
                                                plVar6 = local_398;
                                                if ((-1 < (int)*plVar14) &&
                                                   (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar14);
                                                  (**(code **)(lVar7 + 0x30))(plVar14);
                                                  plVar16 = local_398;
                                                  plVar6 = local_398;
                                                }
                                                goto LAB_1800bc7b2;
                                              }
                                              iVar3 = (int)*plVar8 + 1;
                                              if (iVar3 != 0) {
                                                *(int *)plVar8 = iVar3;
                                              }
                                              local_3a0 = plVar8;
                                              if (((plVar21 != (longlong *)0x0) &&
                                                  (-1 < (int)*local_248)) &&
                                                 (*local_248 = *local_248 + -1, *local_248 == 0)) {
                                                lVar10 = FUN_18000cc40(plVar21);
                                                (**(code **)(lVar10 + 0x30))(plVar21);
                                              }
                                              if (local_298 == (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#46="assess" */
                                                FUN_180370fe0(lVar7,&local_3a8,DAT_1803b20d0);
                                                FUN_18000d1d0(lVar7,local_3a8);
                                                param_1 = lVar7;
                                                plVar21 = local_398;
                                                uVar18 = 0xe3;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              *(undefined4 *)(plVar5 + 5) = 0xe3;
                                              plVar16 = (longlong *)
                                                        FUN_180376770(lVar7,local_298,plVar8);
                                              plVar21 = local_2b0;
                                              if (plVar16 == (longlong *)0x0) {
                                                FUN_18000ce30(lVar7,&local_3a8);
                                                param_1 = lVar7;
                                                plVar21 = local_398;
                                                uVar18 = 0xe3;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              plVar6 = plVar16;
                                              if (((local_2b0 != (longlong *)0x0) &&
                                                  (-1 < (int)*local_2b0)) &&
                                                 (*local_2b0 = *local_2b0 + -1, *local_2b0 == 0)) {
                                                local_2b0 = plVar16;
                                                lVar10 = FUN_18000cc40(plVar21);
                                                (**(code **)(lVar10 + 0x30))(plVar21);
                                                plVar6 = local_2b0;
                                              }
                                              local_2b0 = plVar6;
                                              if (local_318 == (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#83="snapshots" */
                                                FUN_180370fe0(lVar7,&local_3a8,DAT_1803b21f8);
                                                FUN_18000d1d0(lVar7,local_3a8);
                                                param_1 = lVar7;
                                                plVar21 = local_398;
                                                uVar18 = 0xe4;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              FUN_1803725b0(local_318,plVar16);
                                              lVar10 = FUN_1800b4fc0(lVar7);
                                              if (lVar10 == 0) {
/* novels.pipeline.chapter_review_cycle#57="_is_passed" */
                                                FUN_180373340(lVar7,&local_3a8,DAT_1803b2128);
                                                param_1 = lVar7;
                                                plVar21 = local_398;
                                                uVar18 = 0xe5;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              plVar21 = (longlong *)
/* novels.pipeline.chapter_review_cycle#49="audit" */
                                                        FUN_18037f0c0(lVar7,plVar16,DAT_1803b20e8);
                                              if (plVar21 == (longlong *)0x0) {
                                                FUN_18000ce30(lVar7,&local_3a8);
                                                param_1 = lVar7;
                                                plVar21 = local_398;
                                                uVar18 = 0xe5;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              plVar6 = (longlong *)
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                       FUN_18037f0c0(lVar7,plVar16,DAT_1803b2110);
                                              if (plVar6 == (longlong *)0x0) {
                                                FUN_18000ce30(lVar7,&local_3a8);
                                                if (((int)*plVar21 < 0) ||
                                                   (*plVar21 = *plVar21 + -1, *plVar21 != 0)) {
LAB_1800bbf31:
                                                  param_1 = lVar7;
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                  uVar18 = 0xe5;
                                                  lVar7 = local_3a8;
                                                }
                                                else {
                                                  lVar10 = FUN_18000cc40(plVar21);
                                                  (**(code **)(lVar10 + 0x30))(plVar21);
                                                  param_1 = lVar7;
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                  uVar18 = 0xe5;
                                                  lVar7 = local_3a8;
                                                }
                                                goto LAB_1800bc74b;
                                              }
                                              plVar14 = (longlong *)
/* novels.pipeline.chapter_review_cycle#55="length_in_range" */
                                                        FUN_18037f0c0(lVar7,plVar16,DAT_1803b2118);
                                              lVar7 = local_390;
                                              if (plVar14 == (longlong *)0x0) {
                                                FUN_18000ce30(local_390,&local_3a8);
                                                if ((-1 < (int)*plVar21) &&
                                                   (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                  lVar10 = FUN_18000cc40(plVar21);
                                                  (**(code **)(lVar10 + 0x30))(plVar21);
                                                }
                                                if (((int)*plVar6 < 0) ||
                                                   (*plVar6 = *plVar6 + -1, *plVar6 != 0))
                                                goto LAB_1800bbf31;
                                                lVar10 = FUN_18000cc40(plVar6);
                                                (**(code **)(lVar10 + 0x30))(plVar6);
                                                param_1 = lVar7;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xe5;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              if (local_2d8 == (longlong *)0x0) {
                                                if ((-1 < (int)*plVar21) &&
                                                   (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar21);
                                                  (**(code **)(lVar7 + 0x30))(plVar21);
                                                }
                                                if ((-1 < (int)*plVar6) &&
                                                   (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar6);
                                                  (**(code **)(lVar7 + 0x30))(plVar6);
                                                }
                                                if ((-1 < (int)*plVar14) &&
                                                   (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar14);
                                                  (**(code **)(lVar7 + 0x30))(plVar14);
                                                }
                                                param_1 = local_390;
/* novels.pipeline.chapter_review_cycle#84="pass_score" */
                                                FUN_180370fe0(local_390,&local_3a8,DAT_1803b2200);
                                                FUN_18000d1d0(param_1,local_3a8);
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xe5;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              *(undefined4 *)(plVar5 + 5) = 0xe5;
                                              local_168 = local_2d8;
                                              local_180 = plVar21;
                                              local_178 = plVar6;
                                              local_170 = plVar14;
                                              plVar8 = (longlong *)
                                                       FUN_180379a80(local_390,lVar10,&local_180);
                                              if ((-1 < (int)*plVar21) &&
                                                 (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar21);
                                                (**(code **)(lVar7 + 0x30))(plVar21);
                                              }
                                              if ((-1 < (int)*plVar6) &&
                                                 (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar6);
                                                (**(code **)(lVar7 + 0x30))(plVar6);
                                              }
                                              if ((-1 < (int)*plVar14) &&
                                                 (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar14);
                                                (**(code **)(lVar7 + 0x30))(plVar14);
                                              }
                                              param_1 = local_390;
                                              if (plVar8 == (longlong *)0x0) {
                                                FUN_18000ce30(local_390,&local_3a8);
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xe5;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              iVar3 = FUN_180001510(plVar8);
                                              lVar7 = local_390;
                                              if (iVar3 == -1) {
                                                FUN_18000ce30(local_390,&local_3a8);
                                                if ((-1 < (int)*plVar8) &&
                                                   (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                  lVar10 = FUN_18000cc40(plVar8);
                                                  (**(code **)(lVar10 + 0x30))(plVar8);
                                                }
                                                goto LAB_1800bbf31;
                                              }
                                              if ((-1 < (int)*plVar8) &&
                                                 (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar8);
                                                (**(code **)(lVar7 + 0x30))(plVar8);
                                              }
                                              param_1 = local_390;
/* novels.pipeline.chapter_review_cycle#85="修复后达到通过线（" */
                                              piVar1 = DAT_1803b2208;
                                              if (iVar3 != 0) {
                                                lVar7 = local_328[2];
                                                if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#41="log" */
                                                  FUN_180370fe0(local_390,&local_3a8,DAT_1803b20a8);
                                                  FUN_18000d1d0(param_1,local_3a8);
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                  uVar18 = 0xe6;
                                                  lVar7 = local_3a8;
                                                  goto LAB_1800bc74b;
                                                }
                                                plVar6 = (longlong *)FUN_180374a40(local_390,3);
                                                if (*piVar1 + 1 != 0) {
                                                  *piVar1 = *piVar1 + 1;
                                                }
                                                FUN_18000cc80(plVar6,0,piVar1);
                                                plVar21 = (longlong *)
                                                          FUN_18037f0c0(param_1,plVar16,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                                        DAT_1803b2110);
                                                if (plVar21 != (longlong *)0x0) {
                                                  lVar10 = FUN_180375520(param_1,plVar21,
/* novels.pipeline.chapter_review_cycle#59=".0f" */
                                                                         DAT_1803b2138);
                                                  if ((-1 < (int)*plVar21) &&
                                                     (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                    lVar11 = FUN_18000cc40(plVar21);
                                                    (**(code **)(lVar11 + 0x30))(plVar21);
                                                  }
                                                  param_1 = local_390;
                                                  if (lVar10 != 0) {
                                                    FUN_18000cc80(plVar6,1,lVar10);
/* novels.pipeline.chapter_review_cycle#86=" 分），退出循环" */
                                                    if (*DAT_1803b2210 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#86=" 分），退出循环" */
                                                      *DAT_1803b2210 = *DAT_1803b2210 + 1;
                                                    }
                                                    FUN_18000cc80(plVar6,2);
                                                    plVar21 = (longlong *)
/* novels.pipeline.chapter_review_cycle#18="" */
                                                              PyUnicode_Join(DAT_1803b1ff0,plVar6);
                                                    if ((-1 < (int)*plVar6) &&
                                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                      lVar10 = FUN_18000cc40(plVar6);
                                                      (**(code **)(lVar10 + 0x30))(plVar6);
                                                    }
                                                    param_1 = local_390;
                                                    if (plVar21 == (longlong *)0x0) {
                                                      FUN_18000ce30(local_390,&local_3a8);
                                                      plVar21 = local_398;
                                                      plVar8 = local_3a0;
                                                      uVar18 = 0xe6;
                                                      lVar7 = local_3a8;
                                                      goto LAB_1800bc74b;
                                                    }
                                                    *(undefined4 *)(plVar5 + 5) = 0xe6;
                                                    plVar6 = (longlong *)
                                                             FUN_180376770(local_390,lVar7,plVar21);
                                                    if ((-1 < (int)*plVar21) &&
                                                       (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                      lVar7 = FUN_18000cc40(plVar21);
                                                      (**(code **)(lVar7 + 0x30))(plVar21);
                                                    }
                                                    if (plVar6 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_3a8);
                                                      plVar21 = local_398;
                                                      plVar8 = local_3a0;
                                                      uVar18 = 0xe6;
                                                      lVar7 = local_3a8;
                                                      goto LAB_1800bc74b;
                                                    }
                                                    if ((-1 < (int)*plVar6) &&
                                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                      lVar7 = FUN_18000cc40(plVar6);
                                                      (**(code **)(lVar7 + 0x30))(plVar6);
                                                    }
                                                    plVar21 = local_398;
                                                    iVar3 = (int)*plVar16 + 1;
                                                    if (iVar3 != 0) {
                                                      *(int *)plVar16 = iVar3;
                                                    }
                                                    plVar6 = plVar16;
                                                    if ((-1 < (int)*local_398) &&
                                                       (*local_398 = *local_398 + -1,
                                                       *local_398 == 0)) {
                                                      local_398 = plVar16;
                                                      lVar7 = FUN_18000cc40(plVar21);
                                                      (**(code **)(lVar7 + 0x30))(plVar21);
                                                      plVar6 = local_398;
                                                    }
                                                    goto LAB_1800bc7b2;
                                                  }
                                                }
                                                uVar18 = 0xe6;
LAB_1800bbb26:
                                                FUN_18000ce30(param_1,&local_3a8);
                                                goto LAB_1800bbb2f;
                                              }
                                              plVar6 = (longlong *)
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                       FUN_18037f0c0(local_390,plVar16,DAT_1803b2110
                                                                    );
                                              plVar21 = local_398;
                                              if (plVar6 == (longlong *)0x0) {
                                                FUN_18000ce30(param_1,&local_3a8);
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                uVar18 = 0xe9;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              plVar14 = (longlong *)
                                                        FUN_18037f0c0(param_1,local_398,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                                      DAT_1803b2110);
                                              if (plVar14 == (longlong *)0x0) {
LAB_1800bbc77:
                                                FUN_18000ce30(param_1,&local_3a8);
                                                if ((-1 < (int)*plVar6) &&
                                                   (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar6);
                                                  (**(code **)(lVar7 + 0x30))(plVar6);
                                                }
LAB_1800bbc9c:
                                                plVar8 = local_3a0;
                                                uVar18 = 0xe9;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              lVar7 = FUN_1800b4bf0(param_1);
                                              if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#87="NET_IMPROVEMENT_EPSILON" */
                                                FUN_180373340(param_1,&local_3a8,DAT_1803b2218);
                                                if ((-1 < (int)*plVar6) &&
                                                   (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                  lVar7 = FUN_18000cc40(plVar6);
                                                  (**(code **)(lVar7 + 0x30))(plVar6);
                                                }
                                                if (((int)*plVar14 < 0) ||
                                                   (*plVar14 = *plVar14 + -1, *plVar14 != 0))
                                                goto LAB_1800bbc9c;
                                                lVar7 = FUN_18000cc40(plVar14);
                                                (**(code **)(lVar7 + 0x30))(plVar14);
                                                plVar8 = local_3a0;
                                                uVar18 = 0xe9;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              plVar8 = (longlong *)FUN_180381110(plVar14,lVar7);
                                              if ((-1 < (int)*plVar14) &&
                                                 (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar14);
                                                (**(code **)(lVar7 + 0x30))(plVar14);
                                              }
                                              param_1 = local_390;
                                              if (plVar8 == (longlong *)0x0) goto LAB_1800bbc77;
                                              iVar3 = FUN_180384e70(plVar6,plVar8);
                                              if ((-1 < (int)*plVar6) &&
                                                 (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar6);
                                                (**(code **)(lVar7 + 0x30))(plVar6);
                                              }
                                              if ((-1 < (int)*plVar8) &&
                                                 (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar8);
                                                (**(code **)(lVar7 + 0x30))(plVar8);
                                              }
                                              param_1 = local_390;
/* novels.pipeline.chapter_review_cycle#69="修复轮次 " */
                                              piVar1 = DAT_1803b2188;
                                              if (iVar3 == -1) {
                                                FUN_18000ce30(local_390,&local_3a8);
                                                plVar8 = local_3a0;
                                                uVar18 = 0xe9;
                                                lVar7 = local_3a8;
                                                goto LAB_1800bc74b;
                                              }
                                              if (iVar3 != 1) {
                                                lVar7 = local_328[2];
                                                if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#41="log" */
                                                  FUN_180370fe0(local_390,&local_3a8,DAT_1803b20a8);
                                                  FUN_18000d1d0(param_1,local_3a8);
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                  uVar18 = 0xec;
                                                  lVar7 = local_3a8;
                                                  goto LAB_1800bc74b;
                                                }
                                                plVar6 = (longlong *)FUN_180374a40(local_390,7);
                                                if (*piVar1 + 1 != 0) {
                                                  *piVar1 = *piVar1 + 1;
                                                }
                                                FUN_18000cc80(plVar6,0,piVar1);
                                                lVar10 = FUN_180375520(param_1,local_358,
/* novels.pipeline.chapter_review_cycle#18="" */
                                                                       DAT_1803b1ff0);
                                                if (lVar10 != 0) {
                                                  FUN_18000cc80(plVar6,1,lVar10);
/* novels.pipeline.chapter_review_cycle#88=" 未净提升（" */
                                                  if (*DAT_1803b2220 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#88=" 未净提升（" */
                                                    *DAT_1803b2220 = *DAT_1803b2220 + 1;
                                                  }
                                                  FUN_18000cc80(plVar6,2);
                                                  plVar21 = (longlong *)
                                                            FUN_18037f0c0(param_1,local_398,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                                          DAT_1803b2110);
                                                  if (plVar21 != (longlong *)0x0) {
                                                    lVar10 = FUN_180375520(param_1,plVar21,
/* novels.pipeline.chapter_review_cycle#59=".0f" */
                                                                           DAT_1803b2138);
                                                    if ((-1 < (int)*plVar21) &&
                                                       (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar21);
                                                      (**(code **)(lVar11 + 0x30))(plVar21);
                                                    }
                                                    param_1 = local_390;
                                                    if (lVar10 != 0) {
                                                      FUN_18000cc80(plVar6,3,lVar10);
/* novels.pipeline.chapter_review_cycle#89=" → " */
                                                      if (*DAT_1803b2228 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#89=" → " */
                                                        *DAT_1803b2228 = *DAT_1803b2228 + 1;
                                                      }
                                                      FUN_18000cc80(plVar6,4);
                                                      param_1 = local_390;
                                                      plVar21 = (longlong *)
                                                                FUN_18037f0c0(local_390,plVar16,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                                              DAT_1803b2110);
                                                      if (plVar21 != (longlong *)0x0) {
                                                        lVar10 = FUN_180375520(param_1,plVar21,
/* novels.pipeline.chapter_review_cycle#59=".0f" */
                                                                               DAT_1803b2138);
                                                        if ((-1 < (int)*plVar21) &&
                                                           (*plVar21 = *plVar21 + -1, *plVar21 == 0)
                                                           ) {
                                                          lVar11 = FUN_18000cc40(plVar21);
                                                          (**(code **)(lVar11 + 0x30))(plVar21);
                                                        }
                                                        param_1 = local_390;
                                                        if (lVar10 != 0) {
                                                          FUN_18000cc80(plVar6,5,lVar10);
/* novels.pipeline.chapter_review_cycle#90="），退出循环" */
                                                          if (*DAT_1803b2230 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#90="），退出循环" */
                                                            *DAT_1803b2230 = *DAT_1803b2230 + 1;
                                                          }
                                                          FUN_18000cc80(plVar6,6);
                                                          plVar21 = (longlong *)
/* novels.pipeline.chapter_review_cycle#18="" */
                                                                    PyUnicode_Join(DAT_1803b1ff0,
                                                                                   plVar6);
                                                          if ((-1 < (int)*plVar6) &&
                                                             (*plVar6 = *plVar6 + -1, *plVar6 == 0))
                                                          {
                                                            lVar10 = FUN_18000cc40(plVar6);
                                                            (**(code **)(lVar10 + 0x30))(plVar6);
                                                          }
                                                          param_1 = local_390;
                                                          if (plVar21 == (longlong *)0x0) {
                                                            FUN_18000ce30(local_390,&local_3a8);
                                                            plVar21 = local_398;
                                                            plVar8 = local_3a0;
                                                            uVar18 = 0xec;
                                                            lVar7 = local_3a8;
                                                          }
                                                          else {
                                                            *(undefined4 *)(plVar5 + 5) = 0xec;
                                                            plVar14 = (longlong *)
                                                                      FUN_180376770(local_390,lVar7,
                                                                                    plVar21);
                                                            if ((-1 < (int)*plVar21) &&
                                                               (*plVar21 = *plVar21 + -1,
                                                               *plVar21 == 0)) {
                                                              lVar7 = FUN_18000cc40(plVar21);
                                                              (**(code **)(lVar7 + 0x30))(plVar21);
                                                            }
                                                            if (plVar14 != (longlong *)0x0)
                                                            goto LAB_1800bbc2c;
                                                            FUN_18000ce30(param_1,&local_3a8);
                                                            plVar21 = local_398;
                                                            plVar8 = local_3a0;
                                                            uVar18 = 0xec;
                                                            lVar7 = local_3a8;
                                                          }
                                                          goto LAB_1800bc74b;
                                                        }
                                                      }
                                                    }
                                                  }
                                                }
                                                uVar18 = 0xec;
                                                goto LAB_1800bbb26;
                                              }
                                              iVar3 = (int)*plVar16 + 1;
                                              if (iVar3 != 0) {
                                                *(int *)plVar16 = iVar3;
                                              }
                                              local_398 = plVar16;
                                              if ((-1 < (int)*local_378) &&
                                                 (*local_378 = *local_378 + -1, *local_378 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar21);
                                                (**(code **)(lVar7 + 0x30))(plVar21);
                                              }
                                              plVar21 = (longlong *)FUN_180038d80(local_348);
                                              param_1 = local_390;
                                            } while (plVar21 != (longlong *)0x0);
                                          }
                                          cVar2 = FUN_1800b4700(param_1);
                                          plVar6 = local_398;
                                          if (cVar2 == '\0') {
                                            FUN_18000ce30(param_1,&local_3a8);
                                            plVar21 = plVar16;
                                            plVar8 = local_3a0;
                                            uVar18 = 0xd0;
                                            lVar7 = local_3a8;
LAB_1800bc74b:
                                            local_3a8 = lVar7;
                                            lVar7 = local_3a8;
                                            FUN_18000ce70(&local_3a8);
                                            plVar16 = local_300;
                                            if (((local_300 != (longlong *)0x0) &&
                                                (-1 < (int)*local_300)) &&
                                               (*local_300 = *local_300 + -1, *local_300 == 0)) {
                                              lVar10 = FUN_18000cc40(local_300);
                                              (**(code **)(lVar10 + 0x30))(plVar16);
                                            }
                                            plVar6 = local_348;
                                            plVar16 = local_380;
                                            if ((-1 < (int)*local_348) &&
                                               (*local_348 = *local_348 + -1, plVar16 = local_380,
                                               *local_348 == 0)) {
                                              lVar10 = FUN_18000cc40(local_348);
                                              (**(code **)(lVar10 + 0x30))(plVar6);
                                              plVar16 = local_380;
                                            }
                                          }
                                          else {
LAB_1800bc7b2:
                                            local_398 = plVar6;
                                            plVar21 = local_300;
                                            plVar8 = local_3a0;
                                            local_320 = local_3a0;
                                            local_248 = plVar16;
                                            if (((local_300 != (longlong *)0x0) &&
                                                (-1 < (int)*local_300)) &&
                                               (*local_300 = *local_300 + -1, *local_300 == 0)) {
                                              lVar7 = FUN_18000cc40(local_300);
                                              (**(code **)(lVar7 + 0x30))(plVar21);
                                            }
                                            plVar21 = local_348;
                                            if ((-1 < (int)*local_348) &&
                                               (*local_348 = *local_348 + -1, *local_348 == 0)) {
                                              lVar7 = FUN_18000cc40(local_348);
                                              (**(code **)(lVar7 + 0x30))(plVar21);
                                            }
                                            plVar5 = local_318;
                                            plVar21 = plVar16;
                                            if (local_318 == (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#83="snapshots" */
                                              FUN_180370fe0(param_1,&local_3a8,DAT_1803b21f8);
                                              FUN_18000d1d0(param_1,local_3a8);
                                              plVar16 = local_380;
                                              plVar5 = local_388;
                                              uVar18 = 0xef;
                                              lVar7 = local_3a8;
                                            }
                                            else {
                                              plVar6 = (longlong *)
                                                       FUN_1800b4830(param_1,local_318,DAT_1803ab250
                                                                     ,0);
                                              if (plVar6 == (longlong *)0x0) {
                                                FUN_18000ce30(param_1,&local_3a8);
                                                plVar16 = local_380;
                                                plVar5 = local_388;
                                                uVar18 = 0xef;
                                                lVar7 = local_3a8;
                                              }
                                              else {
                                                local_330 = plVar6;
                                                plVar5 = (longlong *)
                                                         FUN_180024e50(param_1,plVar5,DAT_1803b2238)
                                                ;
                                                if (plVar5 == (longlong *)0x0) {
                                                  FUN_18000ce30(param_1,&local_3a8);
                                                  plVar16 = local_380;
                                                  plVar5 = local_388;
                                                  uVar18 = 0xf0;
                                                  lVar7 = local_3a8;
                                                }
                                                else {
                                                  plVar14 = (longlong *)
                                                            FUN_18000d280(param_1,plVar5);
                                                  if ((-1 < (int)*plVar5) &&
                                                     (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                    lVar7 = FUN_18000cc40(plVar5);
                                                    (**(code **)(lVar7 + 0x30))(plVar5);
                                                  }
                                                  if (plVar14 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_3a8);
                                                    plVar16 = local_380;
                                                    plVar5 = local_388;
                                                    uVar18 = 0xf0;
                                                    lVar7 = local_3a8;
                                                  }
                                                  else {
                                                    plVar21 = (longlong *)FUN_18000d430(plVar14);
                                                    plVar5 = local_338;
                                                    plVar15 = local_2f0;
                                                    while (local_2f0 = plVar15,
                                                          plVar21 != (longlong *)0x0) {
                                                      local_338 = plVar21;
                                                      if (((plVar5 != (longlong *)0x0) &&
                                                          (-1 < (int)*plVar5)) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        lVar7 = FUN_18000cc40();
                                                        (**(code **)(lVar7 + 0x30))(plVar5);
                                                      }
                                                      iVar3 = (int)*plVar21 + 1;
                                                      if (iVar3 != 0) {
                                                        *(int *)plVar21 = iVar3;
                                                      }
                                                      local_2f0 = plVar21;
                                                      if (((plVar15 != (longlong *)0x0) &&
                                                          (-1 < (int)*plVar15)) &&
                                                         (*plVar15 = *plVar15 + -1, *plVar15 == 0))
                                                      {
                                                        lVar7 = FUN_18000cc40(plVar15);
                                                        (**(code **)(lVar7 + 0x30))(plVar15);
                                                      }
                                                      plVar16 = (longlong *)
                                                                FUN_18037f0c0(param_1,plVar21,
/* novels.pipeline.chapter_review_cycle#55="length_in_range" */
                                                                              DAT_1803b2118);
                                                      if (plVar16 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_3a8);
                                                        uVar18 = 0xf1;
                                                        goto LAB_1800bcbe5;
                                                      }
                                                      if (local_330 == (longlong *)0x0) {
                                                        if ((-1 < (int)*plVar16) &&
                                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)
                                                           ) {
                                                          lVar7 = FUN_18000cc40(plVar16);
                                                          (**(code **)(lVar7 + 0x30))(plVar16);
                                                        }
                                                        FUN_180370fe0(param_1,&local_3a8,
/* novels.pipeline.chapter_review_cycle#92="best" */
                                                                      DAT_1803b2240);
                                                        FUN_18000d1d0(param_1,local_3a8);
                                                        uVar18 = 0xf1;
                                                        goto LAB_1800bcbe5;
                                                      }
                                                      plVar5 = (longlong *)
                                                               FUN_18037f0c0(param_1,local_330,
/* novels.pipeline.chapter_review_cycle#55="length_in_range" */
                                                                             DAT_1803b2118);
                                                      if (plVar5 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_3a8);
                                                        if ((-1 < (int)*plVar16) &&
                                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)
                                                           ) {
                                                          lVar7 = FUN_18000cc40(plVar16);
                                                          (**(code **)(lVar7 + 0x30))(plVar16);
                                                        }
                                                        uVar18 = 0xf1;
                                                        goto LAB_1800bcbe5;
                                                      }
                                                      iVar3 = FUN_180385f30(plVar16,plVar5);
                                                      if ((-1 < (int)*plVar16) &&
                                                         (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                                                      {
                                                        lVar7 = FUN_18000cc40(plVar16);
                                                        (**(code **)(lVar7 + 0x30))(plVar16);
                                                      }
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        lVar7 = FUN_18000cc40(plVar5);
                                                        (**(code **)(lVar7 + 0x30))(plVar5);
                                                      }
                                                      param_1 = local_390;
                                                      if (iVar3 == -1) {
                                                        FUN_18000ce30(local_390,&local_3a8);
                                                        uVar18 = 0xf1;
                                                        goto LAB_1800bcbe5;
                                                      }
                                                      if (iVar3 != 1) {
                                                        plVar5 = (longlong *)
                                                                 FUN_18037f0c0(local_390,plVar21,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                                               DAT_1803b2110);
                                                        plVar16 = local_330;
                                                        if (plVar5 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_3a8);
                                                          uVar18 = 0xf5;
                                                          goto LAB_1800bcbe5;
                                                        }
                                                        plVar8 = (longlong *)
                                                                 FUN_18037f0c0(param_1,local_330,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                                               DAT_1803b2110);
                                                        if (plVar8 == (longlong *)0x0) {
LAB_1800bcd82:
                                                          FUN_18000ce30(param_1,&local_3a8);
                                                          if ((-1 < (int)*plVar5) &&
                                                             (*plVar5 = *plVar5 + -1, *plVar5 == 0))
                                                          {
                                                            lVar7 = FUN_18000cc40(plVar5);
                                                            (**(code **)(lVar7 + 0x30))(plVar5);
                                                          }
                                                        }
                                                        else {
                                                          lVar7 = FUN_1800b4bf0(param_1);
                                                          if (lVar7 != 0) {
                                                            plVar15 = (longlong *)
                                                                      FUN_180381110(plVar8,lVar7);
                                                            if ((-1 < (int)*plVar8) &&
                                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0
                                                               )) {
                                                              lVar7 = FUN_18000cc40(plVar8);
                                                              (**(code **)(lVar7 + 0x30))(plVar8);
                                                            }
                                                            param_1 = local_390;
                                                            if (plVar15 == (longlong *)0x0)
                                                            goto LAB_1800bcd82;
                                                            iVar3 = FUN_180384e70(plVar5,plVar15);
                                                            if ((-1 < (int)*plVar5) &&
                                                               (*plVar5 = *plVar5 + -1, *plVar5 == 0
                                                               )) {
                                                              lVar7 = FUN_18000cc40(plVar5);
                                                              (**(code **)(lVar7 + 0x30))(plVar5);
                                                            }
                                                            if ((-1 < (int)*plVar15) &&
                                                               (*plVar15 = *plVar15 + -1,
                                                               *plVar15 == 0)) {
                                                              lVar7 = FUN_18000cc40(plVar15);
                                                              (**(code **)(lVar7 + 0x30))(plVar15);
                                                            }
                                                            param_1 = local_390;
                                                            if (iVar3 != -1) {
                                                              plVar5 = plVar16;
                                                              if (iVar3 == 1) {
                                                                iVar3 = (int)*plVar21 + 1;
                                                                if (iVar3 != 0) {
                                                                  *(int *)plVar21 = iVar3;
                                                                }
                                                                plVar5 = plVar21;
                                                                local_330 = plVar21;
                                                                if ((-1 < (int)*plVar6) &&
                                                                   (*plVar6 = *plVar6 + -1,
                                                                   *plVar6 == 0)) {
                                                                  lVar7 = FUN_18000cc40(plVar16);
                                                                  (**(code **)(lVar7 + 0x30))
                                                                            (plVar16);
                                                                }
                                                              }
                                                              param_1 = local_390;
                                                              cVar2 = FUN_18000d5b0(local_390);
                                                              if (cVar2 != '\0') goto LAB_1800bcc70;
                                                              goto LAB_1800bcbd3;
                                                            }
                                                            FUN_18000ce30(local_390,&local_3a8);
                                                            uVar18 = 0xf5;
                                                            goto LAB_1800bcbe5;
                                                          }
                                                          FUN_180373340(param_1,&local_3a8,
/* novels.pipeline.chapter_review_cycle#87="NET_IMPROVEMENT_EPSILON" */
                                                                        DAT_1803b2218);
                                                          if ((-1 < (int)*plVar5) &&
                                                             (*plVar5 = *plVar5 + -1, *plVar5 == 0))
                                                          {
                                                            lVar7 = FUN_18000cc40(plVar5);
                                                            (**(code **)(lVar7 + 0x30))(plVar5);
                                                          }
                                                          if ((-1 < (int)*plVar8) &&
                                                             (*plVar8 = *plVar8 + -1, *plVar8 == 0))
                                                          {
                                                            lVar7 = FUN_18000cc40(plVar8);
                                                            (**(code **)(lVar7 + 0x30))(plVar8);
                                                            uVar18 = 0xf5;
                                                            goto LAB_1800bcbe5;
                                                          }
                                                        }
                                                        uVar18 = 0xf5;
                                                        goto LAB_1800bcbe5;
                                                      }
                                                      plVar16 = (longlong *)
                                                                FUN_18037f0c0(local_390,plVar21,
/* novels.pipeline.chapter_review_cycle#55="length_in_range" */
                                                                              DAT_1803b2118);
                                                      if (plVar16 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_3a8);
                                                        uVar18 = 0xf2;
                                                        goto LAB_1800bcbe5;
                                                      }
                                                      iVar3 = FUN_180001510(plVar16);
                                                      if (iVar3 == -1) {
                                                        FUN_18000ce30(param_1,&local_3a8);
                                                        if ((-1 < (int)*plVar16) &&
                                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)
                                                           ) {
                                                          lVar7 = FUN_18000cc40(plVar16);
                                                          (**(code **)(lVar7 + 0x30))(plVar16);
                                                        }
                                                        uVar18 = 0xf2;
                                                        goto LAB_1800bcbe5;
                                                      }
                                                      if ((-1 < (int)*plVar16) &&
                                                         (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                                                      {
                                                        lVar7 = FUN_18000cc40(plVar16);
                                                        (**(code **)(lVar7 + 0x30))(plVar16);
                                                      }
                                                      plVar16 = local_330;
                                                      plVar5 = local_330;
                                                      if (iVar3 != 0) {
                                                        iVar3 = (int)*plVar21 + 1;
                                                        if (iVar3 != 0) {
                                                          *(int *)plVar21 = iVar3;
                                                        }
                                                        plVar5 = plVar21;
                                                        local_330 = plVar21;
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar7 = FUN_18000cc40(plVar16);
                                                          (**(code **)(lVar7 + 0x30))(plVar16);
                                                        }
                                                      }
LAB_1800bcc70:
                                                      plVar21 = (longlong *)FUN_18000d430(plVar14);
                                                      plVar6 = plVar5;
                                                      plVar16 = local_248;
                                                      plVar8 = local_320;
                                                      plVar5 = local_338;
                                                      plVar15 = local_2f0;
                                                    }
                                                    local_338 = plVar5;
                                                    cVar2 = FUN_1800b4700(param_1);
                                                    plVar21 = local_2f0;
                                                    if (cVar2 == '\0') {
LAB_1800bcbd3:
                                                      FUN_18000ce30(param_1,&local_3a8);
                                                      uVar18 = 0xf0;
LAB_1800bcbe5:
                                                      lVar7 = local_3a8;
                                                      FUN_18000ce70(&local_3a8);
                                                      plVar21 = local_2f0;
                                                      if (((local_2f0 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_2f0)) &&
                                                         (*local_2f0 = *local_2f0 + -1,
                                                         *local_2f0 == 0)) {
                                                        lVar10 = FUN_18000cc40(local_2f0);
                                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                                      }
                                                      plVar16 = local_380;
                                                      plVar5 = local_388;
                                                      plVar21 = local_398;
                                                      plVar8 = local_3a0;
                                                      if ((-1 < (int)*plVar14) &&
                                                         (*plVar14 = *plVar14 + -1, *plVar14 == 0))
                                                      {
                                                        lVar10 = FUN_18000cc40(plVar14);
                                                        (**(code **)(lVar10 + 0x30))(plVar14);
                                                        plVar16 = local_380;
                                                        plVar5 = local_388;
                                                        plVar21 = local_398;
                                                        plVar8 = local_3a0;
                                                      }
                                                    }
                                                    else {
                                                      if (((local_2f0 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_2f0)) &&
                                                         (*local_2f0 = *local_2f0 + -1,
                                                         *local_2f0 == 0)) {
                                                        lVar7 = FUN_18000cc40(local_2f0);
                                                        (**(code **)(lVar7 + 0x30))(plVar21);
                                                      }
                                                      if ((-1 < (int)*plVar14) &&
                                                         (*plVar14 = *plVar14 + -1, *plVar14 == 0))
                                                      {
                                                        lVar7 = FUN_18000cc40(plVar14);
                                                        (**(code **)(lVar7 + 0x30))(plVar14);
                                                      }
/* novels.pipeline.chapter_review_cycle#92="best" */
                                                      uVar9 = DAT_1803b2240;
                                                      if (plVar6 == (longlong *)0x0) {
LAB_1800bccf6:
                                                        FUN_180370fe0(param_1,&local_3a8,uVar9);
                                                        FUN_18000d1d0(param_1,local_3a8);
LAB_1800bcd0b:
                                                        uVar18 = 0xf8;
                                                        plVar16 = local_380;
                                                        plVar5 = local_388;
                                                        plVar21 = local_398;
                                                        plVar8 = local_3a0;
                                                        lVar7 = local_3a8;
                                                      }
                                                      else {
                                                        plVar21 = (longlong *)
                                                                  FUN_18037f0c0(param_1,plVar6,
/* novels.pipeline.chapter_review_cycle#81="content" */
                                                                                DAT_1803b21e8);
                                                        if (plVar21 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_3a8);
                                                          uVar18 = 0xf8;
                                                          plVar16 = local_380;
                                                          plVar5 = local_388;
                                                          plVar21 = local_398;
                                                          plVar8 = local_3a0;
                                                          lVar7 = local_3a8;
                                                        }
                                                        else {
                                                          if (local_3a0 == (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#81="content" */
                                                            uVar9 = DAT_1803b21e8;
                                                            if ((-1 < (int)*plVar21) &&
                                                               (*plVar21 = *plVar21 + -1,
/* novels.pipeline.chapter_review_cycle#81="content" */
                                                               uVar9 = DAT_1803b21e8, *plVar21 == 0)
                                                               ) {
                                                              lVar7 = FUN_18000cc40(plVar21);
                                                              (**(code **)(lVar7 + 0x30))(plVar21);
/* novels.pipeline.chapter_review_cycle#81="content" */
                                                              uVar9 = DAT_1803b21e8;
                                                            }
                                                            goto LAB_1800bccf6;
                                                          }
                                                          pcVar20 = (code *)FUN_180385e10(plVar21,
                                                  local_3a0);
                                                  if ((-1 < (int)*plVar21) &&
                                                     (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                    lVar7 = FUN_18000cc40(plVar21);
                                                    (**(code **)(lVar7 + 0x30))(plVar21);
                                                  }
                                                  if (pcVar20 == (code *)0x0) {
                                                    FUN_18000ce30(param_1,&local_3a8);
                                                    uVar18 = 0xf8;
                                                    plVar16 = local_380;
                                                    plVar5 = local_388;
                                                    plVar21 = local_398;
                                                    plVar8 = local_3a0;
                                                    lVar7 = local_3a8;
                                                    goto LAB_1800bcc4b;
                                                  }
                                                  iVar3 = FUN_180001510(pcVar20);
                                                  if (iVar3 == -1) {
                                                    FUN_18000ce30(param_1,&local_3a8);
                                                    if ((-1 < *(int *)pcVar20) &&
                                                       (*(longlong *)pcVar20 =
                                                             *(longlong *)pcVar20 + -1,
                                                       *(longlong *)pcVar20 == 0)) {
                                                      lVar7 = FUN_18000cc40(pcVar20);
                                                      (**(code **)(lVar7 + 0x30))(pcVar20);
                                                      uVar18 = 0xf8;
                                                      plVar16 = local_380;
                                                      plVar5 = local_388;
                                                      plVar21 = local_398;
                                                      plVar8 = local_3a0;
                                                      lVar7 = local_3a8;
                                                      goto LAB_1800bcc4b;
                                                    }
                                                    goto LAB_1800bcd0b;
                                                  }
                                                  plVar14 = local_330;
                                                  if (iVar3 == 1) {
                                                    if ((-1 < *(int *)pcVar20) &&
                                                       (*(longlong *)pcVar20 =
                                                             *(longlong *)pcVar20 + -1,
                                                       *(longlong *)pcVar20 == 0)) {
                                                      lVar7 = FUN_18000cc40(pcVar20);
                                                      (**(code **)(lVar7 + 0x30))(pcVar20);
                                                    }
                                                    plVar14 = local_330;
                                                    pcVar20 = (code *)FUN_18037f0c0(param_1,
/* novels.pipeline.chapter_review_cycle#55="length_in_range" */
                                                  local_330,DAT_1803b2118);
                                                  if (pcVar20 == (code *)0x0) {
                                                    FUN_18000ce30(param_1,&local_3a8);
                                                    uVar18 = 0xf9;
                                                    plVar16 = local_380;
                                                    plVar5 = local_388;
                                                    plVar21 = local_398;
                                                    plVar8 = local_3a0;
                                                    lVar7 = local_3a8;
                                                  }
                                                  else {
                                                    iVar3 = FUN_180001510(pcVar20);
                                                    if (iVar3 == -1) {
                                                      FUN_18000ce30(param_1,&local_3a8);
                                                      if ((-1 < *(int *)pcVar20) &&
                                                         (*(longlong *)pcVar20 =
                                                               *(longlong *)pcVar20 + -1,
                                                         *(longlong *)pcVar20 == 0)) {
                                                        lVar7 = FUN_18000cc40(pcVar20);
                                                        (**(code **)(lVar7 + 0x30))(pcVar20);
                                                      }
                                                      uVar18 = 0xf9;
                                                      plVar16 = local_380;
                                                      plVar5 = local_388;
                                                      plVar21 = local_398;
                                                      plVar8 = local_3a0;
                                                      lVar7 = local_3a8;
                                                    }
                                                    else {
                                                      plVar21 = local_398;
                                                      if (iVar3 == 1) {
                                                        if ((-1 < *(int *)pcVar20) &&
                                                           (*(longlong *)pcVar20 =
                                                                 *(longlong *)pcVar20 + -1,
                                                           *(longlong *)pcVar20 == 0)) {
                                                          lVar7 = FUN_18000cc40(pcVar20);
                                                          (**(code **)(lVar7 + 0x30))(pcVar20);
                                                        }
                                                        plVar21 = local_398;
                                                        if (local_398 == (longlong *)0x0) {
                                                          FUN_180370fe0(param_1,&local_3a8,
/* novels.pipeline.chapter_review_cycle#72="current" */
                                                                        DAT_1803b21a0);
                                                          FUN_18000d1d0(param_1,local_3a8);
                                                          plVar16 = local_380;
                                                          plVar5 = local_388;
                                                          plVar8 = local_3a0;
                                                          uVar18 = 0xf9;
                                                          lVar7 = local_3a8;
                                                        }
                                                        else {
                                                          plVar5 = (longlong *)
                                                                   FUN_18037f0c0(param_1,local_398,
/* novels.pipeline.chapter_review_cycle#55="length_in_range" */
                                                                                 DAT_1803b2118);
                                                          if (plVar5 != (longlong *)0x0) {
                                                            iVar3 = FUN_180001510(plVar5);
                                                            if ((-1 < (int)*plVar5) &&
                                                               (*plVar5 = *plVar5 + -1, *plVar5 == 0
                                                               )) {
                                                              lVar7 = FUN_18000cc40(plVar5);
                                                              (**(code **)(lVar7 + 0x30))(plVar5);
                                                            }
                                                            if (iVar3 != -1) {
                                                              pcVar20 = _Py_FalseStruct_exref;
                                                              if (iVar3 == 0) {
                                                                pcVar20 = _Py_TrueStruct_exref;
                                                              }
                                                              if (*(int *)pcVar20 + 1 != 0) {
                                                                *(int *)pcVar20 =
                                                                     *(int *)pcVar20 + 1;
                                                              }
                                                              goto LAB_1800bd121;
                                                            }
                                                          }
                                                          FUN_18000ce30(param_1,&local_3a8);
LAB_1800bd0cb:
                                                          plVar16 = local_380;
                                                          plVar5 = local_388;
                                                          plVar8 = local_3a0;
                                                          uVar18 = 0xf9;
                                                          lVar7 = local_3a8;
                                                        }
                                                      }
                                                      else {
LAB_1800bd121:
                                                        iVar3 = FUN_180001510(pcVar20);
                                                        if (iVar3 == -1) {
                                                          FUN_18000ce30(param_1,&local_3a8);
                                                          if ((*(int *)pcVar20 < 0) ||
                                                             (*(longlong *)pcVar20 =
                                                                   *(longlong *)pcVar20 + -1,
                                                             *(longlong *)pcVar20 != 0))
                                                          goto LAB_1800bd0cb;
                                                          lVar7 = FUN_18000cc40(pcVar20);
                                                          (**(code **)(lVar7 + 0x30))(pcVar20);
                                                          plVar16 = local_380;
                                                          plVar5 = local_388;
                                                          plVar8 = local_3a0;
                                                          uVar18 = 0xf9;
                                                          lVar7 = local_3a8;
                                                        }
                                                        else {
                                                          if (iVar3 == 1) goto LAB_1800bd351;
                                                          if ((-1 < *(int *)pcVar20) &&
                                                             (*(longlong *)pcVar20 =
                                                                   *(longlong *)pcVar20 + -1,
                                                             *(longlong *)pcVar20 == 0)) {
                                                            lVar7 = FUN_18000cc40(pcVar20);
                                                            (**(code **)(lVar7 + 0x30))(pcVar20);
                                                          }
                                                          plVar5 = (longlong *)
                                                                   FUN_18037f0c0(param_1,plVar14,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                                                 DAT_1803b2110);
                                                          if (plVar5 == (longlong *)0x0) {
LAB_1800bd1a0:
                                                            FUN_18000ce30(param_1,&local_3a8);
                                                            plVar16 = local_380;
                                                            plVar5 = local_388;
                                                            plVar8 = local_3a0;
                                                            uVar18 = 0xfa;
                                                            lVar7 = local_3a8;
                                                          }
                                                          else if (plVar21 == (longlong *)0x0) {
                                                            if ((-1 < (int)*plVar5) &&
                                                               (*plVar5 = *plVar5 + -1, *plVar5 == 0
                                                               )) {
                                                              lVar7 = FUN_18000cc40(plVar5);
                                                              (**(code **)(lVar7 + 0x30))(plVar5);
                                                            }
                                                            FUN_180370fe0(param_1,&local_3a8,
/* novels.pipeline.chapter_review_cycle#72="current" */
                                                                          DAT_1803b21a0);
                                                            FUN_18000d1d0(param_1,local_3a8);
                                                            plVar16 = local_380;
                                                            plVar5 = local_388;
                                                            plVar8 = local_3a0;
                                                            uVar18 = 0xfa;
                                                            lVar7 = local_3a8;
                                                          }
                                                          else {
                                                            plVar6 = (longlong *)
                                                                     FUN_18037f0c0(param_1,plVar21,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                                                   DAT_1803b2110);
                                                            if (plVar6 == (longlong *)0x0) {
LAB_1800bd21f:
                                                              FUN_18000ce30(param_1,&local_3a8);
                                                              if ((-1 < (int)*plVar5) &&
                                                                 (*plVar5 = *plVar5 + -1,
                                                                 *plVar5 == 0)) {
                                                                lVar7 = FUN_18000cc40(plVar5);
                                                                (**(code **)(lVar7 + 0x30))(plVar5);
                                                              }
                                                            }
                                                            else {
                                                              lVar7 = FUN_1800b4bf0(param_1);
                                                              if (lVar7 != 0) {
                                                                plVar15 = (longlong *)
                                                                          FUN_180381110(plVar6,lVar7
                                                                                       );
                                                                if ((-1 < (int)*plVar6) &&
                                                                   (*plVar6 = *plVar6 + -1,
                                                                   *plVar6 == 0)) {
                                                                  lVar7 = FUN_18000cc40(plVar6);
                                                                  (**(code **)(lVar7 + 0x30))
                                                                            (plVar6);
                                                                }
                                                                param_1 = local_390;
                                                                if (plVar15 != (longlong *)0x0) {
                                                                  pcVar20 = (code *)FUN_180384d50(
                                                  plVar5,plVar15);
                                                  if ((-1 < (int)*plVar5) &&
                                                     (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                    lVar7 = FUN_18000cc40(plVar5);
                                                    (**(code **)(lVar7 + 0x30))(plVar5);
                                                  }
                                                  if ((-1 < (int)*plVar15) &&
                                                     (*plVar15 = *plVar15 + -1, *plVar15 == 0)) {
                                                    lVar7 = FUN_18000cc40(plVar15);
                                                    (**(code **)(lVar7 + 0x30))(plVar15);
                                                  }
                                                  param_1 = local_390;
                                                  if (pcVar20 != (code *)0x0) goto LAB_1800bd351;
                                                  goto LAB_1800bd1a0;
                                                  }
                                                  goto LAB_1800bd21f;
                                                  }
/* novels.pipeline.chapter_review_cycle#87="NET_IMPROVEMENT_EPSILON" */
                                                  FUN_180373340(param_1,&local_3a8,DAT_1803b2218);
                                                  if ((-1 < (int)*plVar5) &&
                                                     (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                    lVar7 = FUN_18000cc40(plVar5);
                                                    (**(code **)(lVar7 + 0x30))(plVar5);
                                                  }
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    lVar7 = FUN_18000cc40(plVar6);
                                                    (**(code **)(lVar7 + 0x30))(plVar6);
                                                    plVar16 = local_380;
                                                    plVar5 = local_388;
                                                    plVar8 = local_3a0;
                                                    uVar18 = 0xfa;
                                                    lVar7 = local_3a8;
                                                    goto LAB_1800bcc4b;
                                                  }
                                                  }
                                                  plVar16 = local_380;
                                                  plVar5 = local_388;
                                                  plVar8 = local_3a0;
                                                  uVar18 = 0xfa;
                                                  lVar7 = local_3a8;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  else {
LAB_1800bd351:
                                                    local_2e0 = pcVar20;
                                                    iVar3 = FUN_180001510(pcVar20);
/* novels.pipeline.chapter_review_cycle#93="回退到最高分版本（" */
                                                    piVar1 = DAT_1803b2248;
                                                    if (iVar3 == -1) {
                                                      FUN_18000ce30(param_1,&local_3a8);
                                                      uVar18 = 0xfc;
                                                      plVar16 = local_380;
                                                      plVar5 = local_388;
                                                      plVar21 = local_398;
                                                      plVar8 = local_3a0;
                                                      lVar7 = local_3a8;
                                                    }
                                                    else {
                                                      plVar5 = local_388;
                                                      plVar21 = local_398;
                                                      if (iVar3 == 0) {
LAB_1800bd6a6:
                                                        if (local_310 == (longlong *)0x0) {
                                                          FUN_180370fe0(param_1,&local_3a8,
/* novels.pipeline.chapter_review_cycle#95="iterations" */
                                                                        DAT_1803b2258);
                                                          FUN_18000d1d0(param_1,local_3a8);
                                                          plVar16 = local_380;
                                                          plVar8 = local_3a0;
                                                          uVar18 = 0x101;
                                                          lVar7 = local_3a8;
                                                        }
                                                        else {
                                                          iVar3 = FUN_1803858f0(local_310,
                                                                                DAT_1803ab250);
                                                          if (iVar3 == -1) {
                                                            FUN_18000ce30(param_1,&local_3a8);
                                                            plVar16 = local_380;
                                                            plVar8 = local_3a0;
                                                            uVar18 = 0x101;
                                                            lVar7 = local_3a8;
                                                          }
                                                          else if (iVar3 == 1) {
                                                            lVar7 = FUN_1800b4fc0(param_1);
                                                            param_1 = local_390;
                                                            if (lVar7 == 0) {
                                                              FUN_180373340(local_390,&local_3a8,
/* novels.pipeline.chapter_review_cycle#57="_is_passed" */
                                                                            DAT_1803b2128);
                                                              plVar16 = local_380;
                                                              plVar8 = local_3a0;
                                                              uVar18 = 0x102;
                                                              lVar7 = local_3a8;
                                                            }
                                                            else if (plVar21 == (longlong *)0x0) {
                                                              FUN_180370fe0(local_390,&local_3a8,
/* novels.pipeline.chapter_review_cycle#72="current" */
                                                                            DAT_1803b21a0);
                                                              FUN_18000d1d0(param_1,local_3a8);
                                                              plVar16 = local_380;
                                                              plVar8 = local_3a0;
                                                              uVar18 = 0x102;
                                                              lVar7 = local_3a8;
                                                            }
                                                            else {
                                                              plVar16 = (longlong *)
                                                                        FUN_18037f0c0(local_390,
                                                                                      plVar21,
/* novels.pipeline.chapter_review_cycle#49="audit" */
                                                  DAT_1803b20e8);
                                                  if (plVar16 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_3a8);
                                                    plVar16 = local_380;
                                                    plVar8 = local_3a0;
                                                    uVar18 = 0x102;
                                                    lVar7 = local_3a8;
                                                  }
                                                  else {
                                                    plVar6 = (longlong *)
                                                             FUN_18037f0c0(param_1,plVar21,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                                           DAT_1803b2110);
                                                    if (plVar6 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_3a8);
                                                      if ((-1 < (int)*plVar16) &&
                                                         (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                                                      {
                                                        lVar7 = FUN_18000cc40(plVar16);
                                                        (**(code **)(lVar7 + 0x30))(plVar16);
                                                      }
                                                      plVar16 = local_380;
                                                      plVar8 = local_3a0;
                                                      uVar18 = 0x102;
                                                      lVar7 = local_3a8;
                                                    }
                                                    else {
                                                      plVar21 = (longlong *)
                                                                FUN_18037f0c0(param_1,plVar21,
/* novels.pipeline.chapter_review_cycle#55="length_in_range" */
                                                                              DAT_1803b2118);
                                                      if (plVar21 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_3a8);
                                                        if ((-1 < (int)*plVar16) &&
                                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)
                                                           ) {
                                                          lVar7 = FUN_18000cc40(plVar16);
                                                          (**(code **)(lVar7 + 0x30))(plVar16);
                                                        }
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar7 = FUN_18000cc40(plVar6);
                                                          (**(code **)(lVar7 + 0x30))(plVar6);
                                                        }
                                                        plVar16 = local_380;
                                                        plVar21 = local_398;
                                                        plVar8 = local_3a0;
                                                        uVar18 = 0x102;
                                                        lVar7 = local_3a8;
                                                      }
                                                      else if (local_2d8 == (longlong *)0x0) {
                                                        if ((-1 < (int)*plVar16) &&
                                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)
                                                           ) {
                                                          lVar7 = FUN_18000cc40(plVar16);
                                                          (**(code **)(lVar7 + 0x30))(plVar16);
                                                        }
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar7 = FUN_18000cc40(plVar6);
                                                          (**(code **)(lVar7 + 0x30))(plVar6);
                                                        }
                                                        if ((-1 < (int)*plVar21) &&
                                                           (*plVar21 = *plVar21 + -1, *plVar21 == 0)
                                                           ) {
                                                          lVar7 = FUN_18000cc40(plVar21);
                                                          (**(code **)(lVar7 + 0x30))(plVar21);
                                                        }
                                                        FUN_180370fe0(param_1,&local_3a8,
/* novels.pipeline.chapter_review_cycle#84="pass_score" */
                                                                      DAT_1803b2200);
                                                        FUN_18000d1d0(param_1,local_3a8);
                                                        plVar16 = local_380;
                                                        plVar21 = local_398;
                                                        plVar8 = local_3a0;
                                                        uVar18 = 0x102;
                                                        lVar7 = local_3a8;
                                                      }
                                                      else {
                                                        uVar18 = 0x102;
                                                        *(undefined4 *)(plVar5 + 5) = 0x102;
                                                        local_148 = local_2d8;
                                                        local_160 = plVar16;
                                                        local_158 = plVar6;
                                                        local_150 = plVar21;
                                                        plVar14 = (longlong *)
                                                                  FUN_180379a80(param_1,lVar7,
                                                                                &local_160);
                                                        if ((-1 < (int)*plVar16) &&
                                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)
                                                           ) {
                                                          lVar7 = FUN_18000cc40(plVar16);
                                                          (**(code **)(lVar7 + 0x30))(plVar16);
                                                        }
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar7 = FUN_18000cc40(plVar6);
                                                          (**(code **)(lVar7 + 0x30))(plVar6);
                                                        }
                                                        if ((-1 < (int)*plVar21) &&
                                                           (*plVar21 = *plVar21 + -1, *plVar21 == 0)
                                                           ) {
                                                          lVar7 = FUN_18000cc40(plVar21);
                                                          (**(code **)(lVar7 + 0x30))(plVar21);
                                                        }
                                                        if (plVar14 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_3a8);
                                                          plVar16 = local_380;
                                                          plVar21 = local_398;
                                                          plVar8 = local_3a0;
                                                          lVar7 = local_3a8;
                                                        }
                                                        else {
                                                          iVar3 = FUN_180001510(plVar14);
                                                          if (iVar3 == -1) {
                                                            FUN_18000ce30(param_1,&local_3a8);
                                                            plVar16 = local_380;
                                                            plVar21 = local_398;
                                                            plVar8 = local_3a0;
                                                            lVar7 = local_3a8;
                                                            if ((-1 < (int)*plVar14) &&
                                                               (*plVar14 = *plVar14 + -1,
                                                               plVar16 = local_380, *plVar14 == 0))
                                                            {
                                                              lVar7 = FUN_18000cc40(plVar14);
                                                              (**(code **)(lVar7 + 0x30))(plVar14);
                                                              plVar16 = local_380;
                                                              plVar21 = local_398;
                                                              plVar8 = local_3a0;
                                                              lVar7 = local_3a8;
                                                            }
                                                          }
                                                          else {
                                                            if ((-1 < (int)*plVar14) &&
                                                               (*plVar14 = *plVar14 + -1,
                                                               *plVar14 == 0)) {
                                                              lVar7 = FUN_18000cc40(plVar14);
                                                              (**(code **)(lVar7 + 0x30))(plVar14);
                                                            }
/* novels.pipeline.chapter_review_cycle#96="最终审稿 " */
                                                            piVar1 = DAT_1803b2260;
                                                            lVar7 = local_328[2];
                                                            if (iVar3 == 0) {
                                                              if (lVar7 != 0) {
                                                                plVar6 = (longlong *)
                                                                         FUN_180374a40(param_1,4);
                                                                if (*piVar1 + 1 != 0) {
                                                                  *piVar1 = *piVar1 + 1;
                                                                }
                                                                FUN_18000cc80(plVar6,0,piVar1);
                                                                plVar21 = (longlong *)
                                                                          FUN_18037f0c0(param_1,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                  local_398,DAT_1803b2110);
                                                  if (plVar21 != (longlong *)0x0) {
                                                    lVar10 = FUN_180375520(param_1,plVar21,
/* novels.pipeline.chapter_review_cycle#59=".0f" */
                                                                           DAT_1803b2138);
                                                    if ((-1 < (int)*plVar21) &&
                                                       (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                                      lVar11 = FUN_18000cc40(plVar21);
                                                      (**(code **)(lVar11 + 0x30))(plVar21);
                                                    }
                                                    if (lVar10 == 0) {
                                                      uVar19 = 0x105;
                                                      goto LAB_1800bdb26;
                                                    }
                                                    FUN_18000cc80(plVar6,1,lVar10);
/* novels.pipeline.chapter_review_cycle#98=" 分，未达通过线 " */
                                                    if (*DAT_1803b2270 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#98=" 分，未达通过线 " */
                                                      *DAT_1803b2270 = *DAT_1803b2270 + 1;
                                                    }
                                                    FUN_18000cc80(plVar6,2);
                                                    lVar10 = FUN_180375520(param_1,local_2d8,
/* novels.pipeline.chapter_review_cycle#59=".0f" */
                                                                           DAT_1803b2138);
                                                    if (lVar10 != 0) {
                                                      FUN_18000cc80(plVar6,3,lVar10);
                                                      plVar21 = (longlong *)
/* novels.pipeline.chapter_review_cycle#18="" */
                                                                PyUnicode_Join(DAT_1803b1ff0,plVar6)
                                                      ;
                                                      if ((-1 < (int)*plVar6) &&
                                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                        lVar10 = FUN_18000cc40(plVar6);
                                                        (**(code **)(lVar10 + 0x30))(plVar6);
                                                      }
                                                      if (plVar21 != (longlong *)0x0) {
                                                        uVar18 = 0x105;
                                                        goto LAB_1800bdda2;
                                                      }
                                                      FUN_18000ce30(param_1,&local_3a8);
                                                      plVar16 = local_380;
                                                      plVar21 = local_398;
                                                      plVar8 = local_3a0;
                                                      uVar18 = 0x105;
                                                      lVar7 = local_3a8;
                                                      goto LAB_1800bcc4b;
                                                    }
                                                  }
                                                  uVar19 = 0x105;
                                                  goto LAB_1800bdb26;
                                                  }
/* novels.pipeline.chapter_review_cycle#41="log" */
                                                  FUN_180370fe0(param_1,&local_3a8,DAT_1803b20a8);
                                                  FUN_18000d1d0(param_1,local_3a8);
                                                  plVar16 = local_380;
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                  uVar18 = 0x105;
                                                  lVar7 = local_3a8;
                                                  }
                                                  else if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#41="log" */
                                                    FUN_180370fe0(param_1,&local_3a8,DAT_1803b20a8);
                                                    FUN_18000d1d0(param_1,local_3a8);
                                                    plVar16 = local_380;
                                                    plVar21 = local_398;
                                                    plVar8 = local_3a0;
                                                    uVar18 = 0x103;
                                                    lVar7 = local_3a8;
                                                  }
                                                  else {
                                                    plVar6 = (longlong *)FUN_180374a40(param_1,3);
                                                    if (*piVar1 + 1 != 0) {
                                                      *piVar1 = *piVar1 + 1;
                                                    }
                                                    FUN_18000cc80(plVar6,0,piVar1);
                                                    plVar21 = (longlong *)
                                                              FUN_18037f0c0(param_1,local_398,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                                            DAT_1803b2110);
                                                    if (plVar21 != (longlong *)0x0) {
                                                      lVar10 = FUN_180375520(param_1,plVar21,
/* novels.pipeline.chapter_review_cycle#59=".0f" */
                                                                             DAT_1803b2138);
                                                      if ((-1 < (int)*plVar21) &&
                                                         (*plVar21 = *plVar21 + -1, *plVar21 == 0))
                                                      {
                                                        lVar11 = FUN_18000cc40(plVar21);
                                                        (**(code **)(lVar11 + 0x30))(plVar21);
                                                      }
                                                      if (lVar10 != 0) {
                                                        FUN_18000cc80(plVar6,1,lVar10);
/* novels.pipeline.chapter_review_cycle#97=" 分，已通过" */
                                                        if (*DAT_1803b2268 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#97=" 分，已通过" */
                                                          *DAT_1803b2268 = *DAT_1803b2268 + 1;
                                                        }
                                                        FUN_18000cc80(plVar6,2);
                                                        plVar21 = (longlong *)
/* novels.pipeline.chapter_review_cycle#18="" */
                                                                  PyUnicode_Join(DAT_1803b1ff0,
                                                                                 plVar6);
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar10 = FUN_18000cc40(plVar6);
                                                          (**(code **)(lVar10 + 0x30))(plVar6);
                                                        }
                                                        if (plVar21 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_3a8);
                                                          plVar16 = local_380;
                                                          plVar21 = local_398;
                                                          plVar8 = local_3a0;
                                                          uVar18 = 0x103;
                                                          lVar7 = local_3a8;
                                                        }
                                                        else {
                                                          uVar18 = 0x103;
LAB_1800bdda2:
                                                          *(undefined4 *)(plVar5 + 5) = uVar18;
                                                          plVar16 = (longlong *)
                                                                    FUN_180376770(param_1,lVar7,
                                                                                  plVar21);
                                                          if ((-1 < (int)*plVar21) &&
                                                             (*plVar21 = *plVar21 + -1,
                                                             *plVar21 == 0)) {
                                                            lVar7 = FUN_18000cc40(plVar21);
                                                            (**(code **)(lVar7 + 0x30))(plVar21);
                                                          }
                                                          if (plVar16 != (longlong *)0x0) {
                                                            if ((-1 < (int)*plVar16) &&
                                                               (*plVar16 = *plVar16 + -1,
                                                               *plVar16 == 0)) {
                                                              lVar7 = FUN_18000cc40(plVar16);
                                                              (**(code **)(lVar7 + 0x30))(plVar16);
                                                            }
                                                            goto LAB_1800bde0e;
                                                          }
                                                          FUN_18000ce30(param_1,&local_3a8);
                                                          plVar16 = local_380;
                                                          plVar21 = local_398;
                                                          plVar8 = local_3a0;
                                                          lVar7 = local_3a8;
                                                        }
                                                        goto LAB_1800bcc4b;
                                                      }
                                                    }
                                                    uVar19 = 0x103;
LAB_1800bdb26:
                                                    local_368 = uVar19;
                                                    FUN_18000ce30(param_1,&local_3a8);
                                                    plVar16 = local_380;
                                                    plVar21 = local_398;
                                                    plVar8 = local_3a0;
                                                    uVar18 = local_368;
                                                    lVar7 = local_3a8;
                                                    if ((-1 < (int)*plVar6) &&
                                                       (*plVar6 = *plVar6 + -1, plVar16 = local_380,
                                                       *plVar6 == 0)) {
                                                      lVar7 = FUN_18000cc40(plVar6);
                                                      (**(code **)(lVar7 + 0x30))(plVar6);
                                                      plVar16 = local_380;
                                                      plVar21 = local_398;
                                                      plVar8 = local_3a0;
                                                      uVar18 = uVar19;
                                                      lVar7 = local_3a8;
                                                    }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  else {
LAB_1800bde0e:
                                                    lVar7 = FUN_1800b4ae0(param_1);
                                                    plVar21 = local_398;
                                                    plVar8 = local_3a0;
                                                    if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#52="ChapterReviewCycleResult" */
                                                      FUN_180373340(param_1,&local_3a8,DAT_1803b2100
                                                                   );
                                                      plVar16 = local_380;
                                                      plVar21 = local_398;
                                                      plVar8 = local_3a0;
                                                      uVar18 = 0x107;
                                                      lVar7 = local_3a8;
                                                    }
                                                    else if (local_3a0 == (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#81="content" */
                                                      FUN_180370fe0(param_1,&local_3a8,DAT_1803b21e8
                                                                   );
                                                      FUN_18000d1d0(param_1,local_3a8);
                                                      plVar16 = local_380;
                                                      plVar21 = local_398;
                                                      uVar18 = 0x108;
                                                      lVar7 = local_3a8;
                                                    }
                                                    else if (local_398 == (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#72="current" */
                                                      FUN_180370fe0(param_1,&local_3a8,DAT_1803b21a0
                                                                   );
                                                      FUN_18000d1d0(param_1,local_3a8);
                                                      plVar16 = local_380;
                                                      uVar18 = 0x109;
                                                      lVar7 = local_3a8;
                                                    }
                                                    else {
                                                      plVar16 = (longlong *)
                                                                FUN_18037f0c0(param_1,local_398,
/* novels.pipeline.chapter_review_cycle#53="word_count" */
                                                                              DAT_1803b2108);
                                                      if (plVar16 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_3a8);
                                                        plVar16 = local_380;
                                                        uVar18 = 0x109;
                                                        lVar7 = local_3a8;
                                                      }
                                                      else {
                                                        plVar6 = (longlong *)
                                                                 FUN_180375370(param_1,local_318);
                                                        if (plVar6 == (longlong *)0x0) {
LAB_1800bdef7:
                                                          FUN_18000ce30(param_1,&local_3a8);
                                                          if ((-1 < (int)*plVar16) &&
                                                             (*plVar16 = *plVar16 + -1,
                                                             *plVar16 == 0)) {
                                                            lVar7 = FUN_18000cc40(plVar16);
                                                            (**(code **)(lVar7 + 0x30))(plVar16);
                                                          }
                                                          plVar16 = local_380;
                                                          uVar18 = 0x10a;
                                                          lVar7 = local_3a8;
                                                        }
                                                        else {
                                                          cVar2 = thunk_FUN_180385d40(plVar6,1);
                                                          if ((-1 < (int)*plVar6) &&
                                                             (*plVar6 = *plVar6 + -1, *plVar6 == 0))
                                                          {
                                                            lVar10 = FUN_18000cc40(plVar6);
                                                            (**(code **)(lVar10 + 0x30))(plVar6);
                                                          }
                                                          pcVar20 = _Py_FalseStruct_exref;
                                                          if (cVar2 != '\0') {
                                                            pcVar20 = _Py_TrueStruct_exref;
                                                          }
                                                          iVar3 = FUN_180001510(pcVar20);
                                                          if (iVar3 == 1) {
                                                            pcVar20 = (code *)FUN_180385e10(plVar8,
                                                  local_290);
                                                  param_1 = local_390;
                                                  if (pcVar20 == (code *)0x0) goto LAB_1800bdef7;
                                                  }
                                                  else if (*(int *)pcVar20 + 1 != 0) {
                                                    *(int *)pcVar20 = *(int *)pcVar20 + 1;
                                                  }
                                                  param_1 = local_390;
                                                  plVar5 = (longlong *)
                                                           FUN_18037f0c0(local_390,plVar21,
/* novels.pipeline.chapter_review_cycle#49="audit" */
                                                                         DAT_1803b20e8);
                                                  if (plVar5 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_3a8);
                                                    if ((-1 < (int)*plVar16) &&
                                                       (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                      lVar7 = FUN_18000cc40(plVar16);
                                                      (**(code **)(lVar7 + 0x30))(plVar16);
                                                    }
                                                    if ((-1 < *(int *)pcVar20) &&
                                                       (*(longlong *)pcVar20 =
                                                             *(longlong *)pcVar20 + -1,
                                                       *(longlong *)pcVar20 == 0)) {
                                                      lVar7 = FUN_18000cc40(pcVar20);
                                                      (**(code **)(lVar7 + 0x30))(pcVar20);
                                                    }
                                                    plVar16 = local_380;
                                                    plVar5 = local_388;
                                                    uVar18 = 0x10b;
                                                    lVar7 = local_3a8;
                                                  }
                                                  else {
                                                    plVar6 = (longlong *)
                                                             FUN_18037f0c0(param_1,plVar21,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                                           DAT_1803b2110);
                                                    if (plVar6 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_3a8);
                                                      if ((-1 < (int)*plVar16) &&
                                                         (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                                                      {
                                                        lVar7 = FUN_18000cc40(plVar16);
                                                        (**(code **)(lVar7 + 0x30))(plVar16);
                                                      }
                                                      if ((-1 < *(int *)pcVar20) &&
                                                         (*(longlong *)pcVar20 =
                                                               *(longlong *)pcVar20 + -1,
                                                         *(longlong *)pcVar20 == 0)) {
                                                        lVar7 = FUN_18000cc40(pcVar20);
                                                        (**(code **)(lVar7 + 0x30))(pcVar20);
                                                      }
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        lVar7 = FUN_18000cc40(plVar5);
                                                        (**(code **)(lVar7 + 0x30))(plVar5);
                                                      }
                                                      plVar16 = local_380;
                                                      plVar5 = local_388;
                                                      plVar21 = local_398;
                                                      uVar18 = 0x10c;
                                                      lVar7 = local_3a8;
                                                    }
                                                    else {
                                                      plVar21 = (longlong *)
                                                                FUN_18037f0c0(param_1,local_398,
/* novels.pipeline.chapter_review_cycle#55="length_in_range" */
                                                                              DAT_1803b2118);
                                                      if (plVar21 != (longlong *)0x0) {
                                                        local_368 = 0x107;
                                                        *(undefined4 *)(local_388 + 5) = 0x107;
                                                        local_98 = local_3a0;
                                                        local_68 = local_310;
                                                        local_60 = local_2e0;
                                                        local_58 = local_318;
                                                        local_90 = plVar16;
                                                        local_88 = pcVar20;
                                                        local_80 = plVar5;
                                                        local_78 = plVar6;
                                                        local_70 = plVar21;
                                                        lVar7 = FUN_18037ccb0(param_1,lVar7,
                                                                              &local_98,
                                                                              DAT_1803b2278);
                                                        if ((-1 < (int)*plVar16) &&
                                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)
                                                           ) {
                                                          lVar10 = FUN_18000cc40(plVar16);
                                                          (**(code **)(lVar10 + 0x30))(plVar16);
                                                        }
                                                        if ((-1 < *(int *)pcVar20) &&
                                                           (*(longlong *)pcVar20 =
                                                                 *(longlong *)pcVar20 + -1,
                                                           *(longlong *)pcVar20 == 0)) {
                                                          lVar10 = FUN_18000cc40(pcVar20);
                                                          (**(code **)(lVar10 + 0x30))(pcVar20);
                                                        }
                                                        if ((-1 < (int)*plVar5) &&
                                                           (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                          lVar10 = FUN_18000cc40(plVar5);
                                                          (**(code **)(lVar10 + 0x30))(plVar5);
                                                        }
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          lVar10 = FUN_18000cc40(plVar6);
                                                          (**(code **)(lVar10 + 0x30))(plVar6);
                                                        }
                                                        if ((-1 < (int)*plVar21) &&
                                                           (*plVar21 = *plVar21 + -1, *plVar21 == 0)
                                                           ) {
                                                          lVar10 = FUN_18000cc40(plVar21);
                                                          goto LAB_1800b9d3d;
                                                        }
                                                        goto LAB_1800b9d43;
                                                      }
                                                      FUN_18000ce30(param_1,&local_3a8);
                                                      if ((-1 < (int)*plVar16) &&
                                                         (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                                                      {
                                                        lVar7 = FUN_18000cc40(plVar16);
                                                        (**(code **)(lVar7 + 0x30))(plVar16);
                                                      }
                                                      if ((-1 < *(int *)pcVar20) &&
                                                         (*(longlong *)pcVar20 =
                                                               *(longlong *)pcVar20 + -1,
                                                         *(longlong *)pcVar20 == 0)) {
                                                        lVar7 = FUN_18000cc40(pcVar20);
                                                        (**(code **)(lVar7 + 0x30))(pcVar20);
                                                      }
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        lVar7 = FUN_18000cc40(plVar5);
                                                        (**(code **)(lVar7 + 0x30))(plVar5);
                                                      }
                                                      if ((-1 < (int)*plVar6) &&
                                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                        lVar7 = FUN_18000cc40(plVar6);
                                                        (**(code **)(lVar7 + 0x30))(plVar6);
                                                      }
                                                      uVar18 = 0x10d;
                                                      plVar16 = local_380;
                                                      plVar5 = local_388;
                                                      plVar21 = local_398;
                                                      plVar8 = local_3a0;
                                                      lVar7 = local_3a8;
                                                    }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  else {
                                                    lVar7 = local_328[2];
                                                    if (lVar7 != 0) {
                                                      plVar6 = (longlong *)FUN_180374a40(param_1,5);
                                                      if (*piVar1 + 1 != 0) {
                                                        *piVar1 = *piVar1 + 1;
                                                      }
                                                      FUN_18000cc80(plVar6,0,piVar1);
                                                      plVar21 = (longlong *)
                                                                FUN_18037f0c0(param_1,plVar14,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                                              DAT_1803b2110);
                                                      if (plVar21 != (longlong *)0x0) {
                                                        lVar10 = FUN_180375520(param_1,plVar21,
/* novels.pipeline.chapter_review_cycle#59=".0f" */
                                                                               DAT_1803b2138);
                                                        if ((-1 < (int)*plVar21) &&
                                                           (*plVar21 = *plVar21 + -1, *plVar21 == 0)
                                                           ) {
                                                          lVar11 = FUN_18000cc40(plVar21);
                                                          (**(code **)(lVar11 + 0x30))(plVar21);
                                                        }
                                                        param_1 = local_390;
                                                        if (lVar10 != 0) {
                                                          FUN_18000cc80(plVar6,1,lVar10);
/* novels.pipeline.chapter_review_cycle#94=" 分 vs 当前 " */
                                                          if (*DAT_1803b2250 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#94=" 分 vs 当前 " */
                                                            *DAT_1803b2250 = *DAT_1803b2250 + 1;
                                                          }
                                                          FUN_18000cc80(plVar6,2);
                                                          param_1 = local_390;
                                                          if (local_398 == (longlong *)0x0) {
                                                            FUN_180370fe0(local_390,&local_3a8,
/* novels.pipeline.chapter_review_cycle#72="current" */
                                                                          DAT_1803b21a0);
                                                            FUN_18000d1d0(param_1,local_3a8);
                                                            uVar19 = 0xfd;
                                                            local_368 = 0xfd;
                                                            goto LAB_1800ba041;
                                                          }
                                                          plVar21 = (longlong *)
                                                                    FUN_18037f0c0(local_390,
                                                                                  local_398,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                                                  DAT_1803b2110);
                                                          if (plVar21 != (longlong *)0x0) {
                                                            lVar10 = FUN_180375520(param_1,plVar21,
/* novels.pipeline.chapter_review_cycle#59=".0f" */
                                                                                   DAT_1803b2138);
                                                            if ((-1 < (int)*plVar21) &&
                                                               (*plVar21 = *plVar21 + -1,
                                                               *plVar21 == 0)) {
                                                              lVar11 = FUN_18000cc40(plVar21);
                                                              (**(code **)(lVar11 + 0x30))(plVar21);
                                                            }
                                                            param_1 = local_390;
                                                            if (lVar10 != 0) {
                                                              FUN_18000cc80(plVar6,3,lVar10);
/* novels.pipeline.chapter_review_cycle#60=" 分）" */
                                                              if (*DAT_1803b2140 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#60=" 分）" */
                                                                *DAT_1803b2140 = *DAT_1803b2140 + 1;
                                                              }
                                                              FUN_18000cc80(plVar6,4);
                                                              plVar21 = (longlong *)
/* novels.pipeline.chapter_review_cycle#18="" */
                                                                        PyUnicode_Join(DAT_1803b1ff0
                                                                                       ,plVar6);
                                                              if ((-1 < (int)*plVar6) &&
                                                                 (*plVar6 = *plVar6 + -1,
                                                                 *plVar6 == 0)) {
                                                                lVar10 = FUN_18000cc40(plVar6);
                                                                (**(code **)(lVar10 + 0x30))(plVar6)
                                                                ;
                                                              }
                                                              plVar5 = local_388;
                                                              param_1 = local_390;
                                                              if (plVar21 == (longlong *)0x0) {
                                                                FUN_18000ce30(local_390,&local_3a8);
                                                                uVar18 = 0xfd;
                                                                plVar16 = local_380;
                                                                plVar5 = local_388;
                                                                plVar21 = local_398;
                                                                plVar8 = local_3a0;
                                                                lVar7 = local_3a8;
                                                              }
                                                              else {
                                                                local_368 = 0xfd;
                                                                *(undefined4 *)(local_388 + 5) =
                                                                     0xfd;
                                                                plVar6 = (longlong *)
                                                                         FUN_180376770(local_390,
                                                                                       lVar7,plVar21
                                                                                      );
                                                                if ((-1 < (int)*plVar21) &&
                                                                   (*plVar21 = *plVar21 + -1,
                                                                   *plVar21 == 0)) {
                                                                  lVar7 = FUN_18000cc40(plVar21);
                                                                  (**(code **)(lVar7 + 0x30))
                                                                            (plVar21);
                                                                }
                                                                if (plVar6 == (longlong *)0x0) {
                                                                  FUN_18000ce30(param_1,&local_3a8);
                                                                  plVar16 = local_380;
                                                                  plVar21 = local_398;
                                                                  plVar8 = local_3a0;
                                                                  uVar18 = local_368;
                                                                  lVar7 = local_3a8;
                                                                }
                                                                else {
                                                                  if ((-1 < (int)*plVar6) &&
                                                                     (*plVar6 = *plVar6 + -1,
                                                                     *plVar6 == 0)) {
                                                                    lVar7 = FUN_18000cc40(plVar6);
                                                                    (**(code **)(lVar7 + 0x30))
                                                                              (plVar6);
                                                                  }
                                                                  plVar21 = local_330;
                                                                  plVar14 = (longlong *)
                                                                            FUN_18037f0c0(param_1,
/* novels.pipeline.chapter_review_cycle#81="content" */
                                                  local_330,DAT_1803b21e8);
                                                  plVar6 = local_3a0;
                                                  if (plVar14 != (longlong *)0x0) {
                                                    local_3a0 = plVar14;
                                                    if ((-1 < (int)*plVar8) &&
                                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                      lVar7 = FUN_18000cc40(plVar6);
                                                      (**(code **)(lVar7 + 0x30))(plVar6);
                                                    }
                                                    plVar6 = local_398;
                                                    local_398 = plVar21;
                                                    iVar3 = (int)*plVar21 + 1;
                                                    if (iVar3 != 0) {
                                                      *(int *)plVar21 = iVar3;
                                                    }
                                                    if ((-1 < (int)*plVar16) &&
                                                       (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                      lVar7 = FUN_18000cc40(plVar6);
                                                      (**(code **)(lVar7 + 0x30))(plVar6);
                                                    }
                                                    goto LAB_1800bd6a6;
                                                  }
                                                  FUN_18000ce30(param_1,&local_3a8);
                                                  plVar16 = local_380;
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                  uVar18 = 0xfe;
                                                  lVar7 = local_3a8;
                                                  }
                                                  }
                                                  goto LAB_1800bcc4b;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  FUN_18000ce30(param_1,&local_3a8);
                                                  uVar19 = 0xfd;
                                                  local_368 = 0xfd;
                                                  goto LAB_1800ba041;
                                                  }
/* novels.pipeline.chapter_review_cycle#41="log" */
                                                  FUN_180370fe0(param_1,&local_3a8,DAT_1803b20a8);
                                                  FUN_18000d1d0(param_1,local_3a8);
                                                  uVar18 = 0xfd;
                                                  plVar16 = local_380;
                                                  plVar5 = local_388;
                                                  plVar21 = local_398;
                                                  plVar8 = local_3a0;
                                                  lVar7 = local_3a8;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                }
                                              }
                                            }
                                          }
                                        }
                                      }
                                    }
                                    else {
                                      lVar7 = local_328[2];
                                      plVar6 = (longlong *)FUN_180374a40(param_1,3);
                                      if (*piVar1 + 1 != 0) {
                                        *piVar1 = *piVar1 + 1;
                                      }
                                      FUN_18000cc80(plVar6,0,piVar1);
                                      plVar21 = (longlong *)
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                FUN_18037f0c0(param_1,local_228,DAT_1803b2110);
                                      if (plVar21 != (longlong *)0x0) {
/* novels.pipeline.chapter_review_cycle#59=".0f" */
                                        lVar10 = FUN_180375520(param_1,plVar21,DAT_1803b2138);
                                        if ((-1 < (int)*plVar21) &&
                                           (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                          lVar11 = FUN_18000cc40(plVar21);
                                          (**(code **)(lVar11 + 0x30))(plVar21);
                                        }
                                        if (lVar10 != 0) {
                                          FUN_18000cc80(plVar6,1,lVar10);
/* novels.pipeline.chapter_review_cycle#60=" 分）" */
                                          if (*DAT_1803b2140 + 1 != 0) {
/* novels.pipeline.chapter_review_cycle#60=" 分）" */
                                            *DAT_1803b2140 = *DAT_1803b2140 + 1;
                                          }
                                          FUN_18000cc80(plVar6,2);
/* novels.pipeline.chapter_review_cycle#18="" */
                                          plVar21 = (longlong *)PyUnicode_Join(DAT_1803b1ff0,plVar6)
                                          ;
                                          if ((-1 < (int)*plVar6) &&
                                             (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                            lVar10 = FUN_18000cc40(plVar6);
                                            (**(code **)(lVar10 + 0x30))(plVar6);
                                          }
                                          if (plVar21 == (longlong *)0x0) {
                                            FUN_18000ce30(param_1,&local_3a8);
                                            uVar18 = 0xc3;
                                            plVar16 = local_380;
                                            plVar5 = local_388;
                                            plVar21 = local_398;
                                            plVar8 = local_3a0;
                                            lVar7 = local_3a8;
                                          }
                                          else {
                                            uVar18 = 0xc3;
                                            *(undefined4 *)(local_388 + 5) = 0xc3;
                                            plVar16 = (longlong *)
                                                      FUN_180376770(param_1,lVar7,plVar21);
                                            if ((-1 < (int)*plVar21) &&
                                               (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                              lVar7 = FUN_18000cc40(plVar21);
                                              (**(code **)(lVar7 + 0x30))(plVar21);
                                            }
                                            if (plVar16 == (longlong *)0x0) {
                                              FUN_18000ce30(param_1,&local_3a8);
                                              plVar16 = local_380;
                                              plVar5 = local_388;
                                              plVar21 = local_398;
                                              plVar8 = local_3a0;
                                              lVar7 = local_3a8;
                                            }
                                            else {
                                              if ((-1 < (int)*plVar16) &&
                                                 (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                lVar7 = FUN_18000cc40(plVar16);
                                                (**(code **)(lVar7 + 0x30))(plVar16);
                                              }
                                              lVar7 = FUN_1800b4ae0(param_1);
                                              plVar21 = local_228;
                                              if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#52="ChapterReviewCycleResult" */
                                                FUN_180373340(param_1,&local_3a8,DAT_1803b2100);
                                                uVar18 = 0xc4;
                                                plVar16 = local_380;
                                                plVar5 = local_388;
                                                plVar21 = local_398;
                                                plVar8 = local_3a0;
                                                lVar7 = local_3a8;
                                              }
                                              else {
                                                plVar16 = (longlong *)
                                                          FUN_18037f0c0(param_1,local_228,
/* novels.pipeline.chapter_review_cycle#53="word_count" */
                                                                        DAT_1803b2108);
                                                if (plVar16 == (longlong *)0x0) {
                                                  FUN_18000ce30(param_1,&local_3a8);
                                                  plVar16 = local_380;
                                                  plVar5 = local_388;
                                                  plVar8 = local_3a0;
                                                  uVar18 = 0xc6;
                                                  lVar7 = local_3a8;
                                                }
                                                else {
                                                  pcVar20 = _Py_FalseStruct_exref;
                                                  plVar5 = (longlong *)
                                                           FUN_18037f0c0(param_1,plVar21,
/* novels.pipeline.chapter_review_cycle#49="audit" */
                                                                         DAT_1803b20e8);
                                                  if (plVar5 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_3a8);
                                                    if ((-1 < (int)*plVar16) &&
                                                       (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                      lVar7 = FUN_18000cc40(plVar16);
                                                      (**(code **)(lVar7 + 0x30))(plVar16);
                                                    }
                                                    plVar16 = local_380;
                                                    plVar5 = local_388;
                                                    plVar8 = local_3a0;
                                                    uVar18 = 200;
                                                    lVar7 = local_3a8;
                                                  }
                                                  else {
                                                    plVar6 = (longlong *)
                                                             FUN_18037f0c0(param_1,plVar21,
/* novels.pipeline.chapter_review_cycle#54="score" */
                                                                           DAT_1803b2110);
                                                    param_1 = local_390;
                                                    if (plVar6 == (longlong *)0x0) {
                                                      FUN_18000ce30(local_390,&local_3a8);
                                                      if ((-1 < (int)*plVar16) &&
                                                         (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                                                      {
                                                        lVar7 = FUN_18000cc40(plVar16);
                                                        (**(code **)(lVar7 + 0x30))(plVar16);
                                                      }
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        lVar7 = FUN_18000cc40(plVar5);
                                                        (**(code **)(lVar7 + 0x30))(plVar5);
                                                      }
                                                      plVar16 = local_380;
                                                      plVar5 = local_388;
                                                      plVar8 = local_3a0;
                                                      uVar18 = 0xc9;
                                                      lVar7 = local_3a8;
                                                    }
                                                    else {
                                                      plVar21 = (longlong *)
                                                                FUN_18037f0c0(local_390,plVar21,
/* novels.pipeline.chapter_review_cycle#55="length_in_range" */
                                                                              DAT_1803b2118);
                                                      if (plVar21 != (longlong *)0x0) {
                                                        pplVar17 = &local_d8;
                                                        local_a8 = local_318;
                                                        local_368 = 0xc4;
                                                        *(undefined4 *)(local_388 + 5) = 0xc4;
                                                        local_d8 = plVar8;
                                                        local_d0 = plVar16;
                                                        local_c8 = pcVar20;
                                                        local_c0 = plVar5;
                                                        local_b8 = plVar6;
                                                        local_b0 = plVar21;
                                                        goto LAB_1800b9cb6;
                                                      }
                                                      FUN_18000ce30(local_390,&local_3a8);
                                                      if ((-1 < (int)*plVar16) &&
                                                         (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                                                      {
                                                        lVar7 = FUN_18000cc40(plVar16);
                                                        (**(code **)(lVar7 + 0x30))(plVar16);
                                                      }
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        lVar7 = FUN_18000cc40(plVar5);
                                                        (**(code **)(lVar7 + 0x30))(plVar5);
                                                      }
                                                      if ((-1 < (int)*plVar6) &&
                                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                        lVar7 = FUN_18000cc40(plVar6);
                                                        (**(code **)(lVar7 + 0x30))(plVar6);
                                                      }
                                                      uVar18 = 0xca;
                                                      param_1 = local_390;
                                                      plVar16 = local_380;
                                                      plVar5 = local_388;
                                                      plVar21 = local_398;
                                                      plVar8 = local_3a0;
                                                      lVar7 = local_3a8;
                                                    }
                                                  }
                                                }
                                              }
                                            }
                                          }
                                          goto LAB_1800bcc4b;
                                        }
                                      }
                                      uVar19 = 0xc3;
                                      local_368 = 0xc3;
                                      FUN_18000ce30(param_1,&local_3a8);
LAB_1800ba041:
                                      plVar16 = local_380;
                                      plVar5 = local_388;
                                      plVar21 = local_398;
                                      plVar8 = local_3a0;
                                      uVar18 = local_368;
                                      lVar7 = local_3a8;
                                      if ((-1 < (int)*plVar6) &&
                                         (*plVar6 = *plVar6 + -1, plVar16 = local_380, *plVar6 == 0)
                                         ) {
                                        lVar7 = FUN_18000cc40(plVar6);
                                        (**(code **)(lVar7 + 0x30))(plVar6);
                                        plVar16 = local_380;
                                        plVar5 = local_388;
                                        plVar21 = local_398;
                                        plVar8 = local_3a0;
                                        uVar18 = uVar19;
                                        lVar7 = local_3a8;
                                      }
                                    }
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                      else {
                        lVar7 = local_328[2];
                        *(undefined4 *)(plVar5 + 5) = 0xb7;
                        plVar16 = (longlong *)FUN_1803770d0(param_1,lVar7,DAT_1803b20f8);
                        if (plVar16 == (longlong *)0x0) {
                          FUN_18000ce30(param_1,&local_3a8);
                          plVar16 = local_380;
                          plVar8 = local_3a0;
                          uVar18 = 0xb7;
                          lVar7 = local_3a8;
                        }
                        else {
                          if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                            lVar7 = FUN_18000cc40(plVar16);
                            (**(code **)(lVar7 + 0x30))(plVar16);
                          }
                          lVar7 = FUN_1800b4ae0(param_1);
                          if (lVar7 == 0) {
/* novels.pipeline.chapter_review_cycle#52="ChapterReviewCycleResult" */
                            FUN_180373340(param_1,&local_3a8,DAT_1803b2100);
                            plVar16 = local_380;
                            plVar5 = local_388;
                            plVar8 = local_3a0;
                            uVar18 = 0xb8;
                            lVar7 = local_3a8;
                          }
                          else {
/* novels.pipeline.chapter_review_cycle#53="word_count" */
                            plVar16 = (longlong *)FUN_18037f0c0(param_1,plVar21,DAT_1803b2108);
                            pcVar20 = _Py_FalseStruct_exref;
                            if (plVar16 == (longlong *)0x0) {
                              FUN_18000ce30(param_1,&local_3a8);
                              plVar16 = local_380;
                              plVar5 = local_388;
                              plVar8 = local_3a0;
                              uVar18 = 0xba;
                              lVar7 = local_3a8;
                            }
                            else {
/* novels.pipeline.chapter_review_cycle#49="audit" */
                              plVar5 = (longlong *)FUN_18037f0c0(param_1,plVar21,DAT_1803b20e8);
                              if (plVar5 == (longlong *)0x0) {
                                FUN_18000ce30(param_1,&local_3a8);
                                if ((-1 < (int)*plVar16) &&
                                   (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                  lVar7 = FUN_18000cc40(plVar16);
                                  (**(code **)(lVar7 + 0x30))(plVar16);
                                }
                                plVar16 = local_380;
                                plVar5 = local_388;
                                plVar8 = local_3a0;
                                uVar18 = 0xbc;
                                lVar7 = local_3a8;
                              }
                              else {
/* novels.pipeline.chapter_review_cycle#54="score" */
                                plVar6 = (longlong *)FUN_18037f0c0(param_1,plVar21,DAT_1803b2110);
                                param_1 = local_390;
                                if (plVar6 == (longlong *)0x0) {
                                  FUN_18000ce30(local_390,&local_3a8);
                                  if ((-1 < (int)*plVar16) &&
                                     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                    lVar7 = FUN_18000cc40(plVar16);
                                    (**(code **)(lVar7 + 0x30))(plVar16);
                                  }
                                  if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0))
                                  {
                                    lVar7 = FUN_18000cc40(plVar5);
                                    (**(code **)(lVar7 + 0x30))(plVar5);
                                  }
                                  plVar16 = local_380;
                                  plVar5 = local_388;
                                  plVar8 = local_3a0;
                                  uVar18 = 0xbd;
                                  lVar7 = local_3a8;
                                }
                                else {
                                  plVar21 = (longlong *)
/* novels.pipeline.chapter_review_cycle#55="length_in_range" */
                                            FUN_18037f0c0(local_390,plVar21,DAT_1803b2118);
                                  if (plVar21 == (longlong *)0x0) {
                                    FUN_18000ce30(local_390,&local_3a8);
                                    if ((-1 < (int)*plVar16) &&
                                       (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                      lVar7 = FUN_18000cc40(plVar16);
                                      (**(code **)(lVar7 + 0x30))(plVar16);
                                    }
                                    if ((-1 < (int)*plVar5) &&
                                       (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                      lVar7 = FUN_18000cc40(plVar5);
                                      (**(code **)(lVar7 + 0x30))(plVar5);
                                    }
                                    if ((-1 < (int)*plVar6) &&
                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                      lVar7 = FUN_18000cc40(plVar6);
                                      (**(code **)(lVar7 + 0x30))(plVar6);
                                    }
                                    param_1 = local_390;
                                    plVar16 = local_380;
                                    plVar5 = local_388;
                                    plVar21 = local_398;
                                    plVar8 = local_3a0;
                                    uVar18 = 0xbe;
                                    lVar7 = local_3a8;
                                  }
                                  else {
                                    pplVar17 = &local_110;
                                    local_368 = 0xb8;
                                    *(undefined4 *)(local_388 + 5) = 0xb8;
                                    local_100 = pcVar20;
                                    local_e0 = local_318;
                                    local_110 = plVar8;
                                    local_108 = plVar16;
                                    local_f8 = plVar5;
                                    local_f0 = plVar6;
                                    local_e8 = plVar21;
LAB_1800b9cb6:
                                    param_1 = local_390;
                                    lVar7 = FUN_18037ccb0(local_390,lVar7,pplVar17,DAT_1803b2120);
                                    if ((-1 < (int)*plVar16) &&
                                       (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                      lVar10 = FUN_18000cc40(plVar16);
                                      (**(code **)(lVar10 + 0x30))(plVar16);
                                    }
                                    if ((-1 < (int)*plVar5) &&
                                       (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                      lVar10 = FUN_18000cc40(plVar5);
                                      (**(code **)(lVar10 + 0x30))(plVar5);
                                    }
                                    if ((-1 < (int)*plVar6) &&
                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                      lVar10 = FUN_18000cc40(plVar6);
                                      (**(code **)(lVar10 + 0x30))(plVar6);
                                    }
                                    if ((-1 < (int)*plVar21) &&
                                       (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
                                      lVar10 = FUN_18000cc40(plVar21);
LAB_1800b9d3d:
                                      (**(code **)(lVar10 + 0x30))(plVar21);
                                    }
LAB_1800b9d43:
                                    if (lVar7 != 0) {
                                      FUN_180001d80(param_1);
                                      plVar21 = local_318;
                                      if (((local_318 != (longlong *)0x0) && (-1 < (int)*local_318))
                                         && (*local_318 = *local_318 + -1, *local_318 == 0)) {
                                        lVar10 = FUN_18000cc40(local_318);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_380;
                                      if (((local_380 != (longlong *)0x0) && (-1 < (int)*local_380))
                                         && (*local_380 = *local_380 + -1, *local_380 == 0)) {
                                        lVar10 = FUN_18000cc40(local_380);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_328;
                                      if ((-1 < (int)*local_328) &&
                                         (*local_328 = *local_328 + -1, *local_328 == 0)) {
                                        lVar10 = FUN_18000cc40(local_328);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_3a0;
                                      if (((local_3a0 != (longlong *)0x0) && (-1 < (int)*local_3a0))
                                         && (*local_3a0 = *local_3a0 + -1, *local_3a0 == 0)) {
                                        lVar10 = FUN_18000cc40(local_3a0);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      if ((-1 < (int)*local_370) &&
                                         (*local_370 = *local_370 + -1, *local_370 == 0)) {
                                        lVar10 = FUN_18000cc40(local_2a8);
                                        (**(code **)(lVar10 + 0x30))(local_2a8);
                                      }
                                      plVar21 = local_298;
                                      if (((local_298 != (longlong *)0x0) && (-1 < (int)*local_298))
                                         && (*local_298 = *local_298 + -1, *local_298 == 0)) {
                                        lVar10 = FUN_18000cc40(local_298);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      if ((-1 < (int)*local_228) &&
                                         (*local_228 = *local_228 + -1, *local_228 == 0)) {
                                        lVar10 = FUN_18000cc40(local_268);
                                        (**(code **)(lVar10 + 0x30))(local_268);
                                      }
                                      plVar21 = local_398;
                                      if (((local_398 != (longlong *)0x0) && (-1 < (int)*local_398))
                                         && (*local_398 = *local_398 + -1, *local_398 == 0)) {
                                        lVar10 = FUN_18000cc40(local_398);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_310;
                                      if (((local_310 != (longlong *)0x0) && (-1 < (int)*local_310))
                                         && (*local_310 = *local_310 + -1, *local_310 == 0)) {
                                        lVar10 = FUN_18000cc40(local_310);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_300;
                                      if (((local_300 != (longlong *)0x0) && (-1 < (int)*local_300))
                                         && (*local_300 = *local_300 + -1, *local_300 == 0)) {
                                        lVar10 = FUN_18000cc40(local_300);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      pcVar20 = local_2a0;
                                      if (((local_2a0 != (code *)0x0) && (-1 < *(int *)local_2a0))
                                         && (*(longlong *)local_2a0 = *(longlong *)local_2a0 + -1,
                                            *(longlong *)local_2a0 == 0)) {
                                        lVar10 = FUN_18000cc40(local_2a0);
                                        (**(code **)(lVar10 + 0x30))(pcVar20);
                                      }
                                      plVar21 = local_2d0;
                                      if (((local_2d0 != (longlong *)0x0) && (-1 < (int)*local_2d0))
                                         && (*local_2d0 = *local_2d0 + -1, *local_2d0 == 0)) {
                                        lVar10 = FUN_18000cc40(local_2d0);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_2c8;
                                      if (((local_2c8 != (longlong *)0x0) && (-1 < (int)*local_2c8))
                                         && (*local_2c8 = *local_2c8 + -1, *local_2c8 == 0)) {
                                        lVar10 = FUN_18000cc40(local_2c8);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_2c0;
                                      if (((local_2c0 != (longlong *)0x0) && (-1 < (int)*local_2c0))
                                         && (*local_2c0 = *local_2c0 + -1, *local_2c0 == 0)) {
                                        lVar10 = FUN_18000cc40(local_2c0);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_2b8;
                                      if (((local_2b8 != (longlong *)0x0) && (-1 < (int)*local_2b8))
                                         && (*local_2b8 = *local_2b8 + -1, *local_2b8 == 0)) {
                                        lVar10 = FUN_18000cc40(local_2b8);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_2f8;
                                      if (((local_2f8 != (longlong *)0x0) && (-1 < (int)*local_2f8))
                                         && (*local_2f8 = *local_2f8 + -1, *local_2f8 == 0)) {
                                        lVar10 = FUN_18000cc40(local_2f8);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_2b0;
                                      if (((local_2b0 != (longlong *)0x0) && (-1 < (int)*local_2b0))
                                         && (*local_2b0 = *local_2b0 + -1, *local_2b0 == 0)) {
                                        lVar10 = FUN_18000cc40(local_2b0);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_330;
                                      if (((local_330 != (longlong *)0x0) && (-1 < (int)*local_330))
                                         && (*local_330 = *local_330 + -1, *local_330 == 0)) {
                                        lVar10 = FUN_18000cc40(local_330);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_2f0;
                                      if (((local_2f0 != (longlong *)0x0) && (-1 < (int)*local_2f0))
                                         && (*local_2f0 = *local_2f0 + -1, *local_2f0 == 0)) {
                                        lVar10 = FUN_18000cc40(local_2f0);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      pcVar20 = local_2e0;
                                      if (((local_2e0 != (code *)0x0) && (-1 < *(int *)local_2e0))
                                         && (*(longlong *)local_2e0 = *(longlong *)local_2e0 + -1,
                                            *(longlong *)local_2e0 == 0)) {
                                        lVar10 = FUN_18000cc40(local_2e0);
                                        (**(code **)(lVar10 + 0x30))(pcVar20);
                                      }
                                      plVar21 = local_290;
                                      if ((-1 < (int)*local_290) &&
                                         (*local_290 = *local_290 + -1, *local_290 == 0)) {
                                        lVar10 = FUN_18000cc40(local_290);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_220;
                                      if ((-1 < (int)*local_220) &&
                                         (*local_220 = *local_220 + -1, *local_220 == 0)) {
                                        lVar10 = FUN_18000cc40(local_220);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_260;
                                      if ((-1 < (int)*local_260) &&
                                         (*local_260 = *local_260 + -1, *local_260 == 0)) {
                                        lVar10 = FUN_18000cc40(local_260);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_258;
                                      if ((-1 < (int)*local_258) &&
                                         (*local_258 = *local_258 + -1, *local_258 == 0)) {
                                        lVar10 = FUN_18000cc40(local_258);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_2e8;
                                      if ((-1 < (int)*local_2e8) &&
                                         (*local_2e8 = *local_2e8 + -1, *local_2e8 == 0)) {
                                        lVar10 = FUN_18000cc40(local_2e8);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_250;
                                      if ((-1 < (int)*local_250) &&
                                         (*local_250 = *local_250 + -1, *local_250 == 0)) {
                                        lVar10 = FUN_18000cc40(local_250);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_2d8;
                                      if ((-1 < (int)*local_2d8) &&
                                         (*local_2d8 = *local_2d8 + -1, *local_2d8 == 0)) {
                                        lVar10 = FUN_18000cc40(local_2d8);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_218;
                                      if ((-1 < (int)*local_218) &&
                                         (*local_218 = *local_218 + -1, *local_218 == 0)) {
                                        lVar10 = FUN_18000cc40(local_218);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_288;
                                      if ((-1 < (int)*local_288) &&
                                         (*local_288 = *local_288 + -1, *local_288 == 0)) {
                                        lVar10 = FUN_18000cc40(local_288);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_240;
                                      if ((-1 < (int)*local_240) &&
                                         (*local_240 = *local_240 + -1, *local_240 == 0)) {
                                        lVar10 = FUN_18000cc40(local_240);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_238;
                                      if ((-1 < (int)*local_238) &&
                                         (*local_238 = *local_238 + -1, *local_238 == 0)) {
                                        lVar10 = FUN_18000cc40(local_238);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_230;
                                      if ((-1 < (int)*local_230) &&
                                         (*local_230 = *local_230 + -1, *local_230 == 0)) {
                                        lVar10 = FUN_18000cc40(local_230);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_280;
                                      if ((-1 < (int)*local_280) &&
                                         (*local_280 = *local_280 + -1, *local_280 == 0)) {
                                        lVar10 = FUN_18000cc40(local_280);
                                        (**(code **)(lVar10 + 0x30))(plVar21);
                                      }
                                      plVar21 = local_278;
                                      if ((int)*local_278 < 0) {
                                        return lVar7;
                                      }
                                      *local_278 = *local_278 + -1;
                                      if (*local_278 != 0) {
                                        return lVar7;
                                      }
                                      lVar10 = FUN_18000cc40(local_278);
                                      (**(code **)(lVar10 + 0x30))(plVar21);
                                      return lVar7;
                                    }
                                    FUN_18000ce30(param_1,&local_3a8);
                                    plVar16 = local_380;
                                    plVar5 = local_388;
                                    plVar21 = local_398;
                                    plVar8 = local_3a0;
                                    uVar18 = local_368;
                                    lVar7 = local_3a8;
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                      goto LAB_1800bcc4b;
                    }
                    FUN_18000ce30(param_1,&local_3a8);
                    if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                      lVar7 = FUN_18000cc40(plVar16);
                      (**(code **)(lVar7 + 0x30))(plVar16);
                      plVar16 = local_380;
                      plVar8 = local_3a0;
                      lVar7 = local_3a8;
                      goto LAB_1800bcc4b;
                    }
                  }
                  plVar16 = local_380;
                  plVar8 = local_3a0;
                  uVar18 = 0xb6;
                  lVar7 = local_3a8;
                }
              }
            }
          }
        }
      }
    }
  }
LAB_1800bcc4b:
  local_3a8 = lVar7;
  piVar1 = *(int **)(local_3a8 + 0x28);
  if (piVar1 == (int *)0x0) {
    lVar7 = FUN_18038e2a0(plVar5,uVar18);
  }
  else {
    if (*(longlong **)(piVar1 + 6) == plVar5) goto LAB_1800be762;
    lVar7 = FUN_18038e2a0(plVar5,uVar18);
    *(int **)(lVar7 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
  }
  FUN_180001300(&local_3a8,lVar7);
LAB_1800be762:
  plVar14 = local_330;
  local_3b0 = local_2e0;
  local_3b8 = local_2f0;
  local_3c0 = local_330;
  local_3c8 = local_2b0;
  local_3d0 = local_2f8;
  local_3d8 = local_2b8;
  local_3e0 = local_2c0;
  local_3e8 = local_2c8;
  local_3f0 = local_2d0;
  local_3f8 = local_2a0;
  local_400 = local_300;
  local_408 = local_310;
  local_418 = local_268;
  local_420 = local_298;
  local_428 = local_2a8;
  local_438 = local_328;
  local_448 = local_318;
  local_450 = local_278;
  local_458 = local_280;
  local_460 = local_230;
  local_468 = local_238;
  local_470 = local_240;
  local_478 = local_288;
  local_480 = (longlong **)local_218;
  local_488 = local_2d8;
  local_490 = local_250;
  local_498 = local_2e8;
  local_4a0 = local_258;
  local_4a8 = local_260;
  local_440 = plVar16;
  local_430 = plVar8;
  local_410 = plVar21;
  FUN_18036d0a0(plVar5,"ooccooococccoooocooooooooooooooooo",local_290,local_220);
  plVar6 = DAT_1803b1f08;
  if (plVar5 == DAT_1803b1f08) {
    if ((-1 < (int)*DAT_1803b1f08) && (*DAT_1803b1f08 = *DAT_1803b1f08 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))();
    }
    DAT_1803b1f08 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  lVar7 = local_3a8;
  local_3a8 = 0;
  if (((local_318 != (longlong *)0x0) && (-1 < (int)*local_318)) &&
     (*local_318 = *local_318 + -1, *local_318 == 0)) {
    (**(code **)(local_318[1] + 0x30))();
  }
  if (((plVar16 != (longlong *)0x0) && (-1 < (int)*plVar16)) &&
     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
    (**(code **)(plVar16[1] + 0x30))(plVar16);
  }
  if ((-1 < (int)*local_328) && (*local_328 = *local_328 + -1, *local_328 == 0)) {
    (**(code **)(local_328[1] + 0x30))();
  }
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))(plVar8);
  }
  if (((local_2a8 != (longlong *)0x0) && (-1 < (int)*local_2a8)) &&
     (*local_2a8 = *local_2a8 + -1, *local_2a8 == 0)) {
    (**(code **)(local_2a8[1] + 0x30))();
  }
  if (((local_298 != (longlong *)0x0) && (-1 < (int)*local_298)) &&
     (*local_298 = *local_298 + -1, *local_298 == 0)) {
    (**(code **)(local_298[1] + 0x30))();
  }
  if (((local_268 != (longlong *)0x0) && (-1 < (int)*local_268)) &&
     (*local_268 = *local_268 + -1, *local_268 == 0)) {
    (**(code **)(local_268[1] + 0x30))();
  }
  if (((plVar21 != (longlong *)0x0) && (-1 < (int)*plVar21)) &&
     (*plVar21 = *plVar21 + -1, *plVar21 == 0)) {
    (**(code **)(plVar21[1] + 0x30))(plVar21);
  }
  if (((local_310 != (longlong *)0x0) && (-1 < (int)*local_310)) &&
     (*local_310 = *local_310 + -1, *local_310 == 0)) {
    (**(code **)(local_310[1] + 0x30))();
  }
  if (((local_300 != (longlong *)0x0) && (-1 < (int)*local_300)) &&
     (*local_300 = *local_300 + -1, *local_300 == 0)) {
    (**(code **)(local_300[1] + 0x30))();
  }
  if (((local_2a0 != (code *)0x0) && (-1 < *(int *)local_2a0)) &&
     (*(longlong *)local_2a0 = *(longlong *)local_2a0 + -1, *(longlong *)local_2a0 == 0)) {
    (**(code **)(*(longlong *)(local_2a0 + 8) + 0x30))();
  }
  if (((local_2d0 != (longlong *)0x0) && (-1 < (int)*local_2d0)) &&
     (*local_2d0 = *local_2d0 + -1, *local_2d0 == 0)) {
    (**(code **)(local_2d0[1] + 0x30))();
  }
  if (((local_2c8 != (longlong *)0x0) && (-1 < (int)*local_2c8)) &&
     (*local_2c8 = *local_2c8 + -1, *local_2c8 == 0)) {
    (**(code **)(local_2c8[1] + 0x30))();
  }
  if (((local_2c0 != (longlong *)0x0) && (-1 < (int)*local_2c0)) &&
     (*local_2c0 = *local_2c0 + -1, *local_2c0 == 0)) {
    (**(code **)(local_2c0[1] + 0x30))();
  }
  if (((local_2b8 != (longlong *)0x0) && (-1 < (int)*local_2b8)) &&
     (*local_2b8 = *local_2b8 + -1, *local_2b8 == 0)) {
    (**(code **)(local_2b8[1] + 0x30))();
  }
  if (((local_2f8 != (longlong *)0x0) && (-1 < (int)*local_2f8)) &&
     (*local_2f8 = *local_2f8 + -1, *local_2f8 == 0)) {
    (**(code **)(local_2f8[1] + 0x30))();
  }
  if (((local_2b0 != (longlong *)0x0) && (-1 < (int)*local_2b0)) &&
     (*local_2b0 = *local_2b0 + -1, *local_2b0 == 0)) {
    (**(code **)(local_2b0[1] + 0x30))();
  }
  if (((plVar14 != (longlong *)0x0) && (-1 < (int)*plVar14)) &&
     (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
    (**(code **)(plVar14[1] + 0x30))(plVar14);
  }
  if (((local_2f0 != (longlong *)0x0) && (-1 < (int)*local_2f0)) &&
     (*local_2f0 = *local_2f0 + -1, *local_2f0 == 0)) {
    (**(code **)(local_2f0[1] + 0x30))();
  }
  if (((local_2e0 != (code *)0x0) && (-1 < *(int *)local_2e0)) &&
     (*(longlong *)local_2e0 = *(longlong *)local_2e0 + -1, *(longlong *)local_2e0 == 0)) {
    (**(code **)(*(longlong *)(local_2e0 + 8) + 0x30))();
  }
  local_3a8 = lVar7;
  if ((-1 < (int)*local_290) && (*local_290 = *local_290 + -1, *local_290 == 0)) {
    (**(code **)(local_290[1] + 0x30))();
  }
  if ((-1 < (int)*local_220) && (*local_220 = *local_220 + -1, *local_220 == 0)) {
    (**(code **)(local_220[1] + 0x30))();
  }
  if ((-1 < (int)*local_260) && (*local_260 = *local_260 + -1, *local_260 == 0)) {
    (**(code **)(local_260[1] + 0x30))();
  }
  if ((-1 < (int)*local_258) && (*local_258 = *local_258 + -1, *local_258 == 0)) {
    (**(code **)(local_258[1] + 0x30))();
  }
  if ((-1 < (int)*local_2e8) && (*local_2e8 = *local_2e8 + -1, *local_2e8 == 0)) {
    (**(code **)(local_2e8[1] + 0x30))();
  }
  if ((-1 < (int)*local_250) && (*local_250 = *local_250 + -1, *local_250 == 0)) {
    (**(code **)(local_250[1] + 0x30))();
  }
  if ((-1 < (int)*local_2d8) && (*local_2d8 = *local_2d8 + -1, *local_2d8 == 0)) {
    (**(code **)(local_2d8[1] + 0x30))();
  }
  if ((-1 < (int)*local_218) && (*local_218 = *local_218 + -1, *local_218 == 0)) {
    (**(code **)(local_218[1] + 0x30))();
  }
  if ((-1 < (int)*local_288) && (*local_288 = *local_288 + -1, *local_288 == 0)) {
    (**(code **)(local_288[1] + 0x30))();
  }
  if ((-1 < (int)*local_240) && (*local_240 = *local_240 + -1, *local_240 == 0)) {
    (**(code **)(local_240[1] + 0x30))();
  }
  if ((-1 < (int)*local_238) && (*local_238 = *local_238 + -1, *local_238 == 0)) {
    (**(code **)(local_238[1] + 0x30))();
  }
  if ((-1 < (int)*local_230) && (*local_230 = *local_230 + -1, *local_230 == 0)) {
    (**(code **)(local_230[1] + 0x30))();
  }
  if ((-1 < (int)*local_280) && (*local_280 = *local_280 + -1, *local_280 == 0)) {
    (**(code **)(local_280[1] + 0x30))();
  }
  if ((-1 < (int)*local_278) && (*local_278 = *local_278 + -1, *local_278 == 0)) {
    (**(code **)(local_278[1] + 0x30))();
  }
  FUN_18000ce40(param_1,&local_3a8);
  return 0;
}
