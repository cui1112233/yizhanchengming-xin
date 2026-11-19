/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x18004a310 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

longlong FUN_18004a310(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  int *piVar1;
  undefined8 uVar2;
  code *pcVar3;
  char cVar4;
  int iVar5;
  uint uVar6;
  int iVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong lVar11;
  longlong *plVar12;
  longlong *plVar13;
  undefined8 uVar14;
  longlong lVar15;
  undefined8 uVar16;
  code *pcVar17;
  code *pcVar18;
  code *pcVar19;
  bool bVar20;
  undefined1 auStack_4d8 [32];
  longlong *local_4b8;
  longlong *local_4b0;
  longlong *local_4a8;
  longlong *local_4a0;
  longlong *local_498;
  code *local_490;
  code *local_488;
  longlong *local_480;
  longlong *local_478;
  code *local_470;
  code *local_468;
  longlong *local_460;
  longlong *local_458;
  longlong *local_450;
  code *local_448;
  longlong *local_440;
  code *local_438;
  code *local_430;
  code *local_428;
  code *local_420;
  code *local_418;
  code *local_410;
  code *local_408;
  code *local_400;
  code *local_3f8;
  code *local_3f0;
  longlong *local_3e8;
  longlong *local_3e0;
  longlong *local_3d8;
  longlong *local_3d0;
  longlong *local_3c8;
  longlong *local_3c0;
  longlong *local_3b8;
  longlong *local_3b0;
  longlong *local_3a8;
  longlong *local_3a0;
  code *local_398;
  longlong local_388;
  code *local_380;
  code *local_378;
  code *local_370;
  code *local_368;
  longlong *local_360;
  longlong *local_358;
  code *local_350;
  code *local_348;
  longlong *local_340;
  code *local_338;
  longlong *local_330;
  longlong *local_328;
  longlong *local_320;
  code *local_318;
  code *local_310;
  code *local_308;
  code *local_300;
  code *local_2f8;
  code *local_2f0;
  code *local_2e8;
  code *local_2e0;
  code *local_2d8;
  code *local_2d0;
  code *local_2c8;
  code *local_2c0;
  longlong *local_2b8;
  code *local_2b0;
  longlong *local_2a8;
  longlong *local_2a0;
  longlong *local_298;
  code *local_290;
  code *local_288;
  code *local_280;
  code *local_278;
  code *local_270;
  code *local_268;
  code *local_260;
  code *local_258;
  longlong *local_250;
  longlong *local_248;
  longlong *local_240;
  longlong *local_238;
  code *local_230;
  longlong *local_228;
  code *local_220;
  code *local_218;
  longlong *local_210;
  longlong *local_208;
  longlong *local_200;
  longlong *local_1f8;
  longlong *local_1f0;
  longlong *local_1e8;
  longlong *local_1e0;
  longlong *local_1d8;
  longlong *local_1d0;
  longlong *local_1c8;
  longlong *local_1c0;
  longlong *local_1b8;
  longlong *local_1b0;
  code *local_1a8;
  code *local_1a0;
  code *local_198;
  code *local_190;
  longlong *local_188;
  code *local_180;
  longlong *local_178;
  longlong *local_170;
  longlong *local_168;
  code *local_160;
  longlong *local_158;
  code *local_150;
  longlong *local_148;
  code *local_140;
  longlong *local_138;
  code *local_130;
  code *local_128;
  code *local_120;
  longlong *local_118;
  longlong *local_110;
  longlong *local_108;
  code *local_100;
  code *local_f8;
  code *local_f0;
  longlong *local_e8;
  code *local_e0;
  code *local_d8;
  code *local_d0;
  longlong *local_c8;
  longlong *local_b8;
  longlong *local_b0;
  code *local_a8;
  code *local_a0;
  code *local_98;
  code *local_90;
  code *local_88;
  code *local_80;
  code *local_78;
  code *local_70;
  code *local_68;
  code *local_60;
  code *local_58;
  ulonglong local_48;
  
  plVar12 = DAT_1803ae720;
  local_48 = DAT_1803a6040 ^ (ulonglong)auStack_4d8;
  local_330 = (longlong *)*param_3;
  pcVar17 = (code *)0x0;
  plVar8 = (longlong *)param_3[3];
  pcVar19 = (code *)0x0;
  local_2b0 = (code *)param_3[1];
  pcVar18 = (code *)0x0;
  local_2a8 = (longlong *)param_3[2];
  local_328 = (longlong *)param_3[4];
  local_358 = (longlong *)param_3[5];
  local_228 = (longlong *)param_3[6];
  local_220 = (code *)param_3[7];
  local_218 = (code *)param_3[8];
  local_388 = DAT_180391c90;
  local_208 = (longlong *)0x0;
  local_250 = (longlong *)0x0;
  local_378 = (code *)0x0;
  local_380 = (code *)0x0;
  local_1c8 = (longlong *)0x0;
  local_2b8 = (longlong *)0x0;
  local_320 = (longlong *)0x0;
  local_370 = (code *)0x0;
  local_200 = (longlong *)0x0;
  local_318 = (code *)0x0;
  local_310 = (code *)0x0;
  local_308 = (code *)0x0;
  local_300 = (code *)0x0;
  local_2f8 = (code *)0x0;
  local_2f0 = (code *)0x0;
  local_2e8 = (code *)0x0;
  local_2e0 = (code *)0x0;
  local_2d8 = (code *)0x0;
  local_2d0 = (code *)0x0;
  local_1f8 = (longlong *)0x0;
  local_1f0 = (longlong *)0x0;
  local_1e8 = (longlong *)0x0;
  local_1e0 = (longlong *)0x0;
  local_1d8 = (longlong *)0x0;
  local_1d0 = (longlong *)0x0;
  local_210 = (longlong *)0x0;
  local_248 = (longlong *)0x0;
  local_240 = (longlong *)0x0;
  local_238 = (longlong *)0x0;
  local_230 = (code *)0x0;
  local_360 = DAT_1803ae720;
  local_1b8 = plVar8;
  if (DAT_1803ae720 == (longlong *)0x0) {
LAB_18004a48c:
    DAT_1803ae720 = (longlong *)FUN_18036cd10(param_1,DAT_1803ae778,DAT_1803ca880,0x138);
  }
  else if ((1 < *DAT_1803ae720) || (DAT_1803ae720[2] != 0)) {
    if ((DAT_1803ae720 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803ae720 && (*DAT_1803ae720 = *DAT_1803ae720 + -1, *plVar12 == 0)))) {
      (**(code **)(plVar12[1] + 0x30))(plVar12);
    }
    goto LAB_18004a48c;
  }
  plVar12 = DAT_1803ae720;
  local_360 = DAT_1803ae720;
  FUN_180001d00(param_1,DAT_1803ae720);
  iVar5 = FUN_180001510(plVar8);
  if (iVar5 == -1) {
LAB_18004a4ce:
    local_388 = *(longlong *)(param_1 + 0x60);
    uVar16 = 0x29;
    *(undefined8 *)(param_1 + 0x60) = 0;
    pcVar18 = pcVar17;
    pcVar19 = pcVar17;
  }
  else {
    plVar9 = DAT_1803ae7a8;
    if (iVar5 == 1) {
      plVar9 = plVar8;
    }
    plVar8 = (longlong *)FUN_1803749a0(plVar9);
    if (plVar8 == (longlong *)0x0) goto LAB_18004a4ce;
    plVar9 = (longlong *)FUN_18037f0c0(param_1,plVar8,DAT_1803ae7b0);
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    uVar16 = 0x29;
    if (plVar9 == (longlong *)0x0) {
      local_388 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
      pcVar17 = (code *)0x0;
    }
    else {
      *(undefined4 *)(plVar12 + 5) = 0x29;
      plVar8 = (longlong *)FUN_180376210(param_1,plVar9);
      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
        (**(code **)(plVar9[1] + 0x30))(plVar9);
      }
      if (plVar8 == (longlong *)0x0) {
        local_388 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        pcVar17 = (code *)0x0;
      }
      else {
        plVar9 = (longlong *)FUN_18037f0c0(param_1,plVar8,DAT_1803ae7b8);
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        if (plVar9 == (longlong *)0x0) {
          local_388 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          pcVar17 = (code *)0x0;
        }
        else {
          *(undefined4 *)(plVar12 + 5) = 0x29;
          plVar8 = (longlong *)FUN_180376210(param_1,plVar9);
          local_1b0 = plVar8;
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          if (plVar8 == (longlong *)0x0) goto LAB_18004a605;
          local_208 = plVar8;
          plVar9 = (longlong *)FUN_18037f0c0(param_1,local_330,DAT_1803ae7a0);
          if (plVar9 == (longlong *)0x0) {
            local_388 = *(longlong *)(param_1 + 0x60);
            uVar16 = 0x2a;
            *(undefined8 *)(param_1 + 0x60) = 0;
            pcVar17 = (code *)0x0;
            pcVar19 = (code *)0x0;
            goto LAB_18004cfd4;
          }
          plVar10 = (longlong *)FUN_18037f0c0(param_1,plVar9,DAT_1803ae7c0);
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          uVar16 = 0x2a;
          if (plVar10 != (longlong *)0x0) {
            *(undefined4 *)(plVar12 + 5) = 0x2a;
            plVar9 = (longlong *)FUN_18037ce10(param_1,plVar10,DAT_1803ae7c8 + 0x18,DAT_1803ae7d0);
            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
              (**(code **)(plVar10[1] + 0x30))(plVar10);
            }
            if (plVar9 != (longlong *)0x0) {
              if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                (**(code **)(plVar9[1] + 0x30))(plVar9);
              }
              lVar15 = FUN_180049f90(param_1);
              plVar9 = local_328;
              if (lVar15 == 0) {
                FUN_180373340(param_1,&local_388,DAT_1803ae7d8);
                uVar16 = 0x2c;
                pcVar17 = (code *)0x0;
                pcVar19 = (code *)0x0;
                goto LAB_18004cfd4;
              }
              iVar5 = FUN_180001510(local_328);
              plVar10 = DAT_1803ae7e8;
              if (iVar5 == -1) {
                uVar16 = 0x2c;
                goto LAB_18004a605;
              }
              if (iVar5 == 0) {
                iVar5 = (int)*DAT_1803ae7e8 + 1;
                if (iVar5 != 0) {
                  *(int *)DAT_1803ae7e8 = iVar5;
                }
              }
              else {
                plVar10 = (longlong *)FUN_18037f0c0(param_1,plVar9,DAT_1803ae7e0);
                if (plVar10 == (longlong *)0x0) {
                  uVar16 = 0x2c;
                  goto LAB_18004a605;
                }
              }
              uVar16 = 0x2c;
              *(undefined4 *)(plVar12 + 5) = 0x2c;
              plVar9 = (longlong *)FUN_180376770(param_1,lVar15,plVar10);
              local_1c0 = plVar9;
              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                (**(code **)(plVar10[1] + 0x30))(plVar10);
              }
              plVar10 = local_328;
              if (plVar9 == (longlong *)0x0) {
                local_388 = *(longlong *)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                pcVar17 = local_378;
                pcVar18 = local_378;
                pcVar19 = local_378;
              }
              else {
                local_250 = plVar9;
                iVar5 = FUN_180001510(local_328);
                if (iVar5 == -1) {
                  local_388 = *(longlong *)(param_1 + 0x60);
                  uVar16 = 0x2d;
                  *(undefined8 *)(param_1 + 0x60) = 0;
                  pcVar17 = local_378;
                  pcVar18 = local_378;
                  pcVar19 = local_378;
                }
                else {
                  pcVar17 = (code *)0x0;
                  uVar6 = (uint)(iVar5 != 0);
                  if (iVar5 == 0) {
LAB_18004a90a:
                    pcVar18 = local_220;
                    pcVar17 = DAT_1803ae7f8;
                    if (uVar6 == 1) {
                      pcVar17 = DAT_1803ae800;
                    }
                    if (*(int *)pcVar17 + 1 != 0) {
                      *(int *)pcVar17 = *(int *)pcVar17 + 1;
                    }
                    local_378 = pcVar17;
                    iVar5 = FUN_180001510(local_220);
                    if (iVar5 == -1) {
                      local_388 = *(longlong *)(param_1 + 0x60);
                      uVar16 = 0x2e;
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      pcVar18 = local_380;
                      pcVar19 = local_380;
                    }
                    else {
                      if (iVar5 == 1) {
                        if (*(int *)pcVar18 + 1 != 0) {
                          *(int *)pcVar18 = *(int *)pcVar18 + 1;
                        }
                      }
                      else {
                        uVar16 = 0x2e;
                        *(undefined4 *)(plVar12 + 5) = 0x2e;
                        pcVar18 = (code *)FUN_18037e6c0(param_1,local_330,DAT_1803ae808,local_2a8);
                        if (pcVar18 == (code *)0x0) {
                          local_388 = *(longlong *)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          pcVar18 = local_380;
                          pcVar19 = local_380;
                          goto LAB_18004cfd4;
                        }
                      }
                      local_380 = pcVar18;
                      iVar5 = FUN_1803836d0(plVar8,DAT_1803ae7a8);
                      if (iVar5 == -1) {
                        local_388 = *(longlong *)(param_1 + 0x60);
                        uVar16 = 0x2f;
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        pcVar18 = local_370;
                        pcVar19 = local_380;
                      }
                      else if (iVar5 == 1) {
                        lVar15 = FUN_180049e80(param_1);
                        if (lVar15 == 0) {
                          FUN_180373340(param_1,&local_388,DAT_1803ae810);
                          uVar16 = 0x2f;
                          pcVar18 = local_370;
                          pcVar19 = local_380;
                        }
                        else {
                          uVar16 = 0x2f;
                          *(undefined4 *)(plVar12 + 5) = 0x2f;
                          local_2a0 = (longlong *)FUN_180376770(param_1,lVar15,pcVar18);
                          if (local_2a0 != (longlong *)0x0) goto LAB_18004aa5b;
                          local_388 = *(longlong *)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          pcVar18 = local_370;
                          pcVar19 = local_380;
                        }
                      }
                      else {
                        local_2a0 = DAT_1803ae818;
                        iVar5 = (int)*DAT_1803ae818 + 1;
                        if (iVar5 != 0) {
                          *(int *)DAT_1803ae818 = iVar5;
                        }
LAB_18004aa5b:
                        local_1c8 = local_2a0;
                        iVar5 = FUN_1803836d0(plVar8,DAT_1803ae7a8);
                        plVar8 = local_2a8;
                        if (iVar5 == -1) {
                          local_388 = *(longlong *)(param_1 + 0x60);
                          uVar16 = 0x31;
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          pcVar18 = local_370;
                          pcVar19 = local_380;
                        }
                        else if (iVar5 == 1) {
                          lVar15 = FUN_180049b50();
                          if (lVar15 == 0) {
                            FUN_180373340(param_1,&local_388,DAT_1803ae820);
                            uVar16 = 0x32;
                            pcVar18 = local_370;
                            pcVar19 = local_380;
                          }
                          else {
                            cVar4 = thunk_FUN_180383620(pcVar17,DAT_1803ae800);
                            pcVar19 = local_380;
                            uVar16 = 0x32;
                            local_1a8 = local_380;
                            local_1a0 = _Py_FalseStruct_exref;
                            if (cVar4 != '\0') {
                              local_1a0 = _Py_TrueStruct_exref;
                            }
                            *(undefined4 *)(plVar12 + 5) = 0x32;
                            plVar8 = (longlong *)FUN_180377750(param_1,lVar15,&local_1a8);
                            if (plVar8 == (longlong *)0x0) {
                              FUN_18000ce30(param_1,&local_388);
                              pcVar18 = local_370;
                            }
                            else {
                              local_2b8 = plVar8;
                              lVar15 = FUN_180049600(param_1);
                              if (lVar15 == 0) {
                                FUN_180373340(param_1,&local_388,DAT_1803ae828);
                                uVar16 = 0x33;
                                pcVar18 = local_370;
                                plVar12 = local_360;
                              }
                              else {
                                lVar11 = FUN_180049930(param_1);
                                if (lVar11 == 0) {
                                  FUN_180373340(param_1,&local_388,DAT_1803ae830);
                                  uVar16 = 0x35;
                                  pcVar18 = local_370;
                                  plVar12 = local_360;
                                }
                                else {
                                  uVar16 = 0x35;
                                  local_128 = _Py_NoneStruct_exref;
                                  *(undefined4 *)(local_360 + 5) = 0x35;
                                  local_120 = pcVar17;
                                  plVar8 = (longlong *)FUN_180377750(param_1,lVar11,&local_128);
                                  if (plVar8 == (longlong *)0x0) {
                                    FUN_18000ce30(param_1,&local_388);
                                    pcVar18 = local_370;
                                    plVar12 = local_360;
                                  }
                                  else {
                                    plVar12 = (longlong *)
                                              FUN_180001b80(param_1,plVar9,DAT_1803ae838);
                                    if (plVar12 == (longlong *)0x0) {
                                      FUN_18000ce30(param_1,&local_388);
                                      if ((-1 < (int)*plVar8) &&
                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                        lVar15 = FUN_18000cc40(plVar8);
                                        (**(code **)(lVar15 + 0x30))(plVar8);
                                      }
                                      uVar16 = 0x36;
                                      pcVar18 = local_370;
                                      plVar12 = local_360;
                                    }
                                    else {
                                      iVar5 = FUN_180001510(plVar12);
                                      if (iVar5 == -1) {
                                        FUN_18000ce30(param_1,&local_388);
                                        if ((-1 < (int)*plVar8) &&
                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                          lVar15 = FUN_18000cc40(plVar8);
                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                        }
                                        if ((-1 < (int)*plVar12) &&
                                           (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                          lVar15 = FUN_18000cc40(plVar12);
                                          (**(code **)(lVar15 + 0x30))(plVar12);
                                        }
                                        uVar16 = 0x36;
                                        pcVar17 = local_378;
                                        pcVar18 = local_370;
                                        plVar12 = local_360;
                                      }
                                      else {
                                        if ((-1 < (int)*plVar12) &&
                                           (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                          lVar11 = FUN_18000cc40(plVar12);
                                          (**(code **)(lVar11 + 0x30))(plVar12);
                                        }
                                        plVar12 = local_328;
                                        pcVar18 = DAT_1803ae848;
                                        if (iVar5 != 0) {
                                          pcVar18 = DAT_1803ae840;
                                        }
                                        iVar5 = FUN_180001510(local_328);
                                        pcVar17 = _Py_NoneStruct_exref;
                                        if (iVar5 == -1) {
                                          FUN_18000ce30(param_1,&local_388);
LAB_18004acd1:
                                          if ((-1 < (int)*plVar8) &&
                                             (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                            lVar15 = FUN_18000cc40(plVar8);
                                            (**(code **)(lVar15 + 0x30))(plVar8);
                                          }
                                          uVar16 = 0x38;
                                          pcVar17 = local_378;
                                          pcVar18 = local_370;
                                          plVar12 = local_360;
                                        }
                                        else if (iVar5 == 0) {
                                          if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
                                            *(int *)_Py_NoneStruct_exref =
                                                 *(int *)_Py_NoneStruct_exref + 1;
                                          }
LAB_18004ae03:
                                          uVar16 = 0x33;
                                          *(undefined4 *)(local_360 + 5) = 0x33;
                                          local_298 = local_1c0;
                                          local_d8 = local_378;
                                          local_c8 = local_2a0;
                                          local_4b8 = (longlong *)DAT_1803ae860;
                                          local_e8 = plVar8;
                                          local_e0 = pcVar18;
                                          local_d0 = pcVar17;
                                          plVar9 = (longlong *)
                                                   FUN_18037d030(param_1,lVar15,&local_298,&local_e8
                                                                );
                                          if ((-1 < (int)*plVar8) &&
                                             (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                            lVar15 = FUN_18000cc40(plVar8);
                                            (**(code **)(lVar15 + 0x30))(plVar8);
                                          }
                                          if ((-1 < *(int *)pcVar17) &&
                                             (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1,
                                             *(longlong *)pcVar17 == 0)) {
                                            lVar15 = FUN_18000cc40(pcVar17);
                                            (**(code **)(lVar15 + 0x30))(pcVar17);
                                          }
                                          pcVar18 = local_370;
                                          if (plVar9 == (longlong *)0x0) goto LAB_18004cfbb;
LAB_18004b39f:
                                          local_320 = plVar9;
                                          plVar8 = local_358;
                                          pcVar19 = DAT_1803ae848;
                                          iVar5 = *(int *)DAT_1803ae848 + 1;
                                          iVar7 = *(int *)DAT_1803ae848;
                                          if (iVar5 != 0) {
                                            *(int *)DAT_1803ae848 = iVar5;
                                            iVar7 = iVar5;
                                          }
                                          pcVar17 = DAT_1803ae848;
                                          iVar5 = iVar7 + 1;
                                          local_318 = pcVar19;
                                          local_350 = DAT_1803ae848;
                                          if (iVar5 != 0) {
                                            *(int *)DAT_1803ae848 = iVar5;
                                            local_350 = DAT_1803ae848;
                                            iVar7 = iVar5;
                                          }
                                          local_308 = local_350;
                                          iVar5 = iVar7 + 1;
                                          local_310 = pcVar17;
                                          if (iVar5 != 0) {
                                            DAT_1803ae848 = local_350;
                                            *(int *)local_350 = iVar5;
                                            local_350 = DAT_1803ae848;
                                            iVar7 = iVar5;
                                          }
                                          local_300 = local_350;
                                          iVar5 = iVar7 + 1;
                                          if (iVar5 != 0) {
                                            DAT_1803ae848 = local_350;
                                            *(int *)local_350 = iVar5;
                                            local_350 = DAT_1803ae848;
                                            iVar7 = iVar5;
                                          }
                                          local_2f8 = local_350;
                                          iVar5 = iVar7 + 1;
                                          if (iVar5 != 0) {
                                            DAT_1803ae848 = local_350;
                                            *(int *)local_350 = iVar5;
                                            local_350 = DAT_1803ae848;
                                            iVar7 = iVar5;
                                          }
                                          local_2f0 = local_350;
                                          iVar5 = iVar7 + 1;
                                          if (iVar5 != 0) {
                                            DAT_1803ae848 = local_350;
                                            *(int *)local_350 = iVar5;
                                            local_350 = DAT_1803ae848;
                                            iVar7 = iVar5;
                                          }
                                          local_2e8 = local_350;
                                          iVar5 = iVar7 + 1;
                                          if (iVar5 != 0) {
                                            DAT_1803ae848 = local_350;
                                            *(int *)local_350 = iVar5;
                                            local_350 = DAT_1803ae848;
                                            iVar7 = iVar5;
                                          }
                                          local_2e0 = local_350;
                                          iVar5 = iVar7 + 1;
                                          if (iVar5 != 0) {
                                            DAT_1803ae848 = local_350;
                                            *(int *)local_350 = iVar5;
                                            local_350 = DAT_1803ae848;
                                            iVar7 = iVar5;
                                          }
                                          local_368 = local_350;
                                          iVar5 = iVar7 + 1;
                                          if (iVar5 != 0) {
                                            DAT_1803ae848 = local_350;
                                            *(int *)local_350 = iVar5;
                                            local_350 = DAT_1803ae848;
                                            iVar7 = iVar5;
                                          }
                                          DAT_1803ae848 = local_350;
                                          if (iVar7 + 1 != 0) {
                                            *(int *)local_350 = iVar7 + 1;
                                          }
                                          local_2d8 = local_368;
                                          local_2d0 = local_350;
                                          local_2c8 = local_2e8;
                                          local_290 = local_308;
                                          local_280 = local_300;
                                          local_270 = local_2f8;
                                          local_260 = local_2f0;
                                          local_128 = local_2e0;
                                          iVar5 = FUN_180001510(local_358);
                                          if (iVar5 == -1) {
                                            FUN_18000ce30(param_1,&local_388);
                                            uVar16 = 0x56;
                                            pcVar17 = local_378;
                                            plVar12 = local_360;
                                            pcVar19 = local_380;
                                            goto LAB_18004cfd4;
                                          }
                                          plVar12 = local_360;
                                          if (iVar5 == 0) goto LAB_18004cc2f;
                                          lVar15 = FUN_180049280(param_1);
                                          if (lVar15 == 0) {
                                            FUN_180373340(param_1,&local_388,DAT_1803ae880);
                                          }
                                          else {
                                            plVar12 = (longlong *)
                                                      FUN_18037f0c0(param_1,lVar15,DAT_1803ae888);
                                            if (plVar12 == (longlong *)0x0) {
                                              FUN_18000ce30(param_1,&local_388);
                                            }
                                            else {
                                              plVar8 = (longlong *)
                                                       FUN_180382540(plVar8,DAT_1803ae890);
                                              if (plVar8 != (longlong *)0x0) {
                                                uVar16 = 0x57;
                                                local_160 = DAT_1803ae898;
                                                *(undefined4 *)(local_360 + 5) = 0x57;
                                                local_168 = plVar8;
                                                pcVar18 = (code *)FUN_180377750(param_1,plVar12,
                                                                                &local_168);
                                                if ((-1 < (int)*plVar12) &&
                                                   (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                  lVar15 = FUN_18000cc40(plVar12);
                                                  (**(code **)(lVar15 + 0x30))(plVar12);
                                                }
                                                if ((-1 < (int)*plVar8) &&
                                                   (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                  lVar15 = FUN_18000cc40(plVar8);
                                                  (**(code **)(lVar15 + 0x30))(plVar8);
                                                }
                                                if (pcVar18 == (code *)0x0) goto LAB_18004cfbb;
                                                local_318 = pcVar18;
                                                if ((-1 < *(int *)pcVar19) &&
                                                   (*(longlong *)pcVar19 = *(longlong *)pcVar19 + -1
                                                   , *(longlong *)pcVar19 == 0)) {
                                                  lVar15 = FUN_18000cc40(pcVar19);
                                                  (**(code **)(lVar15 + 0x30))(pcVar19);
                                                }
                                                lVar15 = FUN_180049280(param_1);
                                                if (lVar15 == 0) {
                                                  FUN_180373340(param_1,&local_388,DAT_1803ae880);
                                                  uVar16 = 0x58;
                                                  pcVar17 = local_378;
                                                  pcVar18 = local_370;
                                                  plVar12 = local_360;
                                                  pcVar19 = local_380;
                                                }
                                                else {
                                                  plVar8 = (longlong *)
                                                           FUN_18037f0c0(param_1,lVar15,
                                                                         DAT_1803ae888);
                                                  if (plVar8 == (longlong *)0x0) {
                                                    FUN_18000ce30(param_1,&local_388);
                                                    uVar16 = 0x58;
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    plVar12 = local_360;
                                                    pcVar19 = local_380;
                                                  }
                                                  else {
                                                    plVar12 = (longlong *)
                                                              FUN_180382540(local_358,DAT_1803ae8a0)
                                                    ;
                                                    if (plVar12 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_388);
                                                      if ((-1 < (int)*plVar8) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        lVar15 = FUN_18000cc40(plVar8);
                                                        (**(code **)(lVar15 + 0x30))(plVar8);
                                                      }
                                                      uVar16 = 0x58;
                                                      pcVar17 = local_378;
                                                      pcVar18 = local_370;
                                                      plVar12 = local_360;
                                                      pcVar19 = local_380;
                                                    }
                                                    else {
                                                      uVar16 = 0x58;
                                                      local_150 = DAT_1803ae898;
                                                      *(undefined4 *)(local_360 + 5) = 0x58;
                                                      local_158 = plVar12;
                                                      plVar9 = (longlong *)
                                                               FUN_180377750(param_1,plVar8,
                                                                             &local_158);
                                                      if ((-1 < (int)*plVar8) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        lVar15 = FUN_18000cc40(plVar8);
                                                        (**(code **)(lVar15 + 0x30))(plVar8);
                                                      }
                                                      if ((-1 < (int)*plVar12) &&
                                                         (*plVar12 = *plVar12 + -1, *plVar12 == 0))
                                                      {
                                                        lVar15 = FUN_18000cc40(plVar12);
                                                        (**(code **)(lVar15 + 0x30))(plVar12);
                                                      }
                                                      if (plVar9 == (longlong *)0x0)
                                                      goto LAB_18004cfbb;
                                                      local_1f8 = plVar9;
                                                      iVar5 = FUN_180386310(plVar9,DAT_1803ae898);
                                                      piVar1 = DAT_1803ae8a8;
                                                      pcVar19 = DAT_1803ae848;
                                                      if (iVar5 == -1) {
LAB_18004b792:
                                                        FUN_18000ce30(param_1,&local_388);
                                                        uVar16 = 0x59;
                                                        pcVar17 = local_378;
                                                        pcVar18 = local_370;
                                                        plVar12 = local_360;
                                                        pcVar19 = local_380;
                                                      }
                                                      else {
                                                        if (iVar5 == 1) {
                                                          plVar8 = (longlong *)
                                                                   FUN_180374a40(param_1,3);
                                                          if (*piVar1 + 1 != 0) {
                                                            *piVar1 = *piVar1 + 1;
                                                          }
                                                          FUN_18000cc80(plVar8,0,piVar1);
                                                          lVar15 = FUN_180375520(param_1,plVar9,
                                                                                 DAT_1803ae848);
                                                          if (lVar15 == 0) {
                                                            FUN_18000ce30(param_1,&local_388);
                                                            uVar16 = 0x59;
                                                            goto LAB_18004adca;
                                                          }
                                                          FUN_18000cc80(plVar8,1,lVar15);
                                                          if (*DAT_1803ae8b0 + 1 != 0) {
                                                            *DAT_1803ae8b0 = *DAT_1803ae8b0 + 1;
                                                          }
                                                          FUN_18000cc80(plVar8,2);
                                                          pcVar19 = (code *)PyUnicode_Join(
                                                  DAT_1803ae848,plVar8);
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar15 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar15 + 0x30))(plVar8);
                                                  }
                                                  if (pcVar19 == (code *)0x0) goto LAB_18004b792;
                                                  }
                                                  else if (*(int *)DAT_1803ae848 + 1 != 0) {
                                                    *(int *)DAT_1803ae848 =
                                                         *(int *)DAT_1803ae848 + 1;
                                                  }
                                                  local_310 = pcVar19;
                                                  if ((-1 < *(int *)pcVar17) &&
                                                     (*(longlong *)pcVar17 =
                                                           *(longlong *)pcVar17 + -1,
                                                     *(longlong *)pcVar17 == 0)) {
                                                    lVar15 = FUN_18000cc40(pcVar17);
                                                    (**(code **)(lVar15 + 0x30))(pcVar17);
                                                  }
                                                  lVar15 = FUN_180049280(param_1);
                                                  if (lVar15 == 0) {
                                                    FUN_180373340(param_1,&local_388,DAT_1803ae880);
                                                    uVar16 = 0x5a;
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    plVar12 = local_360;
                                                    pcVar19 = local_380;
                                                  }
                                                  else {
                                                    plVar8 = (longlong *)
                                                             FUN_18037f0c0(param_1,lVar15,
                                                                           DAT_1803ae888);
                                                    if (plVar8 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_388);
                                                      uVar16 = 0x5a;
                                                      pcVar17 = local_378;
                                                      pcVar18 = local_370;
                                                      plVar12 = local_360;
                                                      pcVar19 = local_380;
                                                    }
                                                    else {
                                                      plVar9 = (longlong *)
                                                               FUN_180382540(local_358,DAT_1803ae8b8
                                                                            );
                                                      plVar12 = local_360;
                                                      if (plVar9 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_388);
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                                        }
                                                        uVar16 = 0x5a;
                                                        pcVar17 = local_378;
                                                        pcVar18 = local_370;
                                                        plVar12 = local_360;
                                                        pcVar19 = local_380;
                                                      }
                                                      else {
                                                        uVar16 = 0x5a;
                                                        local_140 = DAT_1803ae898;
                                                        *(undefined4 *)(local_360 + 5) = 0x5a;
                                                        local_148 = plVar9;
                                                        plVar10 = (longlong *)
                                                                  FUN_180377750(param_1,plVar8,
                                                                                &local_148);
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                                        }
                                                        if ((-1 < (int)*plVar9) &&
                                                           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar9);
                                                          (**(code **)(lVar15 + 0x30))(plVar9);
                                                        }
                                                        if (plVar10 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_388);
                                                          pcVar17 = local_378;
                                                          pcVar18 = local_370;
                                                          pcVar19 = local_380;
                                                        }
                                                        else {
                                                          local_1f0 = plVar10;
                                                          iVar5 = FUN_180386310(plVar10,
                                                  DAT_1803ae898);
                                                  piVar1 = DAT_1803ae8c0;
                                                  pcVar19 = DAT_1803ae848;
                                                  if (iVar5 == -1) {
LAB_18004b9e3:
                                                    FUN_18000ce30(param_1,&local_388);
                                                    uVar16 = 0x5b;
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    pcVar19 = local_380;
                                                  }
                                                  else if (iVar5 == 1) {
                                                    plVar8 = (longlong *)FUN_180374a40(param_1,3);
                                                    if (*piVar1 + 1 != 0) {
                                                      *piVar1 = *piVar1 + 1;
                                                    }
                                                    FUN_18000cc80(plVar8,0,piVar1);
                                                    lVar15 = FUN_180375520(param_1,plVar10,
                                                                           DAT_1803ae848);
                                                    if (lVar15 != 0) {
                                                      FUN_18000cc80(plVar8,1,lVar15);
                                                      if (*DAT_1803ae8b0 + 1 != 0) {
                                                        *DAT_1803ae8b0 = *DAT_1803ae8b0 + 1;
                                                      }
                                                      FUN_18000cc80(plVar8,2);
                                                      pcVar19 = (code *)PyUnicode_Join(DAT_1803ae848
                                                                                       ,plVar8);
                                                      if ((-1 < (int)*plVar8) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        lVar15 = FUN_18000cc40(plVar8);
                                                        (**(code **)(lVar15 + 0x30))(plVar8);
                                                      }
                                                      if (pcVar19 == (code *)0x0)
                                                      goto LAB_18004b9e3;
                                                      goto LAB_18004baf8;
                                                    }
                                                    FUN_18000ce30(param_1,&local_388);
                                                    uVar16 = 0x5b;
LAB_18004ba57:
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    pcVar19 = local_380;
                                                    if ((-1 < (int)*plVar8) &&
                                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                      lVar15 = FUN_18000cc40(plVar8);
                                                      (**(code **)(lVar15 + 0x30))(plVar8);
                                                      pcVar17 = local_378;
                                                      pcVar18 = local_370;
                                                      pcVar19 = local_380;
                                                    }
                                                  }
                                                  else {
                                                    if (*(int *)DAT_1803ae848 + 1 != 0) {
                                                      *(int *)DAT_1803ae848 =
                                                           *(int *)DAT_1803ae848 + 1;
                                                    }
LAB_18004baf8:
                                                    pcVar17 = local_290;
                                                    local_308 = pcVar19;
                                                    if ((-1 < *(int *)local_290) &&
                                                       (*(longlong *)local_290 =
                                                             *(longlong *)local_290 + -1,
                                                       *(longlong *)local_290 == 0)) {
                                                      lVar15 = FUN_18000cc40(local_290);
                                                      (**(code **)(lVar15 + 0x30))(pcVar17);
                                                    }
                                                    lVar15 = FUN_180049280(param_1);
                                                    if (lVar15 == 0) {
                                                      FUN_180373340(param_1,&local_388,DAT_1803ae880
                                                                   );
                                                      uVar16 = 0x5c;
                                                      pcVar17 = local_378;
                                                      pcVar18 = local_370;
                                                      pcVar19 = local_380;
                                                    }
                                                    else {
                                                      plVar8 = (longlong *)
                                                               FUN_18037f0c0(param_1,lVar15,
                                                                             DAT_1803ae888);
                                                      if (plVar8 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_388);
                                                        uVar16 = 0x5c;
                                                        pcVar17 = local_378;
                                                        pcVar18 = local_370;
                                                        pcVar19 = local_380;
                                                      }
                                                      else {
                                                        plVar9 = (longlong *)
                                                                 FUN_180382540(local_358,
                                                                               DAT_1803ae8c8);
                                                        if (plVar9 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_388);
                                                          if ((-1 < (int)*plVar8) &&
                                                             (*plVar8 = *plVar8 + -1, *plVar8 == 0))
                                                          {
                                                            lVar15 = FUN_18000cc40(plVar8);
                                                            (**(code **)(lVar15 + 0x30))(plVar8);
                                                          }
                                                          uVar16 = 0x5c;
                                                          pcVar17 = local_378;
                                                          pcVar18 = local_370;
                                                          pcVar19 = local_380;
                                                        }
                                                        else {
                                                          uVar16 = 0x5c;
                                                          local_338 = DAT_1803ae898;
                                                          *(undefined4 *)(plVar12 + 5) = 0x5c;
                                                          local_340 = plVar9;
                                                          plVar10 = (longlong *)
                                                                    FUN_180377750(param_1,plVar8,
                                                                                  &local_340);
                                                          if ((-1 < (int)*plVar8) &&
                                                             (*plVar8 = *plVar8 + -1, *plVar8 == 0))
                                                          {
                                                            lVar15 = FUN_18000cc40(plVar8);
                                                            (**(code **)(lVar15 + 0x30))(plVar8);
                                                          }
                                                          if ((-1 < (int)*plVar9) &&
                                                             (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                                          {
                                                            lVar15 = FUN_18000cc40(plVar9);
                                                            (**(code **)(lVar15 + 0x30))(plVar9);
                                                          }
                                                          if (plVar10 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_388);
                                                            pcVar17 = local_378;
                                                            pcVar18 = local_370;
                                                            pcVar19 = local_380;
                                                          }
                                                          else {
                                                            local_1e8 = plVar10;
                                                            iVar5 = FUN_180386310(plVar10,
                                                  DAT_1803ae898);
                                                  piVar1 = DAT_1803ae8d0;
                                                  pcVar19 = DAT_1803ae848;
                                                  if (iVar5 == -1) {
LAB_18004bc4c:
                                                    FUN_18000ce30(param_1,&local_388);
                                                    uVar16 = 0x5d;
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    pcVar19 = local_380;
                                                  }
                                                  else {
                                                    if (iVar5 == 1) {
                                                      plVar8 = (longlong *)FUN_180374a40(param_1,3);
                                                      if (*piVar1 + 1 != 0) {
                                                        *piVar1 = *piVar1 + 1;
                                                      }
                                                      FUN_18000cc80(plVar8,0,piVar1);
                                                      lVar15 = FUN_180375520(param_1,plVar10,
                                                                             DAT_1803ae848);
                                                      if (lVar15 == 0) {
                                                        FUN_18000ce30(param_1,&local_388);
                                                        uVar16 = 0x5d;
                                                        goto LAB_18004ba57;
                                                      }
                                                      FUN_18000cc80(plVar8,1,lVar15);
                                                      if (*DAT_1803ae8b0 + 1 != 0) {
                                                        *DAT_1803ae8b0 = *DAT_1803ae8b0 + 1;
                                                      }
                                                      FUN_18000cc80(plVar8,2);
                                                      pcVar19 = (code *)PyUnicode_Join(DAT_1803ae848
                                                                                       ,plVar8);
                                                      if ((-1 < (int)*plVar8) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        lVar15 = FUN_18000cc40(plVar8);
                                                        (**(code **)(lVar15 + 0x30))(plVar8);
                                                      }
                                                      if (pcVar19 == (code *)0x0)
                                                      goto LAB_18004bc4c;
                                                    }
                                                    else if (*(int *)DAT_1803ae848 + 1 != 0) {
                                                      *(int *)DAT_1803ae848 =
                                                           *(int *)DAT_1803ae848 + 1;
                                                    }
                                                    pcVar17 = local_280;
                                                    local_300 = pcVar19;
                                                    if ((-1 < *(int *)local_280) &&
                                                       (*(longlong *)local_280 =
                                                             *(longlong *)local_280 + -1,
                                                       *(longlong *)local_280 == 0)) {
                                                      lVar15 = FUN_18000cc40(local_280);
                                                      (**(code **)(lVar15 + 0x30))(pcVar17);
                                                    }
                                                    lVar15 = FUN_180049280(param_1);
                                                    if (lVar15 == 0) {
                                                      FUN_180373340(param_1,&local_388,DAT_1803ae880
                                                                   );
                                                      uVar16 = 0x5e;
                                                      pcVar17 = local_378;
                                                      pcVar18 = local_370;
                                                      pcVar19 = local_380;
                                                    }
                                                    else {
                                                      plVar8 = (longlong *)
                                                               FUN_18037f0c0(param_1,lVar15,
                                                                             DAT_1803ae888);
                                                      if (plVar8 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_388);
                                                        uVar16 = 0x5e;
                                                        pcVar17 = local_378;
                                                        pcVar18 = local_370;
                                                        pcVar19 = local_380;
                                                      }
                                                      else {
                                                        pcVar19 = (code *)FUN_180382540(local_358,
                                                                                                                                                                                
                                                  DAT_1803ae8d8);
                                                  if (pcVar19 == (code *)0x0) {
                                                    FUN_18000ce30(param_1,&local_388);
                                                    if ((-1 < (int)*plVar8) &&
                                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                      lVar15 = FUN_18000cc40(plVar8);
                                                      (**(code **)(lVar15 + 0x30))(plVar8);
                                                    }
                                                    uVar16 = 0x5e;
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    pcVar19 = local_380;
                                                  }
                                                  else {
                                                    uVar16 = 0x5e;
                                                    local_288 = DAT_1803ae898;
                                                    *(undefined4 *)(plVar12 + 5) = 0x5e;
                                                    local_290 = pcVar19;
                                                    plVar9 = (longlong *)
                                                             FUN_180377750(param_1,plVar8,&local_290
                                                                          );
                                                    if ((-1 < (int)*plVar8) &&
                                                       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                      lVar15 = FUN_18000cc40(plVar8);
                                                      (**(code **)(lVar15 + 0x30))(plVar8);
                                                    }
                                                    if ((-1 < *(int *)pcVar19) &&
                                                       (*(longlong *)pcVar19 =
                                                             *(longlong *)pcVar19 + -1,
                                                       *(longlong *)pcVar19 == 0)) {
                                                      lVar15 = FUN_18000cc40(pcVar19);
                                                      (**(code **)(lVar15 + 0x30))(pcVar19);
                                                    }
                                                    if (plVar9 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_388);
                                                      pcVar17 = local_378;
                                                      pcVar18 = local_370;
                                                      pcVar19 = local_380;
                                                    }
                                                    else {
                                                      local_1e0 = plVar9;
                                                      iVar5 = FUN_180386310(plVar9,DAT_1803ae898);
                                                      piVar1 = DAT_1803ae8e0;
                                                      pcVar19 = DAT_1803ae848;
                                                      if (iVar5 == -1) {
LAB_18004be97:
                                                        FUN_18000ce30(param_1,&local_388);
                                                        uVar16 = 0x5f;
                                                        pcVar17 = local_378;
                                                        pcVar18 = local_370;
                                                        pcVar19 = local_380;
                                                      }
                                                      else {
                                                        if (iVar5 == 1) {
                                                          plVar8 = (longlong *)
                                                                   FUN_180374a40(param_1,3);
                                                          if (*piVar1 + 1 != 0) {
                                                            *piVar1 = *piVar1 + 1;
                                                          }
                                                          FUN_18000cc80(plVar8,0,piVar1);
                                                          lVar15 = FUN_180375520(param_1,plVar9,
                                                                                 DAT_1803ae848);
                                                          if (lVar15 == 0) {
                                                            FUN_18000ce30(param_1,&local_388);
                                                            uVar16 = 0x5f;
                                                            goto LAB_18004ba57;
                                                          }
                                                          FUN_18000cc80(plVar8,1,lVar15);
                                                          if (*DAT_1803ae8b0 + 1 != 0) {
                                                            *DAT_1803ae8b0 = *DAT_1803ae8b0 + 1;
                                                          }
                                                          FUN_18000cc80(plVar8,2);
                                                          pcVar19 = (code *)PyUnicode_Join(
                                                  DAT_1803ae848,plVar8);
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar15 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar15 + 0x30))(plVar8);
                                                  }
                                                  if (pcVar19 == (code *)0x0) goto LAB_18004be97;
                                                  }
                                                  else if (*(int *)DAT_1803ae848 + 1 != 0) {
                                                    *(int *)DAT_1803ae848 =
                                                         *(int *)DAT_1803ae848 + 1;
                                                  }
                                                  pcVar17 = local_270;
                                                  local_2f8 = pcVar19;
                                                  if ((-1 < *(int *)local_270) &&
                                                     (*(longlong *)local_270 =
                                                           *(longlong *)local_270 + -1,
                                                     *(longlong *)local_270 == 0)) {
                                                    lVar15 = FUN_18000cc40(local_270);
                                                    (**(code **)(lVar15 + 0x30))(pcVar17);
                                                  }
                                                  lVar15 = FUN_180049280(param_1);
                                                  if (lVar15 == 0) {
                                                    FUN_180373340(param_1,&local_388,DAT_1803ae880);
                                                    uVar16 = 0x60;
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    pcVar19 = local_380;
                                                  }
                                                  else {
                                                    plVar8 = (longlong *)
                                                             FUN_18037f0c0(param_1,lVar15,
                                                                           DAT_1803ae888);
                                                    if (plVar8 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_388);
                                                      uVar16 = 0x60;
                                                      pcVar17 = local_378;
                                                      pcVar18 = local_370;
                                                      pcVar19 = local_380;
                                                    }
                                                    else {
                                                      pcVar19 = (code *)FUN_180382540(local_358,
                                                                                      DAT_1803ae8e8)
                                                      ;
                                                      if (pcVar19 == (code *)0x0) {
                                                        FUN_18000ce30(param_1,&local_388);
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                                        }
                                                        uVar16 = 0x60;
                                                        pcVar17 = local_378;
                                                        pcVar18 = local_370;
                                                        pcVar19 = local_380;
                                                      }
                                                      else {
                                                        uVar16 = 0x60;
                                                        local_278 = DAT_1803ae898;
                                                        *(undefined4 *)(plVar12 + 5) = 0x60;
                                                        local_280 = pcVar19;
                                                        plVar9 = (longlong *)
                                                                 FUN_180377750(param_1,plVar8,
                                                                               &local_280);
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                                        }
                                                        if ((-1 < *(int *)pcVar19) &&
                                                           (*(longlong *)pcVar19 =
                                                                 *(longlong *)pcVar19 + -1,
                                                           *(longlong *)pcVar19 == 0)) {
                                                          lVar15 = FUN_18000cc40(pcVar19);
                                                          (**(code **)(lVar15 + 0x30))(pcVar19);
                                                        }
                                                        if (plVar9 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_388);
                                                          pcVar17 = local_378;
                                                          pcVar18 = local_370;
                                                          pcVar19 = local_380;
                                                        }
                                                        else {
                                                          local_1d8 = plVar9;
                                                          iVar5 = FUN_180386310(plVar9,DAT_1803ae898
                                                                               );
                                                          piVar1 = DAT_1803ae8f0;
                                                          pcVar19 = DAT_1803ae848;
                                                          if (iVar5 == -1) {
LAB_18004c0e8:
                                                            FUN_18000ce30(param_1,&local_388);
                                                            uVar16 = 0x61;
                                                            pcVar17 = local_378;
                                                            pcVar18 = local_370;
                                                            pcVar19 = local_380;
                                                          }
                                                          else {
                                                            if (iVar5 == 1) {
                                                              plVar8 = (longlong *)
                                                                       FUN_180374a40(param_1,3);
                                                              if (*piVar1 + 1 != 0) {
                                                                *piVar1 = *piVar1 + 1;
                                                              }
                                                              FUN_18000cc80(plVar8,0,piVar1);
                                                              lVar15 = FUN_180375520(param_1,plVar9,
                                                                                     DAT_1803ae848);
                                                              if (lVar15 == 0) {
                                                                FUN_18000ce30(param_1,&local_388);
                                                                uVar16 = 0x61;
                                                                goto LAB_18004ba57;
                                                              }
                                                              FUN_18000cc80(plVar8,1,lVar15);
                                                              if (*DAT_1803ae8b0 + 1 != 0) {
                                                                *DAT_1803ae8b0 = *DAT_1803ae8b0 + 1;
                                                              }
                                                              FUN_18000cc80(plVar8,2);
                                                              pcVar19 = (code *)PyUnicode_Join(
                                                  DAT_1803ae848,plVar8);
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar15 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar15 + 0x30))(plVar8);
                                                  }
                                                  if (pcVar19 == (code *)0x0) goto LAB_18004c0e8;
                                                  }
                                                  else if (*(int *)DAT_1803ae848 + 1 != 0) {
                                                    *(int *)DAT_1803ae848 =
                                                         *(int *)DAT_1803ae848 + 1;
                                                  }
                                                  pcVar17 = local_260;
                                                  local_2f0 = pcVar19;
                                                  if ((-1 < *(int *)local_260) &&
                                                     (*(longlong *)local_260 =
                                                           *(longlong *)local_260 + -1,
                                                     *(longlong *)local_260 == 0)) {
                                                    lVar15 = FUN_18000cc40(local_260);
                                                    (**(code **)(lVar15 + 0x30))(pcVar17);
                                                  }
                                                  lVar15 = FUN_180049280(param_1);
                                                  if (lVar15 == 0) {
                                                    FUN_180373340(param_1,&local_388,DAT_1803ae880);
                                                    uVar16 = 0x62;
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    pcVar19 = local_380;
                                                  }
                                                  else {
                                                    plVar8 = (longlong *)
                                                             FUN_18037f0c0(param_1,lVar15,
                                                                           DAT_1803ae888);
                                                    if (plVar8 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_388);
                                                      uVar16 = 0x62;
                                                      pcVar17 = local_378;
                                                      pcVar18 = local_370;
                                                      pcVar19 = local_380;
                                                    }
                                                    else {
                                                      pcVar19 = (code *)FUN_180382540(local_358,
                                                                                      DAT_1803ae8f8)
                                                      ;
                                                      if (pcVar19 == (code *)0x0) {
                                                        FUN_18000ce30(param_1,&local_388);
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                                        }
                                                        uVar16 = 0x62;
                                                        pcVar17 = local_378;
                                                        pcVar18 = local_370;
                                                        pcVar19 = local_380;
                                                      }
                                                      else {
                                                        uVar16 = 0x62;
                                                        local_268 = DAT_1803ae898;
                                                        *(undefined4 *)(plVar12 + 5) = 0x62;
                                                        local_270 = pcVar19;
                                                        plVar9 = (longlong *)
                                                                 FUN_180377750(param_1,plVar8,
                                                                               &local_270);
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                                        }
                                                        if ((-1 < *(int *)pcVar19) &&
                                                           (*(longlong *)pcVar19 =
                                                                 *(longlong *)pcVar19 + -1,
                                                           *(longlong *)pcVar19 == 0)) {
                                                          lVar15 = FUN_18000cc40(pcVar19);
                                                          (**(code **)(lVar15 + 0x30))(pcVar19);
                                                        }
                                                        if (plVar9 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_388);
                                                          pcVar17 = local_378;
                                                          pcVar18 = local_370;
                                                          pcVar19 = local_380;
                                                        }
                                                        else {
                                                          local_1d0 = plVar9;
                                                          iVar5 = FUN_180386310(plVar9,DAT_1803ae898
                                                                               );
                                                          piVar1 = DAT_1803ae900;
                                                          if (iVar5 == -1) {
                                                            FUN_18000ce30(param_1,&local_388);
                                                            uVar16 = 99;
                                                            pcVar17 = local_378;
                                                            pcVar18 = local_370;
                                                            pcVar19 = local_380;
                                                          }
                                                          else {
                                                            if (iVar5 == 1) {
                                                              plVar8 = (longlong *)
                                                                       FUN_180374a40(param_1,3);
                                                              if (*piVar1 + 1 != 0) {
                                                                *piVar1 = *piVar1 + 1;
                                                              }
                                                              FUN_18000cc80(plVar8,0,piVar1);
                                                              lVar15 = FUN_180375520(param_1,plVar9,
                                                                                     DAT_1803ae848);
                                                              if (lVar15 == 0) {
                                                                FUN_18000ce30(param_1,&local_388);
                                                                uVar16 = 100;
                                                                goto LAB_18004ba57;
                                                              }
                                                              FUN_18000cc80(plVar8,1,lVar15);
                                                              if (*DAT_1803ae8b0 + 1 != 0) {
                                                                *DAT_1803ae8b0 = *DAT_1803ae8b0 + 1;
                                                              }
                                                              FUN_18000cc80(plVar8,2);
                                                              pcVar19 = (code *)PyUnicode_Join(
                                                  DAT_1803ae848,plVar8);
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar15 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar15 + 0x30))(plVar8);
                                                  }
                                                  pcVar17 = local_2c8;
                                                  if (pcVar19 == (code *)0x0) {
                                                    FUN_18000ce30(param_1,&local_388);
                                                    uVar16 = 100;
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    pcVar19 = local_380;
                                                    goto LAB_18004cfd4;
                                                  }
                                                  local_2e8 = pcVar19;
                                                  if ((-1 < *(int *)local_2c8) &&
                                                     (*(longlong *)local_2c8 =
                                                           *(longlong *)local_2c8 + -1,
                                                     *(longlong *)local_2c8 == 0)) {
                                                    lVar15 = FUN_18000cc40(local_2c8);
                                                    (**(code **)(lVar15 + 0x30))(pcVar17);
                                                  }
                                                  }
                                                  lVar15 = FUN_180049280(param_1);
                                                  if (lVar15 == 0) {
                                                    FUN_180373340(param_1,&local_388,DAT_1803ae880);
                                                    uVar16 = 0x65;
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    pcVar19 = local_380;
                                                  }
                                                  else {
                                                    plVar8 = (longlong *)
                                                             FUN_18037f0c0(param_1,lVar15,
                                                                           DAT_1803ae888);
                                                    if (plVar8 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_388);
                                                      uVar16 = 0x65;
                                                      pcVar17 = local_378;
                                                      pcVar18 = local_370;
                                                      pcVar19 = local_380;
                                                    }
                                                    else {
                                                      pcVar19 = (code *)FUN_180382540(local_358,
                                                                                      DAT_1803ae908)
                                                      ;
                                                      if (pcVar19 == (code *)0x0) {
                                                        FUN_18000ce30(param_1,&local_388);
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                                        }
                                                        uVar16 = 0x65;
                                                        pcVar17 = local_378;
                                                        pcVar18 = local_370;
                                                        pcVar19 = local_380;
                                                      }
                                                      else {
                                                        uVar16 = 0x65;
                                                        local_258 = DAT_1803ae898;
                                                        *(undefined4 *)(plVar12 + 5) = 0x65;
                                                        local_260 = pcVar19;
                                                        plVar9 = (longlong *)
                                                                 FUN_180377750(param_1,plVar8,
                                                                               &local_260);
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                                        }
                                                        if ((-1 < *(int *)pcVar19) &&
                                                           (*(longlong *)pcVar19 =
                                                                 *(longlong *)pcVar19 + -1,
                                                           *(longlong *)pcVar19 == 0)) {
                                                          lVar15 = FUN_18000cc40(pcVar19);
                                                          (**(code **)(lVar15 + 0x30))(pcVar19);
                                                        }
                                                        if (plVar9 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_388);
                                                          pcVar17 = local_378;
                                                          pcVar18 = local_370;
                                                          pcVar19 = local_380;
                                                        }
                                                        else {
                                                          local_210 = plVar9;
                                                          iVar5 = FUN_180386310(plVar9,DAT_1803ae898
                                                                               );
                                                          piVar1 = DAT_1803ae910;
                                                          if (iVar5 == -1) {
                                                            FUN_18000ce30(param_1,&local_388);
                                                            uVar16 = 0x66;
                                                            pcVar17 = local_378;
                                                            pcVar18 = local_370;
                                                            pcVar19 = local_380;
                                                          }
                                                          else {
                                                            if (iVar5 == 1) {
                                                              plVar8 = (longlong *)
                                                                       FUN_180374a40(param_1,3);
                                                              if (*piVar1 + 1 != 0) {
                                                                *piVar1 = *piVar1 + 1;
                                                              }
                                                              FUN_18000cc80(plVar8,0,piVar1);
                                                              lVar15 = FUN_180375520(param_1,plVar9,
                                                                                     DAT_1803ae848);
                                                              if (lVar15 == 0) {
                                                                FUN_18000ce30(param_1,&local_388);
                                                                uVar16 = 0x6a;
                                                                goto LAB_18004ba57;
                                                              }
                                                              FUN_18000cc80(plVar8,1,lVar15);
                                                              if (*DAT_1803ae8b0 + 1 != 0) {
                                                                *DAT_1803ae8b0 = *DAT_1803ae8b0 + 1;
                                                              }
                                                              FUN_18000cc80(plVar8,2);
                                                              pcVar19 = (code *)PyUnicode_Join(
                                                  DAT_1803ae848,plVar8);
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar15 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar15 + 0x30))(plVar8);
                                                  }
                                                  pcVar17 = local_128;
                                                  if (pcVar19 == (code *)0x0) {
                                                    FUN_18000ce30(param_1,&local_388);
                                                    uVar16 = 0x6a;
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    pcVar19 = local_380;
                                                    goto LAB_18004cfd4;
                                                  }
                                                  local_2e0 = pcVar19;
                                                  if ((-1 < *(int *)local_128) &&
                                                     (*(longlong *)local_128 =
                                                           *(longlong *)local_128 + -1,
                                                     *(longlong *)local_128 == 0)) {
                                                    lVar15 = FUN_18000cc40(local_128);
                                                    (**(code **)(lVar15 + 0x30))(pcVar17);
                                                  }
                                                  }
                                                  lVar15 = FUN_180049280(param_1);
                                                  if (lVar15 == 0) {
                                                    FUN_180373340(param_1,&local_388,DAT_1803ae880);
                                                    uVar16 = 0x6c;
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    pcVar19 = local_380;
                                                  }
                                                  else {
                                                    plVar8 = (longlong *)
                                                             FUN_18037f0c0(param_1,lVar15,
                                                                           DAT_1803ae888);
                                                    if (plVar8 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_388);
                                                      uVar16 = 0x6c;
                                                      pcVar17 = local_378;
                                                      pcVar18 = local_370;
                                                      pcVar19 = local_380;
                                                    }
                                                    else {
                                                      pcVar19 = (code *)FUN_180382540(local_358,
                                                                                      DAT_1803ae918)
                                                      ;
                                                      if (pcVar19 == (code *)0x0) {
                                                        FUN_18000ce30(param_1,&local_388);
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                                        }
                                                        uVar16 = 0x6c;
                                                        pcVar17 = local_378;
                                                        pcVar18 = local_370;
                                                        pcVar19 = local_380;
                                                      }
                                                      else {
                                                        uVar16 = 0x6c;
                                                        local_2c0 = DAT_1803ae898;
                                                        *(undefined4 *)(plVar12 + 5) = 0x6c;
                                                        local_2c8 = pcVar19;
                                                        plVar9 = (longlong *)
                                                                 FUN_180377750(param_1,plVar8,
                                                                               &local_2c8);
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                                        }
                                                        if ((-1 < *(int *)pcVar19) &&
                                                           (*(longlong *)pcVar19 =
                                                                 *(longlong *)pcVar19 + -1,
                                                           *(longlong *)pcVar19 == 0)) {
                                                          lVar15 = FUN_18000cc40(pcVar19);
                                                          (**(code **)(lVar15 + 0x30))(pcVar19);
                                                        }
                                                        if (plVar9 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_388);
                                                          pcVar17 = local_378;
                                                          pcVar18 = local_370;
                                                          pcVar19 = local_380;
                                                        }
                                                        else {
                                                          local_248 = plVar9;
                                                          iVar5 = FUN_180386310(plVar9,DAT_1803ae898
                                                                               );
                                                          piVar1 = DAT_1803ae920;
                                                          pcVar19 = DAT_1803ae848;
                                                          if (iVar5 == -1) {
LAB_18004c7d1:
                                                            FUN_18000ce30(param_1,&local_388);
                                                            uVar16 = 0x6d;
                                                            pcVar17 = local_378;
                                                            pcVar18 = local_370;
                                                            pcVar19 = local_380;
                                                          }
                                                          else {
                                                            if (iVar5 == 1) {
                                                              plVar8 = (longlong *)
                                                                       FUN_180374a40(param_1,2);
                                                              if (*piVar1 + 1 != 0) {
                                                                *piVar1 = *piVar1 + 1;
                                                              }
                                                              FUN_18000cc80(plVar8,0,piVar1);
                                                              lVar15 = FUN_180375520(param_1,plVar9,
                                                                                     DAT_1803ae848);
                                                              if (lVar15 == 0) {
                                                                FUN_18000ce30(param_1,&local_388);
                                                                uVar16 = 0x6d;
                                                                goto LAB_18004ba57;
                                                              }
                                                              FUN_18000cc80(plVar8,1,lVar15);
                                                              pcVar19 = (code *)PyUnicode_Join(
                                                  DAT_1803ae848,plVar8);
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar15 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar15 + 0x30))(plVar8);
                                                  }
                                                  if (pcVar19 == (code *)0x0) goto LAB_18004c7d1;
                                                  }
                                                  else if (*(int *)DAT_1803ae848 + 1 != 0) {
                                                    *(int *)DAT_1803ae848 =
                                                         *(int *)DAT_1803ae848 + 1;
                                                  }
                                                  pcVar17 = local_368;
                                                  local_2d8 = pcVar19;
                                                  if ((-1 < *(int *)local_368) &&
                                                     (*(longlong *)local_368 =
                                                           *(longlong *)local_368 + -1,
                                                     *(longlong *)local_368 == 0)) {
                                                    lVar15 = FUN_18000cc40(local_368);
                                                    (**(code **)(lVar15 + 0x30))(pcVar17);
                                                  }
                                                  lVar15 = FUN_180049280(param_1);
                                                  if (lVar15 == 0) {
                                                    FUN_180373340(param_1,&local_388,DAT_1803ae880);
                                                    uVar16 = 0x6e;
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    pcVar19 = local_380;
                                                  }
                                                  else {
                                                    plVar8 = (longlong *)
                                                             FUN_18037f0c0(param_1,lVar15,
                                                                           DAT_1803ae888);
                                                    if (plVar8 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_388);
                                                      uVar16 = 0x6e;
                                                      pcVar17 = local_378;
                                                      pcVar18 = local_370;
                                                      pcVar19 = local_380;
                                                    }
                                                    else {
                                                      plVar9 = (longlong *)
                                                               FUN_180382540(local_358,DAT_1803ae928
                                                                            );
                                                      if (plVar9 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_388);
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                                        }
                                                        uVar16 = 0x6e;
                                                        pcVar17 = local_378;
                                                        pcVar18 = local_370;
                                                        pcVar19 = local_380;
                                                      }
                                                      else {
                                                        uVar16 = 0x6e;
                                                        local_130 = DAT_1803ae848;
                                                        *(undefined4 *)(plVar12 + 5) = 0x6e;
                                                        local_138 = plVar9;
                                                        plVar10 = (longlong *)
                                                                  FUN_180377750(param_1,plVar8,
                                                                                &local_138);
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                                        }
                                                        if ((-1 < (int)*plVar9) &&
                                                           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar9);
                                                          (**(code **)(lVar15 + 0x30))(plVar9);
                                                        }
                                                        if (plVar10 == (longlong *)0x0) {
                                                          FUN_18000ce30(param_1,&local_388);
                                                          pcVar17 = local_378;
                                                          pcVar18 = local_370;
                                                          pcVar19 = local_380;
                                                        }
                                                        else {
                                                          local_240 = plVar10;
                                                          plVar8 = (longlong *)
                                                                   FUN_180001b80(param_1,local_1c0,
                                                                                 DAT_1803ae838);
                                                          if (plVar8 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_388);
                                                            uVar16 = 0x6f;
                                                            pcVar17 = local_378;
                                                            pcVar18 = local_370;
                                                            pcVar19 = local_380;
                                                          }
                                                          else {
                                                            iVar5 = FUN_180001510(plVar8);
                                                            if (iVar5 == -1) {
                                                              FUN_18000ce30(param_1,&local_388);
                                                              if ((-1 < (int)*plVar8) &&
                                                                 (*plVar8 = *plVar8 + -1,
                                                                 *plVar8 == 0)) {
                                                                lVar15 = FUN_18000cc40(plVar8);
                                                                (**(code **)(lVar15 + 0x30))(plVar8)
                                                                ;
                                                              }
                                                              uVar16 = 0x6f;
                                                              pcVar17 = local_378;
                                                              pcVar18 = local_370;
                                                              pcVar19 = local_380;
                                                            }
                                                            else {
                                                              bVar20 = iVar5 != 0;
                                                              if ((-1 < (int)*plVar8) &&
                                                                 (*plVar8 = *plVar8 + -1,
                                                                 *plVar8 == 0)) {
                                                                lVar15 = FUN_18000cc40(plVar8);
                                                                (**(code **)(lVar15 + 0x30))(plVar8)
                                                                ;
                                                              }
                                                              if (iVar5 == 0) {
LAB_18004cb42:
                                                                piVar1 = DAT_1803ae938;
                                                                pcVar17 = local_310;
                                                                pcVar19 = pcVar18;
                                                                if (bVar20) {
                                                                  plVar8 = (longlong *)
                                                                           FUN_180374a40(param_1,2);
                                                                  if (*piVar1 + 1 != 0) {
                                                                    *piVar1 = *piVar1 + 1;
                                                                  }
                                                                  FUN_18000cc80(plVar8,0,piVar1);
                                                                  lVar15 = FUN_180375520(param_1,
                                                  plVar10,DAT_1803ae848);
                                                  if (lVar15 == 0) {
                                                    FUN_18000ce30(param_1,&local_388);
                                                    uVar16 = 0x70;
                                                    goto LAB_18004ba57;
                                                  }
                                                  FUN_18000cc80(plVar8,1,lVar15);
                                                  pcVar18 = (code *)PyUnicode_Join(DAT_1803ae848,
                                                                                   plVar8);
                                                  if ((-1 < (int)*plVar8) &&
                                                     (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                    lVar15 = FUN_18000cc40(plVar8);
                                                    (**(code **)(lVar15 + 0x30))(plVar8);
                                                  }
                                                  pcVar3 = local_350;
                                                  if (pcVar18 == (code *)0x0) {
                                                    FUN_18000ce30(param_1,&local_388);
                                                    uVar16 = 0x70;
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    pcVar19 = local_380;
                                                    goto LAB_18004cfd4;
                                                  }
                                                  pcVar17 = local_310;
                                                  local_2d0 = pcVar18;
                                                  if ((-1 < *(int *)local_350) &&
                                                     (*(longlong *)local_350 =
                                                           *(longlong *)local_350 + -1,
                                                     *(longlong *)local_350 == 0)) {
                                                    lVar15 = FUN_18000cc40(local_350);
                                                    (**(code **)(lVar15 + 0x30))(pcVar3);
                                                    pcVar17 = local_310;
                                                  }
                                                  }
LAB_18004cc2f:
                                                  lVar15 = FUN_180049a40(param_1);
                                                  if (lVar15 == 0) {
                                                    FUN_180373340(param_1,&local_388,DAT_1803ae940);
                                                    uVar16 = 0x72;
                                                    pcVar17 = local_378;
                                                    pcVar18 = local_370;
                                                    pcVar19 = local_380;
                                                  }
                                                  else {
                                                    lVar11 = FUN_180049390(param_1);
                                                    if (lVar11 == 0) {
                                                      FUN_180373340(param_1,&local_388,DAT_1803ae948
                                                                   );
                                                      uVar16 = 0x73;
                                                      pcVar17 = local_378;
                                                      pcVar18 = local_370;
                                                      pcVar19 = local_380;
                                                    }
                                                    else {
                                                      local_b8 = local_228;
                                                      uVar16 = 0x73;
                                                      local_b0 = local_2b8;
                                                      local_a0 = local_2b0;
                                                      local_98 = local_2d0;
                                                      local_88 = local_308;
                                                      local_80 = local_300;
                                                      local_78 = local_2f8;
                                                      local_70 = local_2f0;
                                                      local_68 = local_2e8;
                                                      local_60 = local_2e0;
                                                      local_58 = local_2d8;
                                                      *(undefined4 *)(plVar12 + 5) = 0x73;
                                                      local_a8 = pcVar19;
                                                      local_90 = pcVar17;
                                                      plVar8 = (longlong *)
                                                               FUN_18037ccb0(param_1,lVar11,
                                                                             &local_b8,DAT_1803ae950
                                                                            );
                                                      if (plVar8 == (longlong *)0x0) {
                                                        FUN_18000ce30(param_1,&local_388);
                                                        pcVar17 = local_378;
                                                        pcVar18 = local_370;
                                                        pcVar19 = local_380;
                                                      }
                                                      else {
                                                        uVar16 = 0x72;
                                                        *(undefined4 *)(plVar12 + 5) = 0x72;
                                                        plVar9 = (longlong *)
                                                                 FUN_180376770(param_1,lVar15,plVar8
                                                                              );
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                                        }
                                                        if (plVar9 == (longlong *)0x0) {
LAB_18004cfbb:
                                                          FUN_18000ce30(param_1,&local_388);
                                                          pcVar17 = local_378;
                                                          pcVar18 = local_370;
                                                          plVar12 = local_360;
                                                          pcVar19 = local_380;
                                                        }
                                                        else {
                                                          local_238 = plVar9;
                                                          plVar8 = (longlong *)
                                                                   FUN_18037f0c0(param_1,local_330,
                                                                                 DAT_1803ae7a0);
                                                          if (plVar8 == (longlong *)0x0) {
                                                            FUN_18000ce30(param_1,&local_388);
                                                            uVar16 = 0x84;
                                                            pcVar17 = local_378;
                                                            pcVar18 = local_370;
                                                            plVar12 = local_360;
                                                            pcVar19 = local_380;
                                                          }
                                                          else {
                                                            plVar10 = (longlong *)
                                                                      FUN_18037f0c0(param_1,plVar8,
                                                                                    DAT_1803ae958);
                                                            if ((-1 < (int)*plVar8) &&
                                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0
                                                               )) {
                                                              lVar15 = FUN_18000cc40(plVar8);
                                                              (**(code **)(lVar15 + 0x30))(plVar8);
                                                            }
                                                            uVar2 = DAT_1803ae968;
                                                            uVar16 = DAT_1803ae960;
                                                            if (plVar10 == (longlong *)0x0) {
                                                              FUN_18000ce30(param_1,&local_388);
                                                              uVar16 = 0x84;
                                                              pcVar17 = local_378;
                                                              pcVar18 = local_370;
                                                              plVar12 = local_360;
                                                              pcVar19 = local_380;
                                                            }
                                                            else {
                                                              uVar14 = _PyDict_NewPresized(2);
                                                              PyDict_SetItem(uVar14,uVar16,uVar2);
                                                              PyDict_SetItem(uVar14,DAT_1803ae970,
                                                                             local_320);
                                                              pcVar18 = (code *)FUN_180372160(
                                                  param_1,2);
                                                  FUN_180024750(pcVar18,0,uVar14);
                                                  uVar2 = DAT_1803ae978;
                                                  uVar16 = DAT_1803ae960;
                                                  uVar14 = _PyDict_NewPresized(2);
                                                  PyDict_SetItem(uVar14,uVar16,uVar2);
                                                  PyDict_SetItem(uVar14,DAT_1803ae970,plVar9);
                                                  FUN_180024750(pcVar18,1,uVar14);
                                                  plVar12 = local_360;
                                                  local_348 = DAT_1803ae988;
                                                  if (local_218 != _Py_NoneStruct_exref) {
                                                    local_348 = local_218;
                                                  }
                                                  local_350 = DAT_1803ae980;
                                                  uVar16 = 0x84;
                                                  *(undefined4 *)(local_360 + 5) = 0x84;
                                                  local_4b8 = (longlong *)DAT_1803ae990;
                                                  local_368 = pcVar18;
                                                  pcVar19 = (code *)FUN_18037d030(param_1,plVar10,
                                                                                  &local_368,
                                                                                  &local_350);
                                                  if ((-1 < (int)*plVar10) &&
                                                     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                    lVar15 = FUN_18000cc40(plVar10);
                                                    (**(code **)(lVar15 + 0x30))(plVar10);
                                                  }
                                                  if ((-1 < *(int *)pcVar18) &&
                                                     (*(longlong *)pcVar18 =
                                                           *(longlong *)pcVar18 + -1,
                                                     *(longlong *)pcVar18 == 0)) {
                                                    lVar15 = FUN_18000cc40(pcVar18);
                                                    (**(code **)(lVar15 + 0x30))(pcVar18);
                                                  }
                                                  plVar10 = local_1b0;
                                                  plVar8 = local_2a0;
                                                  if (pcVar19 != (code *)0x0) {
                                                    uVar16 = 0x8c;
                                                    *(undefined4 *)(plVar12 + 5) = 0x8c;
                                                    local_120 = local_2b0;
                                                    local_118 = local_1b0;
                                                    local_110 = local_2a0;
                                                    local_230 = pcVar19;
                                                    local_128 = pcVar19;
                                                    lVar15 = FUN_18037eae0(param_1,local_330,
                                                                           DAT_1803ae998,&local_128)
                                                    ;
                                                    if (lVar15 != 0) {
                                                      FUN_180001d80(param_1);
                                                      if ((-1 < (int)*plVar10) &&
                                                         (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                                                      {
                                                        lVar11 = FUN_18000cc40(plVar10);
                                                        (**(code **)(lVar11 + 0x30))(plVar10);
                                                      }
                                                      if ((-1 < (int)*local_1c0) &&
                                                         (*local_1c0 = *local_1c0 + -1,
                                                         *local_1c0 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_250);
                                                        (**(code **)(lVar11 + 0x30))(local_250);
                                                      }
                                                      pcVar18 = local_378;
                                                      if ((-1 < *(int *)local_378) &&
                                                         (*(longlong *)local_378 =
                                                               *(longlong *)local_378 + -1,
                                                         *(longlong *)local_378 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_378);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      pcVar18 = local_380;
                                                      if ((-1 < *(int *)local_380) &&
                                                         (*(longlong *)local_380 =
                                                               *(longlong *)local_380 + -1,
                                                         *(longlong *)local_380 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_380);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      if ((-1 < (int)*plVar8) &&
                                                         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                        lVar11 = FUN_18000cc40(plVar8);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_2b8;
                                                      if ((-1 < (int)*local_2b8) &&
                                                         (*local_2b8 = *local_2b8 + -1,
                                                         *local_2b8 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_2b8);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_320;
                                                      if ((-1 < (int)*local_320) &&
                                                         (*local_320 = *local_320 + -1,
                                                         *local_320 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_320);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      pcVar18 = local_370;
                                                      if (((local_370 != (code *)0x0) &&
                                                          (-1 < *(int *)local_370)) &&
                                                         (*(longlong *)local_370 =
                                                               *(longlong *)local_370 + -1,
                                                         *(longlong *)local_370 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_370);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      plVar8 = local_200;
                                                      if (((local_200 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_200)) &&
                                                         (*local_200 = *local_200 + -1,
                                                         *local_200 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_200);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      pcVar18 = local_318;
                                                      if ((-1 < *(int *)local_318) &&
                                                         (*(longlong *)local_318 =
                                                               *(longlong *)local_318 + -1,
                                                         *(longlong *)local_318 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_318);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      pcVar18 = local_310;
                                                      if ((-1 < *(int *)local_310) &&
                                                         (*(longlong *)local_310 =
                                                               *(longlong *)local_310 + -1,
                                                         *(longlong *)local_310 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_310);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      pcVar18 = local_308;
                                                      if ((-1 < *(int *)local_308) &&
                                                         (*(longlong *)local_308 =
                                                               *(longlong *)local_308 + -1,
                                                         *(longlong *)local_308 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_308);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      pcVar18 = local_300;
                                                      if ((-1 < *(int *)local_300) &&
                                                         (*(longlong *)local_300 =
                                                               *(longlong *)local_300 + -1,
                                                         *(longlong *)local_300 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_300);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      pcVar18 = local_2f8;
                                                      if ((-1 < *(int *)local_2f8) &&
                                                         (*(longlong *)local_2f8 =
                                                               *(longlong *)local_2f8 + -1,
                                                         *(longlong *)local_2f8 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_2f8);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      pcVar18 = local_2f0;
                                                      if ((-1 < *(int *)local_2f0) &&
                                                         (*(longlong *)local_2f0 =
                                                               *(longlong *)local_2f0 + -1,
                                                         *(longlong *)local_2f0 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_2f0);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      pcVar18 = local_2e8;
                                                      if ((-1 < *(int *)local_2e8) &&
                                                         (*(longlong *)local_2e8 =
                                                               *(longlong *)local_2e8 + -1,
                                                         *(longlong *)local_2e8 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_2e8);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      pcVar18 = local_2e0;
                                                      if ((-1 < *(int *)local_2e0) &&
                                                         (*(longlong *)local_2e0 =
                                                               *(longlong *)local_2e0 + -1,
                                                         *(longlong *)local_2e0 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_2e0);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      pcVar18 = local_2d8;
                                                      if ((-1 < *(int *)local_2d8) &&
                                                         (*(longlong *)local_2d8 =
                                                               *(longlong *)local_2d8 + -1,
                                                         *(longlong *)local_2d8 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_2d8);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      pcVar18 = local_2d0;
                                                      if ((-1 < *(int *)local_2d0) &&
                                                         (*(longlong *)local_2d0 =
                                                               *(longlong *)local_2d0 + -1,
                                                         *(longlong *)local_2d0 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_2d0);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      plVar8 = local_1f8;
                                                      if (((local_1f8 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_1f8)) &&
                                                         (*local_1f8 = *local_1f8 + -1,
                                                         *local_1f8 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_1f8);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_1f0;
                                                      if (((local_1f0 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_1f0)) &&
                                                         (*local_1f0 = *local_1f0 + -1,
                                                         *local_1f0 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_1f0);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_1e8;
                                                      if (((local_1e8 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_1e8)) &&
                                                         (*local_1e8 = *local_1e8 + -1,
                                                         *local_1e8 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_1e8);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_1e0;
                                                      if (((local_1e0 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_1e0)) &&
                                                         (*local_1e0 = *local_1e0 + -1,
                                                         *local_1e0 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_1e0);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_1d8;
                                                      if (((local_1d8 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_1d8)) &&
                                                         (*local_1d8 = *local_1d8 + -1,
                                                         *local_1d8 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_1d8);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_1d0;
                                                      if (((local_1d0 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_1d0)) &&
                                                         (*local_1d0 = *local_1d0 + -1,
                                                         *local_1d0 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_1d0);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_210;
                                                      if (((local_210 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_210)) &&
                                                         (*local_210 = *local_210 + -1,
                                                         *local_210 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_210);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_248;
                                                      if (((local_248 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_248)) &&
                                                         (*local_248 = *local_248 + -1,
                                                         *local_248 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_248);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_240;
                                                      if (((local_240 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_240)) &&
                                                         (*local_240 = *local_240 + -1,
                                                         *local_240 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_240);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      if ((-1 < (int)*plVar9) &&
                                                         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                        lVar11 = FUN_18000cc40(plVar9);
                                                        (**(code **)(lVar11 + 0x30))(plVar9);
                                                      }
                                                      if ((-1 < *(int *)pcVar19) &&
                                                         (*(longlong *)pcVar19 =
                                                               *(longlong *)pcVar19 + -1,
                                                         *(longlong *)pcVar19 == 0)) {
                                                        lVar11 = FUN_18000cc40(pcVar19);
                                                        (**(code **)(lVar11 + 0x30))(pcVar19);
                                                      }
                                                      plVar8 = local_330;
                                                      if ((-1 < (int)*local_330) &&
                                                         (*local_330 = *local_330 + -1,
                                                         *local_330 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_330);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      pcVar18 = local_2b0;
                                                      if ((-1 < *(int *)local_2b0) &&
                                                         (*(longlong *)local_2b0 =
                                                               *(longlong *)local_2b0 + -1,
                                                         *(longlong *)local_2b0 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_2b0);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      plVar8 = local_2a8;
                                                      if ((-1 < (int)*local_2a8) &&
                                                         (*local_2a8 = *local_2a8 + -1,
                                                         *local_2a8 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_2a8);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_1b8;
                                                      if ((-1 < (int)*local_1b8) &&
                                                         (*local_1b8 = *local_1b8 + -1,
                                                         *local_1b8 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_1b8);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_328;
                                                      if ((-1 < (int)*local_328) &&
                                                         (*local_328 = *local_328 + -1,
                                                         *local_328 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_328);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_358;
                                                      if ((-1 < (int)*local_358) &&
                                                         (*local_358 = *local_358 + -1,
                                                         *local_358 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_358);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      plVar8 = local_228;
                                                      if ((-1 < (int)*local_228) &&
                                                         (*local_228 = *local_228 + -1,
                                                         *local_228 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_228);
                                                        (**(code **)(lVar11 + 0x30))(plVar8);
                                                      }
                                                      pcVar18 = local_220;
                                                      if ((-1 < *(int *)local_220) &&
                                                         (*(longlong *)local_220 =
                                                               *(longlong *)local_220 + -1,
                                                         *(longlong *)local_220 == 0)) {
                                                        lVar11 = FUN_18000cc40(local_220);
                                                        (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      }
                                                      pcVar18 = local_218;
                                                      if (*(int *)local_218 < 0) {
                                                        return lVar15;
                                                      }
                                                      *(longlong *)local_218 =
                                                           *(longlong *)local_218 + -1;
                                                      if (*(longlong *)local_218 != 0) {
                                                        return lVar15;
                                                      }
                                                      lVar11 = FUN_18000cc40(local_218);
                                                      (**(code **)(lVar11 + 0x30))(pcVar18);
                                                      return lVar15;
                                                    }
                                                    goto LAB_18004cfbb;
                                                  }
                                                  FUN_18000ce30(param_1,&local_388);
                                                  pcVar17 = local_378;
                                                  pcVar18 = local_370;
                                                  pcVar19 = local_380;
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  }
                                                  else {
                                                    plVar9 = (longlong *)
                                                             FUN_18037f0c0(param_1,plVar10,
                                                                           DAT_1803ae7b0);
                                                    if (plVar9 == (longlong *)0x0) {
                                                      FUN_18000ce30(param_1,&local_388);
                                                      uVar16 = 0x6f;
                                                      pcVar17 = local_378;
                                                      pcVar18 = local_370;
                                                      pcVar19 = local_380;
                                                    }
                                                    else {
                                                      uVar16 = 0x6f;
                                                      *(undefined4 *)(plVar12 + 5) = 0x6f;
                                                      plVar8 = (longlong *)
                                                               FUN_180376210(param_1,plVar9);
                                                      if ((-1 < (int)*plVar9) &&
                                                         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                        lVar15 = FUN_18000cc40(plVar9);
                                                        (**(code **)(lVar15 + 0x30))(plVar9);
                                                      }
                                                      if (plVar8 != (longlong *)0x0) {
                                                        iVar5 = FUN_180001510(plVar8);
                                                        if (iVar5 == -1) {
                                                          FUN_18000ce30(param_1,&local_388);
                                                          goto LAB_18004ba57;
                                                        }
                                                        bVar20 = iVar5 != 0;
                                                        if ((-1 < (int)*plVar8) &&
                                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                          lVar15 = FUN_18000cc40(plVar8);
                                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                                        }
                                                        goto LAB_18004cb42;
                                                      }
                                                      FUN_18000ce30(param_1,&local_388);
                                                      pcVar17 = local_378;
                                                      pcVar18 = local_370;
                                                      pcVar19 = local_380;
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
                                                  }
                                                  }
                                                  }
                                                  }
                                                }
                                                goto LAB_18004cfd4;
                                              }
                                              FUN_18000ce30(param_1,&local_388);
                                              if ((-1 < (int)*plVar12) &&
                                                 (*plVar12 = *plVar12 + -1, *plVar12 == 0)) {
                                                lVar15 = FUN_18000cc40(plVar12);
                                                (**(code **)(lVar15 + 0x30))(plVar12);
                                              }
                                            }
                                          }
                                          uVar16 = 0x57;
                                          pcVar17 = local_378;
                                          plVar12 = local_360;
                                          pcVar19 = local_380;
                                        }
                                        else {
                                          lVar11 = FUN_180049820(param_1);
                                          if (lVar11 == 0) {
                                            FUN_180373340(param_1,&local_388,DAT_1803ae850);
                                            goto LAB_18004acd1;
                                          }
                                          pcVar19 = (code *)FUN_18037f0c0(param_1,plVar12,
                                                                          DAT_1803ae858);
                                          if (pcVar19 == (code *)0x0) {
                                            FUN_18000ce30(param_1,&local_388);
                                            if ((-1 < (int)*plVar8) &&
                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                              lVar15 = FUN_18000cc40(plVar8);
                                              (**(code **)(lVar15 + 0x30))(plVar8);
                                            }
                                            uVar16 = 0x38;
                                            pcVar17 = local_378;
                                            pcVar18 = local_370;
                                            plVar12 = local_360;
                                            pcVar19 = local_380;
                                            goto LAB_18004cfd4;
                                          }
                                          uVar16 = 0x38;
                                          *(undefined4 *)(local_360 + 5) = 0x38;
                                          local_2c0 = local_378;
                                          local_2c8 = pcVar19;
                                          pcVar17 = (code *)FUN_180377750(param_1,lVar11,&local_2c8)
                                          ;
                                          if ((-1 < *(int *)pcVar19) &&
                                             (*(longlong *)pcVar19 = *(longlong *)pcVar19 + -1,
                                             *(longlong *)pcVar19 == 0)) {
                                            lVar11 = FUN_18000cc40(pcVar19);
                                            (**(code **)(lVar11 + 0x30))(pcVar19);
                                          }
                                          if (pcVar17 != (code *)0x0) goto LAB_18004ae03;
                                          FUN_18000ce30(param_1,&local_388);
LAB_18004adca:
                                          pcVar17 = local_378;
                                          pcVar18 = local_370;
                                          plVar12 = local_360;
                                          pcVar19 = local_380;
                                          if ((-1 < (int)*plVar8) &&
                                             (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                            lVar15 = FUN_18000cc40(plVar8);
                                            (**(code **)(lVar15 + 0x30))(plVar8);
                                            pcVar17 = local_378;
                                            pcVar18 = local_370;
                                            plVar12 = local_360;
                                            pcVar19 = local_380;
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
                          plVar10 = (longlong *)FUN_18037f0c0(param_1,local_2a8,DAT_1803ae7b0);
                          uVar16 = 0x3c;
                          if (plVar10 == (longlong *)0x0) {
                            local_388 = *(longlong *)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            pcVar17 = local_378;
                            pcVar18 = local_370;
                            pcVar19 = local_380;
                          }
                          else {
                            *(undefined4 *)(plVar12 + 5) = 0x3c;
                            plVar13 = (longlong *)FUN_180376210(param_1,plVar10);
                            if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                              (**(code **)(plVar10[1] + 0x30))(plVar10);
                            }
                            if (plVar13 == (longlong *)0x0) {
                              FUN_18000ce30(param_1,&local_388);
                              pcVar17 = local_378;
                              pcVar18 = local_370;
                              pcVar19 = local_380;
                            }
                            else {
                              iVar5 = FUN_180001510(plVar13);
                              if (iVar5 == -1) {
                                FUN_18000ce30(param_1,&local_388);
                                pcVar17 = local_378;
                                pcVar18 = local_370;
                                pcVar19 = local_380;
                                if ((-1 < (int)*plVar13) &&
                                   (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                  (**(code **)(plVar13[1] + 0x30))(plVar13);
                                  pcVar17 = local_378;
                                  pcVar18 = local_370;
                                  pcVar19 = local_380;
                                }
                              }
                              else {
                                if ((-1 < (int)*plVar13) &&
                                   (*plVar13 = *plVar13 + -1, *plVar13 == 0)) {
                                  (**(code **)(plVar13[1] + 0x30))(plVar13);
                                }
                                if (iVar5 == 0) {
                                  lVar15 = FUN_180049b50(param_1);
                                  pcVar17 = local_378;
                                  if (lVar15 == 0) {
                                    FUN_180373340(param_1,&local_388,DAT_1803ae820);
                                    pcVar17 = local_378;
                                    pcVar18 = local_370;
                                    pcVar19 = local_380;
                                  }
                                  else {
                                    cVar4 = thunk_FUN_180383620(local_378,DAT_1803ae800);
                                    pcVar19 = local_380;
                                    local_198 = local_380;
                                    local_190 = _Py_FalseStruct_exref;
                                    if (cVar4 != '\0') {
                                      local_190 = _Py_TrueStruct_exref;
                                    }
                                    *(undefined4 *)(plVar12 + 5) = 0x3c;
                                    plVar8 = (longlong *)FUN_180377750(param_1,lVar15,&local_198);
                                    if (plVar8 != (longlong *)0x0) goto LAB_18004af9b;
                                    FUN_18000ce30(param_1,&local_388);
                                    pcVar18 = local_370;
                                  }
                                }
                                else {
                                  iVar5 = (int)*plVar8 + 1;
                                  pcVar17 = local_378;
                                  pcVar19 = local_380;
                                  if (iVar5 != 0) {
                                    *(int *)plVar8 = iVar5;
                                  }
LAB_18004af9b:
                                  plVar10 = local_328;
                                  pcVar18 = DAT_1803ae848;
                                  if (*(int *)DAT_1803ae848 + 1 != 0) {
                                    *(int *)DAT_1803ae848 = *(int *)DAT_1803ae848 + 1;
                                  }
                                  local_370 = pcVar18;
                                  local_2b8 = plVar8;
                                  iVar5 = FUN_180001510(local_328);
                                  if (iVar5 == -1) {
                                    FUN_18000ce30(param_1,&local_388);
                                    uVar16 = 0x3e;
                                  }
                                  else if (iVar5 == 0) {
LAB_18004b187:
                                    lVar15 = FUN_180049710(param_1);
                                    if (lVar15 == 0) {
                                      FUN_180373340(param_1,&local_388,DAT_1803ae870);
                                      uVar16 = 0x43;
                                      pcVar17 = local_378;
                                      pcVar18 = local_370;
                                      pcVar19 = local_380;
                                    }
                                    else {
                                      lVar11 = FUN_180049930(param_1);
                                      pcVar17 = local_378;
                                      pcVar18 = _Py_NoneStruct_exref;
                                      if (lVar11 == 0) {
                                        FUN_180373340(param_1,&local_388,DAT_1803ae830);
                                        uVar16 = 0x46;
                                        pcVar17 = local_378;
                                        pcVar18 = local_370;
                                        pcVar19 = local_380;
                                      }
                                      else {
                                        uVar16 = 0x46;
                                        *(undefined4 *)(plVar12 + 5) = 0x46;
                                        local_1a8 = pcVar18;
                                        local_1a0 = local_378;
                                        plVar8 = (longlong *)
                                                 FUN_180377750(param_1,lVar11,&local_1a8);
                                        if (plVar8 == (longlong *)0x0) {
                                          FUN_18000ce30(param_1,&local_388);
                                          pcVar18 = local_370;
                                          pcVar19 = local_380;
                                        }
                                        else {
                                          plVar10 = (longlong *)
                                                    FUN_180001b80(param_1,plVar9,DAT_1803ae838);
                                          if (plVar10 == (longlong *)0x0) {
                                            FUN_18000ce30(param_1,&local_388);
                                            if ((-1 < (int)*plVar8) &&
                                               (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                              lVar15 = FUN_18000cc40(plVar8);
                                              (**(code **)(lVar15 + 0x30))(plVar8);
                                            }
                                            uVar16 = 0x47;
                                            pcVar18 = local_370;
                                            pcVar19 = local_380;
                                          }
                                          else {
                                            iVar5 = FUN_180001510(plVar10);
                                            if (iVar5 == -1) {
                                              FUN_18000ce30(param_1,&local_388);
                                              if ((-1 < (int)*plVar8) &&
                                                 (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                lVar15 = FUN_18000cc40(plVar8);
                                                (**(code **)(lVar15 + 0x30))(plVar8);
                                              }
                                              if ((-1 < (int)*plVar10) &&
                                                 (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                lVar15 = FUN_18000cc40(plVar10);
                                                (**(code **)(lVar15 + 0x30))(plVar10);
                                              }
                                              uVar16 = 0x47;
                                              pcVar17 = local_378;
                                              pcVar18 = local_370;
                                              pcVar19 = local_380;
                                            }
                                            else {
                                              if ((-1 < (int)*plVar10) &&
                                                 (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                                lVar11 = FUN_18000cc40(plVar10);
                                                (**(code **)(lVar11 + 0x30))(plVar10);
                                              }
                                              pcVar18 = local_370;
                                              local_100 = DAT_1803ae848;
                                              if (iVar5 != 0) {
                                                local_100 = DAT_1803ae840;
                                              }
                                              uVar16 = 0x43;
                                              local_f0 = local_378;
                                              *(undefined4 *)(plVar12 + 5) = 0x43;
                                              local_170 = local_1b0;
                                              local_4b8 = (longlong *)DAT_1803ae878;
                                              local_f8 = local_370;
                                              local_178 = plVar9;
                                              local_108 = plVar8;
                                              plVar9 = (longlong *)
                                                       FUN_18037d650(param_1,lVar15,&local_178,
                                                                     &local_108);
                                              if ((-1 < (int)*plVar8) &&
                                                 (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                                lVar15 = FUN_18000cc40(plVar8);
                                                (**(code **)(lVar15 + 0x30))(plVar8);
                                              }
                                              if (plVar9 != (longlong *)0x0) goto LAB_18004b39f;
                                              FUN_18000ce30(param_1,&local_388);
                                              pcVar17 = local_378;
                                              pcVar19 = local_380;
                                            }
                                          }
                                        }
                                      }
                                    }
                                  }
                                  else {
                                    lVar15 = FUN_180049820(param_1);
                                    if (lVar15 == 0) {
                                      FUN_180373340(param_1,&local_388,DAT_1803ae850);
                                      uVar16 = 0x3f;
                                      pcVar19 = local_380;
                                    }
                                    else {
                                      plVar8 = (longlong *)
                                               FUN_18037f0c0(param_1,plVar10,DAT_1803ae858);
                                      if (plVar8 == (longlong *)0x0) {
                                        FUN_18000ce30(param_1,&local_388);
                                        uVar16 = 0x3f;
                                        pcVar17 = local_378;
                                        pcVar19 = local_380;
                                      }
                                      else {
                                        uVar16 = 0x3f;
                                        local_180 = local_378;
                                        *(undefined4 *)(plVar12 + 5) = 0x3f;
                                        local_188 = plVar8;
                                        plVar10 = (longlong *)
                                                  FUN_180377750(param_1,lVar15,&local_188);
                                        if ((-1 < (int)*plVar8) &&
                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                          lVar15 = FUN_18000cc40(plVar8);
                                          (**(code **)(lVar15 + 0x30))(plVar8);
                                        }
                                        if (plVar10 != (longlong *)0x0) {
                                          local_370 = DAT_1803ae868;
                                          if (*(int *)DAT_1803ae868 + 1 != 0) {
                                            *(int *)DAT_1803ae868 = *(int *)DAT_1803ae868 + 1;
                                          }
                                          local_200 = plVar10;
                                          if ((-1 < *(int *)pcVar18) &&
                                             (*(longlong *)pcVar18 = *(longlong *)pcVar18 + -1,
                                             *(longlong *)pcVar18 == 0)) {
                                            lVar15 = FUN_18000cc40(pcVar18);
                                            (**(code **)(lVar15 + 0x30))(pcVar18);
                                          }
                                          goto LAB_18004b187;
                                        }
                                        FUN_18000ce30(param_1,&local_388);
                                        pcVar17 = local_378;
                                        pcVar19 = local_380;
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
                    pcVar18 = (code *)FUN_18037f0c0(param_1,plVar10,DAT_1803ae7f0);
                    if (pcVar18 == (code *)0x0) {
                      local_388 = *(longlong *)(param_1 + 0x60);
                      uVar16 = 0x2d;
                      pcVar19 = (code *)0x0;
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      pcVar18 = (code *)0x0;
                    }
                    else {
                      iVar5 = FUN_180001510(pcVar18);
                      if (iVar5 == -1) {
                        local_388 = *(longlong *)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < *(int *)pcVar18) &&
                           (*(longlong *)pcVar18 = *(longlong *)pcVar18 + -1,
                           *(longlong *)pcVar18 == 0)) {
                          (**(code **)(*(longlong *)(pcVar18 + 8) + 0x30))(pcVar18);
                        }
                        uVar16 = 0x2d;
                        pcVar19 = (code *)0x0;
                        pcVar18 = (code *)0x0;
                      }
                      else {
                        if (iVar5 != 1) {
                          if ((-1 < *(int *)pcVar18) &&
                             (*(longlong *)pcVar18 = *(longlong *)pcVar18 + -1,
                             *(longlong *)pcVar18 == 0)) {
                            (**(code **)(*(longlong *)(pcVar18 + 8) + 0x30))(pcVar18);
                          }
                          pcVar18 = DAT_1803ae7f8;
                          if (*(int *)DAT_1803ae7f8 + 1 != 0) {
                            *(int *)DAT_1803ae7f8 = *(int *)DAT_1803ae7f8 + 1;
                          }
                        }
                        uVar6 = FUN_1803836d0(pcVar18,DAT_1803ae800);
                        if ((-1 < *(int *)pcVar18) &&
                           (*(longlong *)pcVar18 = *(longlong *)pcVar18 + -1,
                           *(longlong *)pcVar18 == 0)) {
                          (**(code **)(*(longlong *)(pcVar18 + 8) + 0x30))(pcVar18);
                        }
                        if (uVar6 != 0xffffffff) goto LAB_18004a90a;
                        local_388 = *(longlong *)(param_1 + 0x60);
                        uVar16 = 0x2d;
                        pcVar19 = (code *)0x0;
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        pcVar18 = (code *)0x0;
                      }
                    }
                  }
                }
              }
              goto LAB_18004cfd4;
            }
          }
LAB_18004a605:
          local_388 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          pcVar17 = (code *)0x0;
          pcVar19 = (code *)0x0;
        }
      }
    }
  }
LAB_18004cfd4:
  piVar1 = *(int **)(local_388 + 0x28);
  if (piVar1 == (int *)0x0) {
    lVar15 = FUN_18038e2a0(plVar12,uVar16);
  }
  else {
    if (*(longlong **)(piVar1 + 6) == plVar12) goto LAB_18004d016;
    lVar15 = FUN_18038e2a0(plVar12,uVar16);
    *(int **)(lVar15 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
  }
  FUN_180001300(&local_388,lVar15);
LAB_18004d016:
  plVar9 = local_2b8;
  local_398 = local_230;
  local_3a0 = local_238;
  local_3a8 = local_240;
  local_3b0 = local_248;
  local_3b8 = local_210;
  local_3c0 = local_1d0;
  local_3c8 = local_1d8;
  local_3d0 = local_1e0;
  local_3d8 = local_1e8;
  local_3e0 = local_1f0;
  local_3e8 = local_1f8;
  local_3f0 = local_2d0;
  local_3f8 = local_2d8;
  local_400 = local_2e0;
  local_408 = local_2e8;
  local_410 = local_2f0;
  local_418 = local_2f8;
  local_420 = local_300;
  local_428 = local_308;
  local_430 = local_310;
  local_438 = local_318;
  local_440 = local_200;
  local_450 = local_320;
  local_458 = local_2b8;
  local_460 = local_1c8;
  local_478 = local_250;
  local_480 = local_208;
  local_488 = local_218;
  local_490 = local_220;
  local_498 = local_228;
  local_4a0 = local_358;
  local_4a8 = local_328;
  local_4b0 = local_1b8;
  local_4b8 = local_2a8;
  local_470 = pcVar17;
  local_468 = pcVar19;
  local_448 = pcVar18;
  FUN_18036d0a0(plVar12,"ooooooooooooooooooooooooooooooooooooooo",local_330,local_2b0);
  plVar8 = DAT_1803ae720;
  if (plVar12 == DAT_1803ae720) {
    if ((-1 < (int)*DAT_1803ae720) && (*DAT_1803ae720 = *DAT_1803ae720 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))();
    }
    DAT_1803ae720 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  lVar15 = local_388;
  local_388 = 0;
  if (((local_208 != (longlong *)0x0) && (-1 < (int)*local_208)) &&
     (*local_208 = *local_208 + -1, *local_208 == 0)) {
    (**(code **)(local_208[1] + 0x30))();
  }
  if (((local_250 != (longlong *)0x0) && (-1 < (int)*local_250)) &&
     (*local_250 = *local_250 + -1, *local_250 == 0)) {
    (**(code **)(local_250[1] + 0x30))();
  }
  if (((pcVar17 != (code *)0x0) && (-1 < *(int *)pcVar17)) &&
     (*(longlong *)pcVar17 = *(longlong *)pcVar17 + -1, *(longlong *)pcVar17 == 0)) {
    (**(code **)(*(longlong *)(pcVar17 + 8) + 0x30))(pcVar17);
  }
  if (((pcVar19 != (code *)0x0) && (-1 < *(int *)pcVar19)) &&
     (*(longlong *)pcVar19 = *(longlong *)pcVar19 + -1, *(longlong *)pcVar19 == 0)) {
    (**(code **)(*(longlong *)(pcVar19 + 8) + 0x30))(pcVar19);
  }
  if (((local_1c8 != (longlong *)0x0) && (-1 < (int)*local_1c8)) &&
     (*local_1c8 = *local_1c8 + -1, *local_1c8 == 0)) {
    (**(code **)(local_1c8[1] + 0x30))();
  }
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  if (((local_320 != (longlong *)0x0) && (-1 < (int)*local_320)) &&
     (*local_320 = *local_320 + -1, *local_320 == 0)) {
    (**(code **)(local_320[1] + 0x30))();
  }
  if (((pcVar18 != (code *)0x0) && (-1 < *(int *)pcVar18)) &&
     (*(longlong *)pcVar18 = *(longlong *)pcVar18 + -1, *(longlong *)pcVar18 == 0)) {
    (**(code **)(*(longlong *)(pcVar18 + 8) + 0x30))(pcVar18);
  }
  if (((local_200 != (longlong *)0x0) && (-1 < (int)*local_200)) &&
     (*local_200 = *local_200 + -1, *local_200 == 0)) {
    (**(code **)(local_200[1] + 0x30))();
  }
  if (((local_318 != (code *)0x0) && (-1 < *(int *)local_318)) &&
     (*(longlong *)local_318 = *(longlong *)local_318 + -1, *(longlong *)local_318 == 0)) {
    (**(code **)(*(longlong *)(local_318 + 8) + 0x30))();
  }
  if (((local_310 != (code *)0x0) && (-1 < *(int *)local_310)) &&
     (*(longlong *)local_310 = *(longlong *)local_310 + -1, *(longlong *)local_310 == 0)) {
    (**(code **)(*(longlong *)(local_310 + 8) + 0x30))();
  }
  if (((local_308 != (code *)0x0) && (-1 < *(int *)local_308)) &&
     (*(longlong *)local_308 = *(longlong *)local_308 + -1, *(longlong *)local_308 == 0)) {
    (**(code **)(*(longlong *)(local_308 + 8) + 0x30))();
  }
  if (((local_300 != (code *)0x0) && (-1 < *(int *)local_300)) &&
     (*(longlong *)local_300 = *(longlong *)local_300 + -1, *(longlong *)local_300 == 0)) {
    (**(code **)(*(longlong *)(local_300 + 8) + 0x30))();
  }
  if (((local_2f8 != (code *)0x0) && (-1 < *(int *)local_2f8)) &&
     (*(longlong *)local_2f8 = *(longlong *)local_2f8 + -1, *(longlong *)local_2f8 == 0)) {
    (**(code **)(*(longlong *)(local_2f8 + 8) + 0x30))();
  }
  if (((local_2f0 != (code *)0x0) && (-1 < *(int *)local_2f0)) &&
     (*(longlong *)local_2f0 = *(longlong *)local_2f0 + -1, *(longlong *)local_2f0 == 0)) {
    (**(code **)(*(longlong *)(local_2f0 + 8) + 0x30))();
  }
  if (((local_2e8 != (code *)0x0) && (-1 < *(int *)local_2e8)) &&
     (*(longlong *)local_2e8 = *(longlong *)local_2e8 + -1, *(longlong *)local_2e8 == 0)) {
    (**(code **)(*(longlong *)(local_2e8 + 8) + 0x30))();
  }
  if (((local_2e0 != (code *)0x0) && (-1 < *(int *)local_2e0)) &&
     (*(longlong *)local_2e0 = *(longlong *)local_2e0 + -1, *(longlong *)local_2e0 == 0)) {
    (**(code **)(*(longlong *)(local_2e0 + 8) + 0x30))();
  }
  if (((local_2d8 != (code *)0x0) && (-1 < *(int *)local_2d8)) &&
     (*(longlong *)local_2d8 = *(longlong *)local_2d8 + -1, *(longlong *)local_2d8 == 0)) {
    (**(code **)(*(longlong *)(local_2d8 + 8) + 0x30))();
  }
  if (((local_2d0 != (code *)0x0) && (-1 < *(int *)local_2d0)) &&
     (*(longlong *)local_2d0 = *(longlong *)local_2d0 + -1, *(longlong *)local_2d0 == 0)) {
    (**(code **)(*(longlong *)(local_2d0 + 8) + 0x30))();
  }
  if (((local_1f8 != (longlong *)0x0) && (-1 < (int)*local_1f8)) &&
     (*local_1f8 = *local_1f8 + -1, *local_1f8 == 0)) {
    (**(code **)(local_1f8[1] + 0x30))();
  }
  if (((local_1f0 != (longlong *)0x0) && (-1 < (int)*local_1f0)) &&
     (*local_1f0 = *local_1f0 + -1, *local_1f0 == 0)) {
    (**(code **)(local_1f0[1] + 0x30))();
  }
  if (((local_1e8 != (longlong *)0x0) && (-1 < (int)*local_1e8)) &&
     (*local_1e8 = *local_1e8 + -1, *local_1e8 == 0)) {
    (**(code **)(local_1e8[1] + 0x30))();
  }
  if (((local_1e0 != (longlong *)0x0) && (-1 < (int)*local_1e0)) &&
     (*local_1e0 = *local_1e0 + -1, *local_1e0 == 0)) {
    (**(code **)(local_1e0[1] + 0x30))();
  }
  if (((local_1d8 != (longlong *)0x0) && (-1 < (int)*local_1d8)) &&
     (*local_1d8 = *local_1d8 + -1, *local_1d8 == 0)) {
    (**(code **)(local_1d8[1] + 0x30))();
  }
  if (((local_1d0 != (longlong *)0x0) && (-1 < (int)*local_1d0)) &&
     (*local_1d0 = *local_1d0 + -1, *local_1d0 == 0)) {
    (**(code **)(local_1d0[1] + 0x30))();
  }
  if (((local_210 != (longlong *)0x0) && (-1 < (int)*local_210)) &&
     (*local_210 = *local_210 + -1, *local_210 == 0)) {
    (**(code **)(local_210[1] + 0x30))();
  }
  if (((local_248 != (longlong *)0x0) && (-1 < (int)*local_248)) &&
     (*local_248 = *local_248 + -1, *local_248 == 0)) {
    (**(code **)(local_248[1] + 0x30))();
  }
  if (((local_240 != (longlong *)0x0) && (-1 < (int)*local_240)) &&
     (*local_240 = *local_240 + -1, *local_240 == 0)) {
    (**(code **)(local_240[1] + 0x30))();
  }
  if (((local_238 != (longlong *)0x0) && (-1 < (int)*local_238)) &&
     (*local_238 = *local_238 + -1, *local_238 == 0)) {
    (**(code **)(local_238[1] + 0x30))();
  }
  if (((local_230 != (code *)0x0) && (-1 < *(int *)local_230)) &&
     (*(longlong *)local_230 = *(longlong *)local_230 + -1, *(longlong *)local_230 == 0)) {
    (**(code **)(*(longlong *)(local_230 + 8) + 0x30))();
  }
  local_388 = lVar15;
  if ((-1 < (int)*local_330) && (*local_330 = *local_330 + -1, *local_330 == 0)) {
    (**(code **)(local_330[1] + 0x30))();
  }
  if ((-1 < *(int *)local_2b0) &&
     (*(longlong *)local_2b0 = *(longlong *)local_2b0 + -1, *(longlong *)local_2b0 == 0)) {
    (**(code **)(*(longlong *)(local_2b0 + 8) + 0x30))();
  }
  if ((-1 < (int)*local_2a8) && (*local_2a8 = *local_2a8 + -1, *local_2a8 == 0)) {
    (**(code **)(local_2a8[1] + 0x30))();
  }
  if ((-1 < (int)*local_1b8) && (*local_1b8 = *local_1b8 + -1, *local_1b8 == 0)) {
    (**(code **)(local_1b8[1] + 0x30))();
  }
  if ((-1 < (int)*local_328) && (*local_328 = *local_328 + -1, *local_328 == 0)) {
    (**(code **)(local_328[1] + 0x30))();
  }
  if ((-1 < (int)*local_358) && (*local_358 = *local_358 + -1, *local_358 == 0)) {
    (**(code **)(local_358[1] + 0x30))();
  }
  if ((-1 < (int)*local_228) && (*local_228 = *local_228 + -1, *local_228 == 0)) {
    (**(code **)(local_228[1] + 0x30))();
  }
  if ((-1 < *(int *)local_220) &&
     (*(longlong *)local_220 = *(longlong *)local_220 + -1, *(longlong *)local_220 == 0)) {
    (**(code **)(*(longlong *)(local_220 + 8) + 0x30))();
  }
  if ((-1 < *(int *)local_218) &&
     (*(longlong *)local_218 = *(longlong *)local_218 + -1, *(longlong *)local_218 == 0)) {
    (**(code **)(*(longlong *)(local_218 + 8) + 0x30))();
  }
  FUN_18000ce40(param_1,&local_388);
  return 0;
}

